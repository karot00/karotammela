"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";

interface CardFormProps {
  t: (key: string, values?: Record<string, string | number>) => string;
  onChange: (data: { senderName: string; recipientEmail: string; greeting: string }) => void;
  onDownload: () => void;
  onSend: (emailData: { senderName: string; recipientEmail: string; greeting: string; website?: string }) => void;
  status: "idle" | "sending" | "success" | "error" | "rateLimit";
  errorMessage?: string;
  successEmail?: string;
}

export function CardForm({
  t,
  onChange,
  onDownload,
  onSend,
  status,
  errorMessage,
  successEmail,
}: CardFormProps) {
  const searchParams = useSearchParams();

  // Initialize fields with URL search params if present
  const [senderName, setSenderName] = useState(() => searchParams.get("name") || "");
  const [recipientEmail, setRecipientEmail] = useState(() => searchParams.get("email") || "");
  const [greeting, setGreeting] = useState("");
  const [website, setWebsite] = useState("");

  const [errors, setErrors] = useState<{ senderName?: string; recipientEmail?: string; greeting?: string }>({});

  // Sync state upward when fields change
  useEffect(() => {
    onChange({ senderName, recipientEmail, greeting });
  }, [senderName, recipientEmail, greeting, onChange]);

  const validate = () => {
    const newErrors: typeof errors = {};

    if (!senderName.trim() || senderName.trim().length < 2) {
      newErrors.senderName = t("form.validation.sender");
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!recipientEmail.trim() || !emailRegex.test(recipientEmail)) {
      newErrors.recipientEmail = t("form.validation.email");
    }

    if (!greeting.trim()) {
      newErrors.greeting = t("form.validation.greeting");
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (status === "sending") return;

    if (validate()) {
      onSend({ senderName, recipientEmail, greeting, website });
    }
  };

  return (
    <form onSubmit={handleSend} className="space-y-4">
      {/* Honeypot field (hidden from users) */}
      <div className="absolute opacity-0 -z-10 pointer-events-none h-0 w-0 overflow-hidden">
        <label htmlFor="website">Website</label>
        <input
          id="website"
          type="text"
          value={website}
          tabIndex={-1}
          autoComplete="off"
          onChange={(e) => setWebsite(e.target.value)}
        />
      </div>

      {/* Sender Name */}
      <div className="space-y-1">
        <label htmlFor="senderName" className="block text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {t("form.senderName")}
        </label>
        <input
          id="senderName"
          type="text"
          value={senderName}
          maxLength={80}
          placeholder={t("form.senderNamePlaceholder")}
          onChange={(e) => {
            setSenderName(e.target.value);
            if (errors.senderName) setErrors((prev) => ({ ...prev, senderName: undefined }));
          }}
          className={`h-11 w-full rounded-xl border bg-background/50 px-4 text-sm outline-none transition-colors placeholder:text-muted-foreground/60 ${
            errors.senderName ? "border-destructive focus-visible:border-destructive" : "border-muted focus-visible:border-primary/60"
          }`}
        />
        {errors.senderName && (
          <p className="text-xs text-destructive mt-1">{errors.senderName}</p>
        )}
      </div>

      {/* Recipient Email */}
      <div className="space-y-1">
        <label htmlFor="recipientEmail" className="block text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {t("form.recipientEmail")}
        </label>
        <input
          id="recipientEmail"
          type="email"
          value={recipientEmail}
          maxLength={200}
          placeholder={t("form.recipientEmailPlaceholder")}
          onChange={(e) => {
            setRecipientEmail(e.target.value);
            if (errors.recipientEmail) setErrors((prev) => ({ ...prev, recipientEmail: undefined }));
          }}
          className={`h-11 w-full rounded-xl border bg-background/50 px-4 text-sm outline-none transition-colors placeholder:text-muted-foreground/60 ${
            errors.recipientEmail ? "border-destructive focus-visible:border-destructive" : "border-muted focus-visible:border-primary/60"
          }`}
        />
        {errors.recipientEmail && (
          <p className="text-xs text-destructive mt-1">{errors.recipientEmail}</p>
        )}
      </div>

      {/* Greeting Text */}
      <div className="space-y-1">
        <div className="flex justify-between items-center">
          <label htmlFor="greeting" className="block text-xs font-medium text-muted-foreground uppercase tracking-wider">
            {t("form.greeting")}
          </label>
          <span className="text-[10px] text-muted-foreground">
            {t("form.charLimit", { count: greeting.length, max: 250 })}
          </span>
        </div>
        <textarea
          id="greeting"
          value={greeting}
          maxLength={250}
          rows={3}
          placeholder={t("form.greetingPlaceholder")}
          onChange={(e) => {
            setGreeting(e.target.value);
            if (errors.greeting) setErrors((prev) => ({ ...prev, greeting: undefined }));
          }}
          className={`w-full min-h-[90px] resize-none rounded-xl border bg-background/50 px-4 py-3 text-sm outline-none transition-colors placeholder:text-muted-foreground/60 ${
            errors.greeting ? "border-destructive focus-visible:border-destructive" : "border-muted focus-visible:border-primary/60"
          }`}
        />
        {errors.greeting && (
          <p className="text-xs text-destructive mt-1">{errors.greeting}</p>
        )}
      </div>

      {/* Actions */}
      <div className="pt-2 grid grid-cols-2 gap-3">
        <Button
          type="button"
          onClick={onDownload}
          variant="outline"
          className="h-11 w-full rounded-xl border-muted-foreground/20 hover:bg-muted/10 font-semibold"
        >
          {t("actions.download")}
        </Button>
        <Button
          type="submit"
          disabled={status === "sending"}
          className="h-11 w-full rounded-xl bg-primary text-primary-foreground font-semibold hover:bg-primary/90 shadow-md shadow-primary/10 transition-transform active:scale-[0.98]"
        >
          {status === "sending" ? t("actions.sending") : t("actions.send")}
        </Button>
      </div>

      {/* Status Messages */}
      {status === "success" && (
        <div className="p-3.5 rounded-xl bg-primary/10 border border-primary/20 text-xs text-primary font-medium animate-in fade-in slide-in-from-top-1 duration-200">
          {t("actions.success", { email: successEmail || "" })}
        </div>
      )}
      {status === "error" && (
        <div className="p-3.5 rounded-xl bg-destructive/10 border border-destructive/20 text-xs text-destructive font-medium animate-in fade-in slide-in-from-top-1 duration-200">
          {t("actions.error", { error: errorMessage || "" })}
        </div>
      )}
      {status === "rateLimit" && (
        <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs text-amber-500 font-medium animate-in fade-in slide-in-from-top-1 duration-200">
          {t("actions.rateLimit")}
        </div>
      )}
    </form>
  );
}
