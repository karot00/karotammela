import { marked } from "marked";
import sanitizeHtml from "sanitize-html";

export function renderBlogMarkdownToSafeHtml(source: string) {
  const rendered = marked.parse(source, {
    breaks: true,
    gfm: true,
  });

  const html = typeof rendered === "string" ? rendered : "";

  return sanitizeHtml(html, {
    allowedTags: sanitizeHtml.defaults.allowedTags.concat(["img"]),
    allowedAttributes: {
      ...sanitizeHtml.defaults.allowedAttributes,
      a: ["href", "name", "target", "rel"],
      img: ["src", "alt", "title", "width", "height", "style", "align"],
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
