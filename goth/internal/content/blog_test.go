package content

import (
	"testing"
	"testing/fstest"
	"time"
)

func TestGetAllBlogPostsProducesSortedVisiblePosts(t *testing.T) {
	posts, err := GetAllBlogPosts("fi")
	if err != nil {
		t.Fatalf("GetAllBlogPosts fi: %v", err)
	}
	if len(posts) == 0 {
		t.Fatal("expected at least one fi post in embedded content")
	}
	// Sorted by PublishedAt descending.
	for i := 1; i < len(posts); i++ {
		if posts[i-1].PublishedAt < posts[i].PublishedAt {
			t.Errorf("posts not sorted desc: %q (%s) after %q (%s)",
				posts[i-1].Slug, posts[i-1].PublishedAt, posts[i].Slug, posts[i].PublishedAt)
		}
	}
	// Every returned post must have a non-empty title + valid slug.
	for _, p := range posts {
		if p.Title == "" {
			t.Errorf("post %q missing title", p.Slug)
		}
		if p.Slug == "" || !slugPattern.MatchString(p.Slug) {
			t.Errorf("post %q has invalid slug", p.Slug)
		}
	}
	// HTML must be sanitized (no raw script injection preserved).
	for _, p := range posts {
		if len(p.HTML) == 0 && len(p.Body) > 0 {
			t.Errorf("post %q has body but no sanitized HTML", p.Slug)
		}
	}
}

func TestGetBlogPostBySlug(t *testing.T) {
	posts, err := GetAllBlogPosts("en")
	if err != nil {
		t.Fatalf("GetAllBlogPosts en: %v", err)
	}
	if len(posts) == 0 {
		t.Skip("no en posts embedded")
	}
	first := posts[0]
	got, err := GetBlogPostBySlug("en", first.Slug)
	if err != nil {
		t.Fatalf("GetBlogPostBySlug: %v", err)
	}
	if got == nil {
		t.Fatalf("expected to find slug %q", first.Slug)
	}
	if got.Slug != first.Slug || got.Title != first.Title {
		t.Errorf("slug lookup mismatch: got %q/%q want %q/%q", got.Slug, got.Title, first.Slug, first.Title)
	}

	// Invalid slug characters must return nil, never a post.
	bad, err := GetBlogPostBySlug("en", "../etc/passwd")
	if err != nil {
		t.Fatalf("GetBlogPostBySlug bad: %v", err)
	}
	if bad != nil {
		t.Errorf("expected nil for invalid slug, got %+v", bad)
	}
}

// TestDraftsHiddenInEveryEnvironment verifies the fail-closed policy: drafts are
// excluded from the public list regardless of the GOTH_ENV value, so there is no
// unauthenticated draft-preview mode.
func TestDraftsHiddenInEveryEnvironment(t *testing.T) {
	for _, env := range []string{"", "development", "staging", "production", "anything"} {
		t.Setenv("GOTH_ENV", env)
		posts, err := GetAllBlogPosts("fi")
		if err != nil {
			t.Fatalf("GetAllBlogPosts (env=%q): %v", env, err)
		}
		for _, p := range posts {
			if p.Draft {
				t.Errorf("env=%q leaked draft post %q", env, p.Slug)
			}
		}
	}
}

// fixtureFS builds an in-memory blog tree with a published, a draft, and a
// future-dated post so filtering behavior can be asserted deterministically.
func fixtureFS() (fstest.MapFS, string) {
	future := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	fsys := fstest.MapFS{
		"blog/fi/2026-01-01-published.md": &fstest.MapFile{Data: []byte(
			"---\ntitle: Published\nslug: published\ndescription: d\npublishedAt: 2026-01-01\n---\nbody\n")},
		"blog/fi/2026-01-02-draft.md": &fstest.MapFile{Data: []byte(
			"---\ntitle: Draft\nslug: draft\ndescription: d\ndraft: true\npublishedAt: 2026-01-02\n---\nbody\n")},
		"blog/fi/" + future + "-future.md": &fstest.MapFile{Data: []byte(
			"---\ntitle: Future\nslug: future\ndescription: d\npublishedAt: " + future + "\n---\nbody\n")},
	}
	return fsys, future
}

func TestLoadPostsExcludesDraftsAndFutureDated(t *testing.T) {
	fsys, _ := fixtureFS()
	posts, err := loadPosts(fsys, "fi")
	if err != nil {
		t.Fatalf("loadPosts: %v", err)
	}
	slugs := map[string]bool{}
	for _, p := range posts {
		slugs[p.Slug] = true
	}
	if !slugs["published"] {
		t.Error("published post missing from results")
	}
	if slugs["draft"] {
		t.Error("draft post leaked despite fail-closed policy")
	}
	if slugs["future"] {
		t.Error("future-dated post leaked; must be excluded in every environment")
	}
}

func TestGetBlogPostBySlugHidesDraftAndFuture(t *testing.T) {
	fsys, _ := fixtureFS()
	// Direct slug access to a draft or future post must return nil (404), even
	// though the slug is well-formed. Only the published post resolves.
	draft, err := firstMatching(fsys, "fi", "draft")
	if err != nil {
		t.Fatalf("lookup draft: %v", err)
	}
	if draft != nil {
		t.Error("direct draft slug should return nil (404)")
	}
	fut, err := firstMatching(fsys, "fi", "future")
	if err != nil {
		t.Fatalf("lookup future: %v", err)
	}
	if fut != nil {
		t.Error("direct future-dated slug should return nil (404)")
	}
	pub, err := firstMatching(fsys, "fi", "published")
	if err != nil {
		t.Fatalf("lookup published: %v", err)
	}
	if pub == nil || pub.Slug != "published" {
		t.Error("published post should resolve by direct slug")
	}
}

// firstMatching is a test helper that scans loadPosts output for a slug.
func firstMatching(fsys fstest.MapFS, locale, slug string) (*BlogPost, error) {
	posts, err := loadPosts(fsys, locale)
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

func TestPaginateBlogPosts(t *testing.T) {
	makePosts := func(n int) []BlogPost {
		out := make([]BlogPost, n)
		for i := range out {
			out[i] = BlogPost{Slug: "p" + string(rune('a'+i))}
		}
		return out
	}

	// 25 posts, page size 10 -> 3 pages.
	res := PaginateBlogPosts(makePosts(25), 1, 10)
	if res.Total != 25 || res.Pages != 3 || !res.HasNext || res.HasPrev {
		t.Errorf("page1 meta wrong: %+v", res)
	}
	if len(res.Items) != 10 {
		t.Errorf("page1 items = %d, want 10", len(res.Items))
	}

	// Last page clamps and has fewer items.
	res = PaginateBlogPosts(makePosts(25), 99, 10)
	if res.Page != 3 || res.HasNext || !res.HasPrev || len(res.Items) != 5 {
		t.Errorf("last page wrong: %+v", res)
	}

	// Out-of-range low clamps to 1.
	res = PaginateBlogPosts(makePosts(25), 0, 10)
	if res.Page != 1 {
		t.Errorf("low page clamp = %d, want 1", res.Page)
	}

	// Default page size when <= 0.
	res = PaginateBlogPosts(makePosts(3), 1, 0)
	if res.PageSize != 10 || res.Pages != 1 {
		t.Errorf("default page size wrong: %+v", res)
	}
}

func TestSortPostsDescending(t *testing.T) {
	posts := []BlogPost{
		{Slug: "a", PublishedAt: "2026-01-01"},
		{Slug: "c", PublishedAt: "2026-03-01"},
		{Slug: "b", PublishedAt: "2026-02-01"},
	}
	sorted := sortPosts(posts)
	want := []string{"c", "b", "a"}
	for i, w := range want {
		if sorted[i].Slug != w {
			t.Errorf("sortPosts[%d] = %q, want %q", i, sorted[i].Slug, w)
		}
	}
}
