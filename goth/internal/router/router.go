package router

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"

	"goth/assets"
	"goth/internal/handler"
	webstatic "goth/web"
)

// New builds the HTTP router with locale-prefixed routes mirroring the Next.js app.
func New(h *handler.Handlers) http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/fi", http.StatusSeeOther)
	})

	r.Get("/api/ping", h.Ping)
	r.Options("/api/ping", h.Ping)
	r.Get("/api/stats", h.Stats)
	r.Get("/api/sentinel/stream", h.SentinelStream)
	r.Post("/api/sentinel", h.SentinelCommit)
	r.Post("/api/contact", h.Contact)
	r.Get("/api/ai-pulse/stocks", h.AiPulseStocks)
	r.Post("/api/ai-pulse/refresh", h.AiPulseRefresh)
	r.Get("/robots.txt", h.Robots)
	r.Get("/sitemap.xml", h.Sitemap)

	// VIP portal (MeetingPackage application). Disabled by default: every
	// route except /api/vip/status answers with an indistinguishable 404 and
	// no navigation shows a link. Go owns the canonical /en/vip; the Next.js
	// stack only links to it (plan §4).
	r.Get("/api/vip/status", h.VIPStatus)
	r.Get("/vip", h.VIPEntry)
	r.Get("/fi/vip", h.VIPEntry)
	r.Get("/en/vip", h.VIPPage)
	r.Post("/api/vip/notify", h.VIPNotify)
	r.Post("/api/vip/login", h.VIPLogin)
	r.Post("/api/vip/logout", h.VIPLogout)
	r.Get("/api/vip/cv", h.VIPCV)
	r.Post("/api/vip/chat", h.VIPChat)

	sub, err := fs.Sub(webstatic.StaticFS, "static")
	if err == nil {
		fileServer := http.FileServer(http.FS(sub))
		r.Handle("/static/*", http.StripPrefix("/static/", fileServer))
	}

	mediaSub, merr := fs.Sub(assets.MediaFS, "media")
	if merr == nil {
		mediaServer := http.FileServer(http.FS(mediaSub))
		r.Handle("/media/*", http.StripPrefix("/media/", mediaServer))
	}

	r.Route("/{locale}", func(r chi.Router) {
		r.Get("/", h.Home)
		r.Get("/blog", h.BlogList)
		r.Get("/blog/{slug}", h.BlogPost)
		r.Get("/privacy", h.Privacy)
		r.Get("/dashboard", h.Dashboard)
	})

	return r
}
