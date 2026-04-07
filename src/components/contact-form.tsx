"use client";

import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";

type ContactFormCopy = {
  nameLabel: string;
  emailLabel: string;
  companyLabel: string;
  messageLabel: string;
  submitLabel: string;
  pendingLabel: string;
  successLabel: string;
  errorLabel: string;
};

type ContactFormProps = {
  copy: ContactFormCopy;
};

export function ContactForm({ copy }: ContactFormProps) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [company, setCompany] = useState("");
  const [message, setMessage] = useState("");
  const [website, setWebsite] = useState("");
  const [status, setStatus] = useState<"idle" | "pending" | "success" | "error">("idle");

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (status === "pending") return;

    setStatus("pending");

    try {
      const response = await fetch("/api/contact", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name,
          email,
          company,
          message,
          website,
        }),
      });

      if (!response.ok) {
        setStatus("error");
        return;
      }

      setName("");
      setEmail("");
      setCompany("");
      setMessage("");
      setWebsite("");
      setStatus("success");
    } catch {
      setStatus("error");
    }
  };

  return (
    <form className="mt-6 space-y-3" onSubmit={onSubmit}>
      <input
        aria-hidden="true"
        tabIndex={-1}
        autoComplete="off"
        className="hidden"
        value={website}
        onChange={(event) => setWebsite(event.target.value)}
        name="website"
      />

      <label className="block text-xs text-muted-foreground">
        {copy.nameLabel}
        <input
          required
          value={name}
          onChange={(event) => setName(event.target.value)}
          className="mt-1 h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:border-primary/60"
        />
      </label>

      <label className="block text-xs text-muted-foreground">
        {copy.emailLabel}
        <input
          type="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="mt-1 h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:border-primary/60"
        />
      </label>

      <label className="block text-xs text-muted-foreground">
        {copy.companyLabel}
        <input
          value={company}
          onChange={(event) => setCompany(event.target.value)}
          className="mt-1 h-10 w-full rounded-lg border border-input bg-background px-3 text-sm outline-none focus-visible:border-primary/60"
        />
      </label>

      <label className="block text-xs text-muted-foreground">
        {copy.messageLabel}
        <textarea
          required
          minLength={10}
          value={message}
          onChange={(event) => setMessage(event.target.value)}
          className="mt-1 min-h-24 w-full resize-y rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus-visible:border-primary/60"
        />
      </label>

      <Button type="submit" className="h-10 w-full" disabled={status === "pending"}>
        {status === "pending" ? copy.pendingLabel : copy.submitLabel}
      </Button>

      {status === "success" ? <p className="text-xs text-primary">{copy.successLabel}</p> : null}
      {status === "error" ? <p className="text-xs text-destructive">{copy.errorLabel}</p> : null}
    </form>
  );
}
