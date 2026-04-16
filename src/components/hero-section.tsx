import Image from "next/image";

type HeroSectionProps = {
  badge: string;
  intro: string;
  body1: string;
  body2: string;
  body3: string;
};

export function HeroSection({
  badge,
  intro,
  body1,
  body2,
  body3,
}: HeroSectionProps) {
  const videoEnabled = process.env.NEXT_PUBLIC_HERO_VIDEO_ENABLED === "1";
  const showVideo = videoEnabled;

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

      <div className="relative">
        <div>
          <p className="text-xs font-semibold tracking-[0.24em] text-primary uppercase">
            {badge}
          </p>

          <p className="mt-4 max-w-3xl text-lg font-medium leading-relaxed text-foreground sm:text-2xl sm:leading-snug">
            {intro}
          </p>

          <div className="mt-6 max-w-2xl space-y-4 text-base leading-relaxed text-muted-foreground sm:text-lg">
            <p>{body1}</p>
            <p>{body2}</p>
            <p>{body3}</p>
          </div>
        </div>
      </div>
    </section>
  );
}
