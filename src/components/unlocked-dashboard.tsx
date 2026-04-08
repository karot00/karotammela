"use client";

import Link from "next/link";
import Image from "next/image";
import { useEffect, useState } from "react";
import { Menu } from "lucide-react";

import { ContactForm } from "@/components/contact-form";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { Button } from "@/components/ui/button";

type DashboardStats = {
  totalAttempts: number;
  unlockedCount: number;
  highestLevel: number;
  latestUnlockAt: string | null;
};

type DashboardCopy = {
  badge: string;
  title: string;
  description: string;
  navigationLabel: string;
  navOverview: string;
  navProjects: string;
  navTech: string;
  navBlog: string;
  navSettings: string;
  menuLabel: string;
  homeLinkLabel: string;
  aboutTitle: string;
  aboutBody: string;
  profileImageAlt: string;
  toolkitTitle: string;
  codingModelsLabel: string;
  chatbotModelLabel: string;
  analyticsTitle: string;
  totalAttemptsLabel: string;
  unlockedCountLabel: string;
  highestLevelLabel: string;
  latestUnlockLabel: string;
  latestUnlockNever: string;
  sourceOffline: string;
  contactTitle: string;
  contactDescription: string;
  projectsTitle: string;
  projectOneTitle: string;
  projectOneDescription: string;
  projectTwoTitle: string;
  projectTwoDescription: string;
  projectThreeTitle: string;
  projectThreeDescription: string;
  projectFourTitle: string;
  projectFourDescription: string;
  projectGithubLabel: string;
  techTitle: string;
  blogPlaceholder: string;
  settingsLanguageTitle: string;
  settingsLanguageDescription: string;
  settingsThemeTitle: string;
  settingsThemeDescription: string;
  lightModeLabel: string;
  darkModeLabel: string;
  contactNameLabel: string;
  contactEmailLabel: string;
  contactCompanyLabel: string;
  contactMessageLabel: string;
  contactSubmitLabel: string;
  contactPendingLabel: string;
  contactSuccessLabel: string;
  contactErrorLabel: string;
};

type UnlockedDashboardProps = {
  locale: string;
  copy: DashboardCopy;
  stats: DashboardStats | null;
};

type DashboardView = "overview" | "projects" | "tech" | "blog" | "settings";
type ThemeMode = "dark" | "light";

function formatDate(value: string, locale: string) {
  const localeCode = locale === "fi" ? "fi-FI" : "en-US";
  const parsed = new Date(value);

  if (Number.isNaN(parsed.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat(localeCode, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(parsed);
}

function SidebarNav({
  activeView,
  copy,
  onSelect,
  onClose,
}: {
  activeView: DashboardView;
  copy: DashboardCopy;
  onSelect: (view: DashboardView) => void;
  onClose?: () => void;
}) {
  const navItems: Array<{ id: DashboardView; label: string }> = [
    { id: "overview", label: copy.navOverview },
    { id: "projects", label: copy.navProjects },
    { id: "tech", label: copy.navTech },
    { id: "blog", label: copy.navBlog },
    { id: "settings", label: copy.navSettings },
  ];

  return (
    <nav className="space-y-1">
      {navItems.map((item) => {
        const isActive = item.id === activeView;

        return (
          <button
            key={item.id}
            type="button"
            onClick={() => {
              onSelect(item.id);
              onClose?.();
            }}
            className={`w-full rounded-lg px-3 py-2 text-left text-sm transition ${
              isActive
                ? "bg-sidebar-primary text-sidebar-primary-foreground"
                : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
            }`}
          >
            {item.label}
          </button>
        );
      })}
    </nav>
  );
}

function OverviewView({ locale, copy, stats }: { locale: string; copy: DashboardCopy; stats: DashboardStats | null }) {
  const latestUnlockText = stats?.latestUnlockAt
    ? formatDate(stats.latestUnlockAt, locale)
    : copy.latestUnlockNever;

  return (
    <div className="space-y-6">
      <section className="rounded-xl border border-border bg-card p-6">
        <div className="grid gap-6 md:grid-cols-[220px_1fr] md:items-start">
          <div className="relative h-64 overflow-hidden rounded-xl border border-border bg-muted/30">
            <Image
              src="/media/Karo%20Tammela.jpg"
              alt={copy.profileImageAlt}
              fill
              className="object-cover"
              sizes="220px"
            />
          </div>

          <div>
            <h2 className="text-xl font-semibold text-foreground">{copy.aboutTitle}</h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{copy.aboutBody}</p>

            <div className="mt-5">
              <p className="text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">{copy.toolkitTitle}</p>
              <div className="mt-3 flex flex-wrap gap-2">
                <span className="rounded-full border border-border bg-muted px-3 py-1 text-xs text-foreground">{copy.codingModelsLabel}</span>
                <span className="rounded-full border border-border bg-muted px-3 py-1 text-xs text-foreground">{copy.chatbotModelLabel}</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="space-y-3">
        <h3 className="text-lg font-semibold text-foreground">{copy.analyticsTitle}</h3>

        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">{copy.totalAttemptsLabel}</p>
            <p className="mt-2 text-2xl font-semibold text-foreground">{stats?.totalAttempts ?? 0}</p>
          </article>

          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">{copy.unlockedCountLabel}</p>
            <p className="mt-2 text-2xl font-semibold text-foreground">{stats?.unlockedCount ?? 0}</p>
          </article>

          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">{copy.highestLevelLabel}</p>
            <p className="mt-2 text-2xl font-semibold text-foreground">{stats?.highestLevel ?? 0}%</p>
          </article>

          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">{copy.latestUnlockLabel}</p>
            <p className="mt-2 text-sm font-medium text-foreground">{latestUnlockText}</p>
          </article>
        </div>

        {!stats ? <p className="text-xs text-destructive">{copy.sourceOffline}</p> : null}
      </section>

      <section className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground">{copy.contactTitle}</h3>
        <p className="mt-2 text-sm text-muted-foreground">{copy.contactDescription}</p>
        <ContactForm
          copy={{
            nameLabel: copy.contactNameLabel,
            emailLabel: copy.contactEmailLabel,
            companyLabel: copy.contactCompanyLabel,
            messageLabel: copy.contactMessageLabel,
            submitLabel: copy.contactSubmitLabel,
            pendingLabel: copy.contactPendingLabel,
            successLabel: copy.contactSuccessLabel,
            errorLabel: copy.contactErrorLabel,
          }}
        />
      </section>
    </div>
  );
}

function ProjectsView({ copy }: { copy: DashboardCopy }) {
  const projects = [
    {
      title: copy.projectOneTitle,
      description: copy.projectOneDescription,
    },
    {
      title: copy.projectTwoTitle,
      description: copy.projectTwoDescription,
    },
    {
      title: copy.projectThreeTitle,
      description: copy.projectThreeDescription,
    },
    {
      title: copy.projectFourTitle,
      description: copy.projectFourDescription,
      githubHref: "https://github.com/karot00/karotammela",
      githubLabel: copy.projectGithubLabel,
    },
  ];

  return (
    <section className="space-y-4">
      <h2 className="text-xl font-semibold text-foreground">{copy.projectsTitle}</h2>

      <div className="grid gap-4 sm:grid-cols-2">
        {projects.map((project) => (
          <article key={project.title} className="rounded-xl border border-border bg-card p-5">
            <h3 className="font-semibold text-foreground">{project.title}</h3>
            <p className="mt-2 text-sm text-muted-foreground">{project.description}</p>

            {project.githubHref ? (
              <Button asChild size="sm" variant="outline" className="mt-4">
                <Link href={project.githubHref} target="_blank" rel="noreferrer">
                  {project.githubLabel}
                </Link>
              </Button>
            ) : null}
          </article>
        ))}
      </div>
    </section>
  );
}

function TechView({ copy }: { copy: DashboardCopy }) {
  const stack = [
    { label: "Next.js", href: "https://nextjs.org" },
    { label: "Tailwind CSS", href: "https://tailwindcss.com" },
    { label: "Vercel", href: "https://vercel.com" },
    { label: "GitHub", href: "https://github.com" },
    { label: "Cloudflare", href: "https://www.cloudflare.com" },
    { label: "PostgreSQL", href: "https://supabase.com" },
    { label: "SQLite", href: "https://turso.tech" },
    { label: "Resend", href: "https://resend.com" },
    { label: "VS Code", href: "https://code.visualstudio.com" },
    { label: "Kilo Code", href: "https://kilocode.ai" },
  ];

  return (
    <section className="space-y-4">
      <h2 className="text-xl font-semibold text-foreground">{copy.techTitle}</h2>

      <div className="flex flex-wrap gap-2">
        {stack.map((item) => (
          <Link
            key={item.label}
            href={item.href}
            target="_blank"
            rel="noreferrer"
            className="rounded-full border border-border bg-muted px-3 py-1.5 text-xs text-foreground transition hover:bg-accent hover:text-accent-foreground"
          >
            {item.label}
          </Link>
        ))}
      </div>
    </section>
  );
}

function BlogView({ copy }: { copy: DashboardCopy }) {
  return (
    <section className="flex min-h-[320px] items-center justify-center rounded-xl border border-border bg-card p-6">
      <p className="text-base text-muted-foreground">{copy.blogPlaceholder}</p>
    </section>
  );
}

function SettingsView({
  copy,
  themeMode,
  onThemeChange,
}: {
  copy: DashboardCopy;
  themeMode: ThemeMode;
  onThemeChange: (mode: ThemeMode) => void;
}) {
  return (
    <section className="space-y-6">
      <article className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold text-foreground">{copy.settingsLanguageTitle}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{copy.settingsLanguageDescription}</p>
        <div className="mt-4">
          <LocaleSwitcher />
        </div>
      </article>

      <article className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold text-foreground">{copy.settingsThemeTitle}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{copy.settingsThemeDescription}</p>

        <div className="mt-4 flex flex-wrap gap-2">
          <Button
            type="button"
            variant={themeMode === "light" ? "default" : "outline"}
            onClick={() => onThemeChange("light")}
          >
            {copy.lightModeLabel}
          </Button>
          <Button
            type="button"
            variant={themeMode === "dark" ? "default" : "outline"}
            onClick={() => onThemeChange("dark")}
          >
            {copy.darkModeLabel}
          </Button>
        </div>
      </article>
    </section>
  );
}

export function UnlockedDashboard({ locale, copy, stats }: UnlockedDashboardProps) {
  const [activeView, setActiveView] = useState<DashboardView>("overview");
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    if (typeof window === "undefined") {
      return "dark";
    }

    return window.localStorage.getItem("karot-theme") === "light" ? "light" : "dark";
  });

  useEffect(() => {
    document.documentElement.classList.toggle("light", themeMode === "light");
  }, [themeMode]);

  const onThemeChange = (mode: ThemeMode) => {
    setThemeMode(mode);
    window.localStorage.setItem("karot-theme", mode);
    document.documentElement.classList.toggle("light", mode === "light");
  };

  let view = <OverviewView locale={locale} copy={copy} stats={stats} />;
  if (activeView === "projects") view = <ProjectsView copy={copy} />;
  if (activeView === "tech") view = <TechView copy={copy} />;
  if (activeView === "blog") view = <BlogView copy={copy} />;
  if (activeView === "settings") view = <SettingsView copy={copy} themeMode={themeMode} onThemeChange={onThemeChange} />;

  return (
    <main className="flex min-h-screen bg-background">
      <aside className="hidden w-72 border-r border-sidebar-border bg-sidebar px-4 py-6 md:block">
        <p className="px-3 text-xs font-semibold tracking-[0.16em] text-sidebar-foreground/70 uppercase">{copy.navigationLabel}</p>
        <div className="mt-4">
          <SidebarNav activeView={activeView} copy={copy} onSelect={setActiveView} />
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 border-b border-border bg-background/95 px-4 py-3 backdrop-blur supports-[backdrop-filter]:bg-background/85 sm:px-6">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2 md:hidden">
              <Button type="button" variant="outline" size="icon" onClick={() => setMobileMenuOpen((open) => !open)}>
                <Menu className="size-4" />
              </Button>
              <span className="text-sm font-medium text-foreground">{copy.menuLabel}</span>
            </div>

            <div className="min-w-0">
              <p className="text-xs tracking-[0.12em] text-muted-foreground uppercase">{copy.badge}</p>
              <h1 className="truncate text-lg font-semibold text-foreground sm:text-xl">{copy.title}</h1>
              {copy.description ? <p className="hidden text-xs text-muted-foreground sm:block">{copy.description}</p> : null}
            </div>

            <Link href={`/${locale}`} className="text-xs font-medium text-primary hover:text-primary/80">
              {copy.homeLinkLabel}
            </Link>
          </div>

          {mobileMenuOpen ? (
            <div className="mt-3 rounded-xl border border-border bg-card p-3 md:hidden">
              <SidebarNav
                activeView={activeView}
                copy={copy}
                onSelect={setActiveView}
                onClose={() => setMobileMenuOpen(false)}
              />
            </div>
          ) : null}
        </header>

        <section className="flex-1 px-4 py-6 sm:px-6">{view}</section>
      </div>
    </main>
  );
}
