package view

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"goth/internal/content"
)

//go:embed templates/*.html templates/**/*.html
var templateFS embed.FS

// Renderer holds parsed HTML templates.
type Renderer struct {
	tmpl *template.Template
}

// NewRenderer parses all embedded templates.
func NewRenderer() (*Renderer, error) {
	funcs := template.FuncMap{
		"json": func(s string) template.JS { return template.JS(s) },
		"safe": func(s string) template.HTML { return template.HTML(s) },
		"inc":  func(i int) int { return i + 1 },
		"dec":  func(i int) int { return i - 1 },
		"date": func(locale, d string) string { return content.FormatDate(locale, d) },
		"active": func(b bool) string {
			if b {
				return "bg-primary/20 text-primary font-semibold"
			}
			return "text-muted-foreground hover:text-foreground"
		},
		// sideActive mirrors the Next.js dashboard SidebarNav item classes.
		"sideActive": func(b bool) string {
			if b {
				return "bg-sidebar-primary text-sidebar-primary-foreground"
			}
			return "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
		},
		"icon": func(name, class string) template.HTML { return iconSVG(name, class) },
		// dict builds a map for calling parameterized sub-templates
		// (e.g. the consent category card) with a named data shape.
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict: expected even argument count, got %d", len(values))
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict: argument %d is not a string key", i)
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
		"changeClass": func(t string) string {
			switch t {
			case "added":
				return "border-emerald-500/20 bg-emerald-500/10 text-emerald-400"
			case "changed":
				return "border-sky-500/20 bg-sky-500/10 text-sky-400"
			case "fixed":
				return "border-amber-500/20 bg-amber-500/10 text-amber-400"
			case "removed":
				return "border-rose-500/20 bg-rose-500/10 text-rose-400"
			default:
				return "border-border bg-muted text-muted-foreground"
			}
		},
	}
	tmpl, err := template.New("goth").Funcs(funcs).ParseFS(templateFS, "templates/*.html", "templates/**/*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: tmpl}, nil
}

// Render executes a named template (e.g. "home") into w.
func (r *Renderer) Render(w io.Writer, name string, data any) error {
	return r.tmpl.ExecuteTemplate(w, name, data)
}

// strokeIcons are 24x24 line icons (lucide-style) rendered with
// stroke="currentColor". Keys cover the dashboard tech categories plus the
// mobile menu toggle.
var strokeIcons = map[string]string{
	"menu":        `<line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="18" y2="18"/>`,
	"nextjs":      `<rect width="18" height="18" x="3" y="3" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/>`,
	"tailwindCss": `<circle cx="13.5" cy="6.5" r=".5" fill="currentColor"/><circle cx="17.5" cy="10.5" r=".5" fill="currentColor"/><circle cx="8.5" cy="7.5" r=".5" fill="currentColor"/><circle cx="6.5" cy="12.5" r=".5" fill="currentColor"/><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z"/>`,
	"postgresql":  `<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5V19A9 3 0 0 0 21 19V5"/><path d="M3 12A9 3 0 0 0 21 12"/>`,
	"sqlite":      `<line x1="22" x2="2" y1="12" y2="12"/><path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/><line x1="6" x2="6.01" y1="16" y2="16"/><line x1="10" x2="10.01" y1="16" y2="16"/>`,
	"vercel":      `<path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/>`,
	"cloudflare":  `<path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z"/>`,
	"github":      `<circle cx="12" cy="18" r="3"/><circle cx="6" cy="6" r="3"/><circle cx="18" cy="6" r="3"/><path d="M18 9v2c0 .6-.5 1-1 1H7c-.6 0-1-.4-1-1V9"/><path d="M12 12v3"/>`,
	"resend":      `<rect width="20" height="16" x="2" y="4" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/>`,
	"star":        `<polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>`,
	"vsCode":      `<path d="m18 16 4-4-4-4"/><path d="m6 8-4 4 4 4"/><path d="m14.5 4-5 16"/>`,
	"kiloCode":    `<rect width="16" height="16" x="4" y="4" rx="2"/><rect width="6" height="6" x="9" y="9"/><path d="M15 2v2"/><path d="M15 20v2"/><path d="M2 15h2"/><path d="M2 9h2"/><path d="M20 15h2"/><path d="M20 9h2"/><path d="M9 2v2"/><path d="M9 20v2"/>`,
}

// fillIcons are 24x24 brand icons rendered with fill="currentColor". Paths
// match the GithubBrandIcon/LinkedinBrandIcon components in the Next.js
// reference (unlocked-dashboard.tsx).
var fillIcons = map[string]string{
	"githubBrand":   `<path d="M12 .5C5.65.5.5 5.65.5 12a11.5 11.5 0 0 0 7.86 10.94c.58.11.79-.25.79-.56v-2c-3.2.7-3.88-1.37-3.88-1.37-.52-1.32-1.28-1.67-1.28-1.67-1.05-.72.08-.7.08-.7 1.16.08 1.77 1.2 1.77 1.2 1.03 1.77 2.7 1.26 3.36.96.1-.75.4-1.26.73-1.55-2.56-.29-5.25-1.28-5.25-5.7 0-1.26.45-2.29 1.19-3.1-.12-.29-.51-1.47.11-3.06 0 0 .97-.31 3.18 1.18a11 11 0 0 1 5.78 0c2.2-1.49 3.17-1.18 3.17-1.18.63 1.59.24 2.77.12 3.06.74.81 1.18 1.84 1.18 3.1 0 4.43-2.7 5.4-5.27 5.69.41.36.77 1.06.77 2.14v3.17c0 .31.21.68.8.56A11.5 11.5 0 0 0 23.5 12C23.5 5.65 18.35.5 12 .5Z"/>`,
	"linkedinBrand": `<path d="M20.45 20.45h-3.56v-5.57c0-1.33-.03-3.04-1.85-3.04-1.85 0-2.14 1.44-2.14 2.94v5.67H9.34V9h3.41v1.56h.05c.48-.9 1.64-1.85 3.38-1.85 3.61 0 4.27 2.38 4.27 5.47v6.27ZM5.34 7.43a2.06 2.06 0 1 1 0-4.12 2.06 2.06 0 0 1 0 4.12ZM7.12 20.45H3.56V9h3.56v11.45ZM22.22 0H1.77C.79 0 0 .77 0 1.72v20.56C0 23.23.79 24 1.77 24h20.45c.98 0 1.78-.77 1.78-1.72V1.72C24 .77 23.2 0 22.22 0Z"/>`,
}

// iconSVG renders a named inline SVG. Unknown names render nothing so a typo
// fails visibly (missing icon) rather than breaking the page.
func iconSVG(name, class string) template.HTML {
	if inner, ok := strokeIcons[name]; ok {
		return template.HTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" class="` + template.HTMLEscapeString(class) + `">` + inner + `</svg>`)
	}
	if inner, ok := fillIcons[name]; ok {
		return template.HTML(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" class="` + template.HTMLEscapeString(class) + `">` + inner + `</svg>`)
	}
	return ""
}

// Has reports whether a named template exists.
func (r *Renderer) Has(name string) bool {
	return r.tmpl.Lookup(name) != nil
}
