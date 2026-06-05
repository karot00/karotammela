import { marked } from "marked";
import sanitizeHtml from "sanitize-html";

/**
 * Render trusted, author-authored Markdown to sanitized HTML.
 *
 * This is the single source of truth for the Markdown sanitize allowlist used
 * across the site (blog posts, dashboard content). Keep the security-relevant
 * configuration here so the two render surfaces can never drift apart.
 */
export function renderMarkdownToSafeHtml(source: string) {
  const rendered = marked.parse(source, {
    breaks: true,
    gfm: true,
  });

  const html = typeof rendered === "string" ? rendered : "";

  return sanitizeHtml(html, {
    allowedTags: sanitizeHtml.defaults.allowedTags.concat(["img", "pre", "code"]),
    allowedAttributes: {
      ...sanitizeHtml.defaults.allowedAttributes,
      a: ["href", "name", "target", "rel"],
      img: ["src", "alt", "title", "width", "height", "style", "align"],
      code: ["class"],
      pre: ["class"],
    },
    allowedStyles: {
      img: {
        width: [/^\d+(px|%)$/],
        height: [/^\d+(px|%)$/],
        margin: [/^[\d\sa-z%.-]+$/i],
        display: [/^(inline|block)$/],
        float: [/^(left|right|none)$/],
        "max-width": [/^\d+(px|%)$/],
      },
    },
    transformTags: {
      a: sanitizeHtml.simpleTransform("a", {
        target: "_blank",
        rel: "noreferrer noopener",
      }),
    },
  });
}
