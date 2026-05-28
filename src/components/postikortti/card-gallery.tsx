"use client";

import Image from "next/image";

export interface TemplateConfig {
  id: string;
  nameKey: string; // key inside translations (e.g. "templates.winter")
  path: string;
  textColor: string;
  fontFamily: string;
  fontSize: number;
  yStart: number;
  xCenter: number;
  maxWidth: number;
  lineHeight: number;
}

export const TEMPLATES: TemplateConfig[] = [
  {
    id: "ystava",
    nameKey: "templates.ystava",
    path: "/assets/postcard-templates/ystava.png",
    textColor: "#9e3b5e",
    fontFamily: "Georgia, serif",
    fontSize: 32,
    yStart: 480,
    xCenter: 600,
    maxWidth: 850,
    lineHeight: 44,
  },
  {
    id: "kalja",
    nameKey: "templates.kalja",
    path: "/assets/postcard-templates/kalja.png",
    textColor: "#0c5b6b",
    fontFamily: "sans-serif",
    fontSize: 34,
    yStart: 490,
    xCenter: 600,
    maxWidth: 900,
    lineHeight: 46,
  },
  {
    id: "apina",
    nameKey: "templates.apina",
    path: "/assets/postcard-templates/apina.png",
    textColor: "#fff4bc",
    fontFamily: "sans-serif",
    fontSize: 32,
    yStart: 540,
    xCenter: 600,
    maxWidth: 900,
    lineHeight: 44,
  },
];

interface CardGalleryProps {
  selectedId: string;
  onSelect: (template: TemplateConfig) => void;
  t: (key: string) => string;
}

export function CardGallery({ selectedId, onSelect, t }: CardGalleryProps) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium text-foreground">{t("selectTemplate")}</h3>
      <div className="grid grid-cols-3 gap-3">
        {TEMPLATES.map((tpl) => {
          const isSelected = tpl.id === selectedId;
          return (
            <button
              key={tpl.id}
              type="button"
              onClick={() => onSelect(tpl)}
              className={`group relative aspect-[3/2] overflow-hidden rounded-xl border-2 text-left transition-all ${
                isSelected
                  ? "border-primary ring-2 ring-primary/20 scale-[1.02]"
                  : "border-muted bg-muted/30 hover:border-muted-foreground/40 hover:scale-[1.01]"
              }`}
            >
              <div className="absolute inset-0 bg-black/20 group-hover:bg-black/10 transition-colors z-10" />
              <Image
                src={tpl.path}
                alt={t(tpl.nameKey)}
                fill
                sizes="(max-width: 768px) 33vw, 200px"
                className="object-cover transition-transform duration-300 group-hover:scale-105"
                priority
              />
              <div className="absolute bottom-0 left-0 right-0 p-2 bg-gradient-to-t from-black/80 to-transparent z-20">
                <span className="text-xs font-semibold text-white truncate block">
                  {t(tpl.nameKey)}
                </span>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}
