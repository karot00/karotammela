"use client";

import { useEffect, useRef, useState } from "react";
import { TemplateConfig } from "./card-gallery";

interface CardCanvasProps {
  template: TemplateConfig;
  senderName: string;
  greeting: string;
  locale: string;
  onCanvasReady: (canvas: HTMLCanvasElement | null) => void;
}

export function CardCanvas({
  template,
  senderName,
  greeting,
  locale,
  onCanvasReady,
}: CardCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [imageLoaded, setImageLoaded] = useState<string | null>(null);
  const imageCacheRef = useRef<Record<string, HTMLImageElement>>({});

  // Trigger parent callback when canvas updates
  useEffect(() => {
    onCanvasReady(canvasRef.current);
  }, [onCanvasReady, template, senderName, greeting, imageLoaded]);

  // Handle template image loading & caching
  useEffect(() => {
    const path = template.path;
    if (imageCacheRef.current[path]) {
      if (imageLoaded !== path) {
        const t = setTimeout(() => {
          setImageLoaded(path);
        }, 0);
        return () => clearTimeout(t);
      }
      return;
    }

    const img = new Image();
    img.crossOrigin = "anonymous"; // prevent CORS dirty canvas issues
    img.src = path;
    img.onload = () => {
      imageCacheRef.current[path] = img;
      setImageLoaded(path);
    };
  }, [template.path, imageLoaded]);

  // Main canvas draw loop
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const img = imageCacheRef.current[template.path];
    if (!img) return;

    // Clear canvas
    ctx.clearRect(0, 0, 1200, 800);

    // Draw background image
    ctx.drawImage(img, 0, 0, 1200, 800);

    // Prepare text options
    ctx.fillStyle = template.textColor;
    ctx.textAlign = "center";
    ctx.textBaseline = "top";

    // Set font size & style
    ctx.font = `italic ${template.fontSize}px ${template.fontFamily}`;

    // Word wrapping and line splitting helper
    const maxWidth = template.maxWidth;
    const lineHeight = template.lineHeight;
    let y = template.yStart;

    // Split text into paragraphs first (support user line breaks)
    const paragraphs = greeting.split("\n");
    const lines: string[] = [];

    for (const paragraph of paragraphs) {
      if (paragraph.trim() === "") {
        lines.push(""); // empty line
        continue;
      }

      const words = paragraph.split(" ");
      let currentLine = "";

      for (let i = 0; i < words.length; i++) {
        const word = words[i];
        const testLine = currentLine === "" ? word : currentLine + " " + word;
        const metrics = ctx.measureText(testLine);

        if (metrics.width > maxWidth && i > 0) {
          lines.push(currentLine);
          currentLine = word;
        } else {
          currentLine = testLine;
        }
      }
      if (currentLine !== "") {
        lines.push(currentLine);
      }
    }

    // Measure total block height to vertically center if possible
    // (We limit text Y to avoid running off the bottom)
    const totalTextHeight = lines.length * lineHeight;
    // Let's adjust y starting point slightly to center the whole text block within the available space
    const maxAvailableHeight = 700 - template.yStart;
    if (totalTextHeight < maxAvailableHeight) {
      // Keep it near the template's yStart, or center it nicely
      y = template.yStart + (maxAvailableHeight - totalTextHeight) / 3;
    }

    // Render greeting lines
    for (const line of lines) {
      if (line !== "") {
        ctx.fillText(line, template.xCenter, y);
      }
      y += lineHeight;
    }

    // Add elegant signature if sender name is provided
    if (senderName.trim()) {
      y += lineHeight * 0.8; // extra margin
      ctx.font = `bold italic ${template.fontSize - 4}px ${template.fontFamily}`;
      const signOff = locale === "fi" ? "Terveisin," : "Best regards,";
      ctx.fillText(signOff, template.xCenter, y);
      y += lineHeight * 0.9;
      ctx.fillText(senderName, template.xCenter, y);
    }
  }, [template, senderName, greeting, imageLoaded, locale]);

  return (
    <div className="relative w-full overflow-hidden rounded-2xl border border-muted bg-neutral-950/20 shadow-xl aspect-[3/2] flex items-center justify-center">
      {/* High resolution canvas rendered at 1200x800px, scaled reactively via CSS */}
      <canvas
        ref={canvasRef}
        width={1200}
        height={800}
        className="w-full h-full object-contain bg-muted/20"
      />
    </div>
  );
}
