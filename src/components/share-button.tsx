"use client";

import { useState } from "react";

import { Button } from "@/components/ui/button";

type ShareButtonProps = {
  url: string;
  title: string;
  text?: string;
  label: string;
  copiedLabel: string;
  errorLabel: string;
  className?: string;
};

export function ShareButton({
  url,
  title,
  text,
  label,
  copiedLabel,
  errorLabel,
  className,
}: ShareButtonProps) {
  const [status, setStatus] = useState<"idle" | "copied" | "error">("idle");

  const resolveShareUrl = () => {
    if (url.startsWith("http://") || url.startsWith("https://")) {
      return url;
    }

    if (typeof window !== "undefined") {
      return `${window.location.origin}${url.startsWith("/") ? url : `/${url}`}`;
    }

    return url;
  };

  const onShare = async () => {
    const shareUrl = resolveShareUrl();

    try {
      if (typeof navigator !== "undefined" && "share" in navigator) {
        await navigator.share({
          title,
          text,
          url: shareUrl,
        });
        setStatus("idle");
        return;
      }

      if (typeof navigator !== "undefined" && navigator.clipboard) {
        await navigator.clipboard.writeText(shareUrl);
        setStatus("copied");
        window.setTimeout(() => setStatus("idle"), 2500);
        return;
      }

      setStatus("error");
      window.setTimeout(() => setStatus("idle"), 2500);
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") {
        return;
      }

      setStatus("error");
      window.setTimeout(() => setStatus("idle"), 2500);
    }
  };

  return (
    <div className={className}>
      <Button type="button" variant="outline" size="sm" onClick={onShare}>
        {label}
      </Button>
      <p
        className="mt-1 min-h-4 text-xs text-muted-foreground"
        aria-live="polite"
      >
        {status === "copied"
          ? copiedLabel
          : status === "error"
            ? errorLabel
            : ""}
      </p>
    </div>
  );
}
