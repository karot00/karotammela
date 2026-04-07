"use client";

import { motion, useReducedMotion } from "framer-motion";
import Image from "next/image";

import { Button } from "@/components/ui/button";

type HeroSectionProps = {
  badge: string;
  title: string;
  description: string;
  primaryCta: string;
  secondaryCta: string;
};

export function HeroSection({
  badge,
  title,
  description,
  primaryCta,
  secondaryCta,
}: HeroSectionProps) {
  const prefersReducedMotion = useReducedMotion();
  const videoEnabled = process.env.NEXT_PUBLIC_HERO_VIDEO_ENABLED === "1";
  const showVideo = !prefersReducedMotion && videoEnabled;

  return (
    <section className="relative overflow-hidden rounded-3xl border border-border/60 bg-card/50 p-8 backdrop-blur-md sm:p-12">
      <div className="pointer-events-none absolute inset-0">
        {showVideo ? (
          <video
            className="absolute inset-0 h-full w-full object-cover opacity-35"
            autoPlay
            loop
            muted
            playsInline
            preload="metadata"
            poster="/media/hero-poster.svg"
            aria-hidden="true"
          >
            <source src="/media/hero-loop.webm" type="video/webm" />
            <source src="/media/hero-loop.mp4" type="video/mp4" />
          </video>
        ) : (
          <Image
            src="/media/hero-poster.svg"
            alt=""
            aria-hidden="true"
            className="absolute inset-0 h-full w-full object-cover opacity-40"
            priority
            fill
            sizes="100vw"
          />
        )}

        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_right,rgba(0,210,255,0.2),transparent_52%),radial-gradient(circle_at_20%_90%,rgba(255,140,0,0.12),transparent_45%)]" />
        <div className="absolute inset-0 bg-gradient-to-b from-background/20 via-background/35 to-background/70" />
      </div>

      <motion.div
        initial={prefersReducedMotion ? false : { opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.65, ease: "easeOut" }}
        className="relative"
      >
        <p className="text-xs font-semibold tracking-[0.24em] text-primary uppercase">
          {badge}
        </p>

        <h1 className="mt-4 max-w-3xl text-3xl font-semibold leading-tight text-foreground sm:text-5xl sm:leading-[1.08]">
          {title}
        </h1>

        <p className="mt-6 max-w-2xl text-base leading-relaxed text-muted-foreground sm:text-lg">
          {description}
        </p>

        <div className="mt-10 flex flex-wrap gap-3">
          <Button asChild size="lg" className="border border-primary/50 bg-primary/15 text-primary hover:bg-primary/25">
            <a href="#sentinel">{primaryCta}</a>
          </Button>

          <Button
            asChild
            size="lg"
            variant="outline"
            className="border-border/70 bg-secondary/40 text-secondary-foreground hover:bg-secondary/70"
          >
            <a href="#sentinel">{secondaryCta}</a>
          </Button>
        </div>
      </motion.div>
    </section>
  );
}
