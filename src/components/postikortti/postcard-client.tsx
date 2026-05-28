"use client";

import { useState, useCallback } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { CardGallery, TEMPLATES, TemplateConfig } from "./card-gallery";
import { CardCanvas } from "./card-canvas";
import { CardForm } from "./card-form";

interface PostcardClientProps {
  locale: string;
}

export function PostcardClient({ locale }: PostcardClientProps) {
  const t = useTranslations("postcard");

  const [selectedTemplate, setSelectedTemplate] = useState<TemplateConfig>(TEMPLATES[0]);
  const [senderName, setSenderName] = useState("");
  const [greeting, setGreeting] = useState("");

  const [canvasElement, setCanvasElement] = useState<HTMLCanvasElement | null>(null);

  // Status for email sending
  const [status, setStatus] = useState<"idle" | "sending" | "success" | "error" | "rateLimit">("idle");
  const [errorMessage, setErrorMessage] = useState("");
  const [successEmail, setSuccessEmail] = useState("");

  const handleFormChange = useCallback((data: { senderName: string; recipientEmail: string; greeting: string }) => {
    setSenderName(data.senderName);
    setGreeting(data.greeting);
  }, []);

  const handleCanvasReady = useCallback((canvas: HTMLCanvasElement | null) => {
    setCanvasElement(canvas);
  }, []);

  const handleDownload = useCallback(() => {
    if (!canvasElement) return;

    try {
      const dataUrl = canvasElement.toDataURL("image/png");
      const link = document.createElement("a");
      link.download = `karotammela-postikortti-${selectedTemplate.id}.png`;
      link.href = dataUrl;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      console.error("Failed to generate download:", err);
    }
  }, [canvasElement, selectedTemplate.id]);

  const handleSend = useCallback(async (emailData: { senderName: string; recipientEmail: string; greeting: string; website?: string }) => {
    if (!canvasElement || status === "sending") return;

    setStatus("sending");
    setErrorMessage("");

    try {
      const base64Image = canvasElement.toDataURL("image/png");

      const response = await fetch("/api/send-card", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          senderName: emailData.senderName,
          recipientEmail: emailData.recipientEmail,
          greeting: emailData.greeting,
          base64Image,
          locale,
          website: emailData.website,
        }),
      });

      if (response.status === 429) {
        setStatus("rateLimit");
        return;
      }

      if (!response.ok) {
        const errData = await response.json().catch(() => ({}));
        setErrorMessage(errData.error || "Unknown error occurred");
        setStatus("error");
        return;
      }

      setSuccessEmail(emailData.recipientEmail);
      setStatus("success");
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : "Failed to contact send endpoint";
      setErrorMessage(errMsg);
      setStatus("error");
    }
  }, [canvasElement, status, locale]);

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 border-b border-muted pb-6">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight text-foreground sm:text-4xl">
            {t("title")}
          </h1>
          <p className="mt-2 text-muted-foreground max-w-2xl text-sm sm:text-base">
            {t("description")}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <LocaleSwitcher />
        </div>
      </div>

      {/* Back button */}
      <div>
        <Button asChild variant="ghost" className="h-auto p-0 text-sm text-muted-foreground hover:text-foreground">
          <Link href={`/${locale}/dashboard`}>
            ← {locale === "fi" ? "Takaisin dashboardille" : "Back to dashboard"}
          </Link>
        </Button>
      </div>

      {/* Two Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left column: Preview Canvas */}
        <div className="lg:col-span-7 space-y-4">
          <CardCanvas
            template={selectedTemplate}
            senderName={senderName}
            greeting={greeting}
            locale={locale}
            onCanvasReady={handleCanvasReady}
          />
          <p className="text-[11px] text-center text-muted-foreground italic">
            {locale === "fi"
              ? "Esikatselu näyttää kortin sellaisena kuin se ladataan tai lähetetään sähköpostitse."
              : "The preview shows the card exactly as it will be downloaded or emailed."}
          </p>
        </div>

        {/* Right column: Gallery & Customizer form */}
        <div className="lg:col-span-5 space-y-6 bg-card/40 border border-muted/60 p-6 rounded-2xl shadow-sm">
          <CardGallery
            selectedId={selectedTemplate.id}
            onSelect={setSelectedTemplate}
            t={t}
          />

          <div className="border-t border-muted/60 my-6" />

          <CardForm
            t={t}
            onChange={handleFormChange}
            onDownload={handleDownload}
            onSend={handleSend}
            status={status}
            errorMessage={errorMessage}
            successEmail={successEmail}
          />
        </div>
      </div>
    </div>
  );
}
