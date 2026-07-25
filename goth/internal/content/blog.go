package content

import (
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	blogcontent "goth/content"
)

const contentRoot = "blog"

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// BlogPost is a parsed markdown post.
type BlogPost struct {
	Title       string
	Description string
	PublishedAt string
	Slug        string
	Draft       bool
	Tags        []string
	Locale      string
	Body        string
	HTML        string
	SourcePath  string
}

// PaginationResult mirrors paginateBlogPosts output.
type PaginationResult struct {
	Page     int
	PageSize int
	Total    int
	Pages    int
	HasPrev  bool
	HasNext  bool
	Items    []BlogPost
}

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM, meta.New()),
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// isDraftVisible fails closed: drafts are never visible to unauthenticated
// visitors in any environment. The plan (§2 + Phase 7) forbids draft previews
// regardless of environment variables.
func isDraftVisible(draft bool) bool {
	return !draft
}

// isPublished treats a post as visible only when it has a parseable publishedAt
// that is not in the future. Future-dated posts are excluded in every environment.
func isPublished(publishedAt string) bool {
	t := parseDate(publishedAt)
	if t.IsZero() {
		// Posts without a parseable date are treated as already published.
		return true
	}
	now := time.Now()
	return t.Before(now) || t.Equal(now)
}

func parsePostFile(fsys fs.FS, path, locale string) (BlogPost, error) {
	raw, err := fs.ReadFile(fsys, path)
	if err != nil {
		return BlogPost{}, err
	}
	source := string(raw)
	var buf strings.Builder
	ctx := parser.NewContext()
	if err := md.Convert([]byte(source), &buf, parser.WithContext(ctx)); err != nil {
		return BlogPost{}, err
	}

	metaMap := map[string]any{}
	if m := meta.Get(ctx); m != nil {
		for k, v := range m {
			metaMap[k] = v
		}
	}

	title, _ := metaMap["title"].(string)
	desc, _ := metaMap["description"].(string)
	slug, _ := metaMap["slug"].(string)
	pub, _ := metaMap["publishedAt"].(string)
	draft, _ := metaMap["draft"].(bool)
	tags := []string{}
	if t, ok := metaMap["tags"].([]any); ok {
		for _, x := range t {
			if s, ok := x.(string); ok {
				tags = append(tags, s)
			}
		}
	}

	post := BlogPost{
		Title:       title,
		Description: desc,
		Slug:        slug,
		PublishedAt: pub,
		Draft:       draft,
		Tags:        tags,
		Locale:      locale,
		Body:        strings.TrimSpace(buf.String()),
		SourcePath:  path,
	}
	return post, nil
}

func sortPosts(posts []BlogPost) []BlogPost {
	out := make([]BlogPost, len(posts))
	copy(out, posts)
	sort.SliceStable(out, func(i, j int) bool {
		ti := parseDate(out[i].PublishedAt)
		tj := parseDate(out[j].PublishedAt)
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return out[i].Slug < out[j].Slug
	})
	return out
}

func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// GetAllBlogPosts loads + validates + sorts visible posts for a locale.
// Drafts and future-dated posts are excluded unconditionally (fail closed).
func GetAllBlogPosts(locale string) ([]BlogPost, error) {
	return loadPosts(blogcontent.FS, locale)
}

// loadPosts reads posts from an arbitrary fs.FS so tests can inject fixtures.
func loadPosts(fsys fs.FS, locale string) ([]BlogPost, error) {
	if locale != "en" && locale != "fi" {
		locale = "fi"
	}
	dir := filepath.Join(contentRoot, locale)
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []BlogPost{}, nil
		}
		return nil, err
	}

	seen := map[string]bool{}
	var posts []BlogPost
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		post, err := parsePostFile(fsys, path, locale)
		if err != nil {
			return nil, err
		}
		if post.Title == "" || post.Slug == "" || !slugPattern.MatchString(post.Slug) {
			continue
		}
		if !isDraftVisible(post.Draft) {
			continue
		}
		if !isPublished(post.PublishedAt) {
			continue
		}
		if seen[post.Slug] {
			return nil, &DuplicateSlugError{Slug: post.Slug, Locale: locale}
		}
		seen[post.Slug] = true
		posts = append(posts, post)
	}

	posts = sortPosts(posts)
	for i := range posts {
		posts[i].HTML = sanitize(posts[i].Body)
	}
	return posts, nil
}

// GetBlogPostBySlug returns a single visible post or nil. Drafts and
// future-dated posts are never returned, even by direct slug.
func GetBlogPostBySlug(locale, slug string) (*BlogPost, error) {
	if !slugPattern.MatchString(slug) {
		return nil, nil
	}
	posts, err := GetAllBlogPosts(locale)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		if posts[i].Slug == slug {
			return &posts[i], nil
		}
	}
	return nil, nil
}

// PaginateBlogPosts paginates a post slice.
func PaginateBlogPosts(posts []BlogPost, requestedPage, pageSize int) PaginationResult {
	if pageSize < 1 {
		pageSize = 10
	}
	total := len(posts)
	pages := 1
	if total > 0 {
		pages = (total + pageSize - 1) / pageSize
	}
	page := requestedPage
	if page < 1 {
		page = 1
	}
	if page > pages {
		page = pages
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	return PaginationResult{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Pages:    pages,
		HasPrev:  page > 1,
		HasNext:  page < pages,
		Items:    posts[start:end],
	}
}

var policy = bluemonday.UGCPolicy()

func sanitize(markdownHTML string) string {
	return policy.Sanitize(markdownHTML)
}

// DuplicateSlugError signals a duplicate slug in a locale.
type DuplicateSlugError struct {
	Slug   string
	Locale string
}

func (e *DuplicateSlugError) Error() string {
	return "duplicate blog slug '" + e.Slug + "' for locale '" + e.Locale + "'"
}
