"use client";

import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { motion, useReducedMotion } from "framer-motion";
import { useLocale } from "next-intl";
import Link from "next/link";

import { Button } from "@/components/ui/button";

type Message = {
  role: "assistant" | "user";
  content: string;
};

type SentinelApiResponse = {
  message: string;
  level: number;
  unlocked: boolean;
  accessCode: string | null;
};

type SentinelTerminalProps = {
  id?: string;
  title: string;
  subtitle: string;
  meterLabel: string;
  inputPlaceholder: string;
  sendLabel: string;
  resetLabel: string;
  initialMessage: string;
  unlockedLabel: string;
  unlockedCta: string;
  pendingLabel: string;
  errorLabel: string;
  hasUnlockedBefore: boolean;
  returnOverlayTitle: string;
  returnOverlayBody: string;
  playAgainLabel: string;
  goDashboardLabel: string;
};

export function SentinelTerminal({
  id = "sentinel",
  title,
  subtitle,
  meterLabel,
  inputPlaceholder,
  sendLabel,
  resetLabel,
  initialMessage,
  unlockedLabel,
  unlockedCta,
  pendingLabel,
  errorLabel,
  hasUnlockedBefore,
  returnOverlayTitle,
  returnOverlayBody,
  playAgainLabel,
  goDashboardLabel,
}: SentinelTerminalProps) {
  const prefersReducedMotion = useReducedMotion();
  const locale = useLocale();
  const [sessionId] = useState(() => crypto.randomUUID());
  const [messages, setMessages] = useState<Message[]>([
    { role: "assistant", content: initialMessage },
  ]);
  const [input, setInput] = useState("");
  const [level, setLevel] = useState(0);
  const [isUnlocked, setIsUnlocked] = useState(false);
  const [showReturnOverlay, setShowReturnOverlay] = useState(hasUnlockedBefore);
  const [isPending, setIsPending] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const logContainerRef = useRef<HTMLDivElement>(null);
  const hasMountedRef = useRef(false);

  useEffect(() => {
    if (!hasMountedRef.current) {
      hasMountedRef.current = true;
      return;
    }

    const container = logContainerRef.current;
    if (!container) return;

    container.scrollTo({
      top: container.scrollHeight,
      behavior: "smooth",
    });
  }, [messages]);

  const meterToneClass = useMemo(() => {
    if (level >= 71) return "bg-accent";
    if (level >= 31) return "bg-primary";
    return "bg-muted-foreground";
  }, [level]);

  const onSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const value = input.trim();
    if (!value || isPending) return;

    const userMessage: Message = { role: "user", content: value };
    const nextMessages = [...messages, userMessage];

    setMessages(nextMessages);
    setInput("");
    setErrorMessage(null);
    setIsPending(true);

    try {
      const response = await fetch("/api/sentinel", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          locale,
          sessionId,
          currentLevel: level,
          messages: nextMessages,
        }),
      });

      if (!response.ok) {
        setErrorMessage(errorLabel);
        return;
      }

      const payload = (await response.json()) as SentinelApiResponse;
      const assistantMessage: Message = {
        role: "assistant",
        content: `${payload.message} [LEVEL:${payload.level}]`,
      };

      setMessages((current) => [...current, assistantMessage]);
      setLevel(payload.level);
      setIsUnlocked(payload.unlocked);
    } catch {
      setErrorMessage(errorLabel);
    } finally {
      setIsPending(false);
    }
  };

  const reset = () => {
    setMessages([{ role: "assistant", content: initialMessage }]);
    setInput("");
    setLevel(0);
    setIsUnlocked(false);
    setIsPending(false);
    setErrorMessage(null);
  };

  return (
    <section
      id={id}
      className="relative rounded-3xl border border-border/70 bg-popover/50 p-6 shadow-2xl backdrop-blur-md sm:p-8"
    >
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold text-foreground sm:text-3xl">{title}</h2>
          <p className="mt-2 text-sm text-muted-foreground sm:text-base">{subtitle}</p>
        </div>
        {isUnlocked ? (
          <div className="flex items-center gap-3">
            <p className="rounded-md border border-accent/50 bg-accent/15 px-3 py-1.5 text-xs font-semibold tracking-[0.14em] text-accent uppercase">
              {unlockedLabel}
            </p>
            <Link
              href={`/${locale}/dashboard`}
              className="text-xs font-semibold tracking-[0.14em] text-primary uppercase hover:text-primary/80"
            >
              {unlockedCta}
            </Link>
          </div>
        ) : null}
      </div>

      <div className="mt-6">
        <div className="mb-2 flex items-center justify-between text-xs uppercase tracking-[0.18em] text-muted-foreground">
          <span>{meterLabel}</span>
          <span>{level}%</span>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-secondary">
          <motion.div
            className={`h-full ${meterToneClass}`}
            initial={false}
            animate={{ width: `${level}%` }}
            transition={prefersReducedMotion ? { duration: 0 } : { type: "spring", stiffness: 110, damping: 22 }}
          />
        </div>
      </div>

      <div
        ref={logContainerRef}
        className="mt-6 max-h-64 overflow-y-auto scroll-smooth rounded-2xl border border-border/60 bg-background/70 p-4 font-mono text-sm"
      >
        <div className="space-y-3">
          {messages.map((message, index) => (
            <motion.p
              key={`${message.role}-${index}`}
              initial={prefersReducedMotion ? false : { opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.2 }}
              className={message.role === "assistant" ? "text-primary" : "text-foreground"}
            >
              <span className="mr-2 text-muted-foreground">
                {message.role === "assistant" ? "sentinel>" : "human>"}
              </span>
              {message.content}
            </motion.p>
          ))}
        </div>
      </div>

      <form onSubmit={onSubmit} className="mt-4 flex flex-col gap-3 sm:flex-row">
        <input
          value={input}
          onChange={(event) => setInput(event.target.value)}
          placeholder={inputPlaceholder}
          className="h-11 flex-1 rounded-lg border border-input bg-background px-3 text-sm text-foreground outline-none ring-0 placeholder:text-muted-foreground focus-visible:border-primary/60"
          disabled={isPending}
        />
        <Button type="submit" className="h-11 px-5">
          {isPending ? pendingLabel : sendLabel}
        </Button>
        <Button type="button" variant="outline" className="h-11 px-5" onClick={reset}>
          {resetLabel}
        </Button>
      </form>

      {errorMessage ? <p className="mt-3 text-sm text-destructive">{errorMessage}</p> : null}

      {showReturnOverlay ? (
        <div className="absolute inset-0 z-10 flex items-center justify-center rounded-3xl bg-background/55 p-6 backdrop-blur-md">
          <div className="w-full max-w-xl rounded-2xl border border-border/70 bg-card/80 p-6 text-center shadow-2xl">
            <p className="text-sm font-semibold tracking-[0.14em] text-primary uppercase">{returnOverlayTitle}</p>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground sm:text-base">{returnOverlayBody}</p>
            <div className="mt-5 flex flex-col gap-3 sm:flex-row sm:justify-center">
              <Button type="button" onClick={() => setShowReturnOverlay(false)} className="h-11 px-5">
                {playAgainLabel}
              </Button>
              <Button type="button" variant="outline" className="h-11 px-5" asChild>
                <Link href={`/${locale}/dashboard`}>{goDashboardLabel}</Link>
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}
