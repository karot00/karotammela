package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"goth/internal/content"
	"goth/internal/i18n"
)

func (h *Handlers) BlogList(w http.ResponseWriter, r *http.Request) {
	locale := h.resolveLocale(r)
	posts, err := content.GetAllBlogPosts(locale)
	if err != nil {
		http.Error(w, "blog error", http.StatusInternalServerError)
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}
	result := content.PaginateBlogPosts(posts, page, 10)

	data := h.common(r, locale)
	tr := func(key string) string { return i18n.T(locale, key) }
	data["Title"] = tr("home.phaseLabel") + " — Blog"
	data["Subtitle"] = tr("home.intro")
	data["HasPosts"] = len(result.Items) > 0
	data["Posts"] = result.Items
	data["Page"] = result.Page
	data["Pages"] = result.Pages
	data["HasPrev"] = result.HasPrev
	data["HasNext"] = result.HasNext
	data["PrevPage"] = result.Page - 1
	data["NextPage"] = result.Page + 1
	data["PrevLabel"] = tr("dashboard.blogPreviousLabel")
	data["NextLabel"] = tr("dashboard.blogNextLabel")
	data["PageLabel"] = tr("dashboard.blogPageLabel")
	data["EmptyLabel"] = tr("dashboard.blogNoPostsLabel")
	h.render(w, "blog_list", data)
}

func (h *Handlers) BlogPost(w http.ResponseWriter, r *http.Request) {
	locale := h.resolveLocale(r)
	slug := chi.URLParam(r, "slug")
	post, err := content.GetBlogPostBySlug(locale, slug)
	if err != nil {
		http.Error(w, "blog error", http.StatusInternalServerError)
		return
	}
	if post == nil {
		http.NotFound(w, r)
		return
	}

	data := h.common(r, locale)
	tr := func(key string) string { return i18n.T(locale, key) }
	data["Title"] = post.Title + " | " + siteName
	data["Description"] = post.Description
	data["OGType"] = "article"
	data["OGArticlePublished"] = articleISO(post.PublishedAt)
	data["OGArticleModified"] = articleISO(post.PublishedAt)
	data["OGArticleTags"] = post.Tags
	data["JSONLDArticle"] = articleJSONLD(h.cfg, post)
	data["BackLabel"] = tr("dashboard.blogBackLabel")
	data["PublishedAt"] = post.PublishedAt
	data["Draft"] = post.Draft
	data["Tags"] = post.Tags
	data["HTML"] = post.HTML
	h.render(w, "blog_post", data)
}
