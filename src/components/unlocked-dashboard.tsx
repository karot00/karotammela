"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import {
  Cloud,
  Code2,
  Cpu,
  Database,
  GitFork,
  type LucideIcon,
  HardDrive,
  Layout,
  Mail,
  Menu,
  Palette,
  Shield,
} from "lucide-react";
import { marked } from "marked";
import sanitizeHtml from "sanitize-html";

import {
  AiPulseShell,
  type AiPulseData,
} from "@/components/ai-pulse/ai-pulse-shell";
import { ContactForm } from "@/components/contact-form";
import { CookieConsentSettingsTrigger } from "@/components/cookie-consent-settings-trigger";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { ShareButton } from "@/components/share-button";
import { Button } from "@/components/ui/button";
import type { ChangelogChangeType, ChangelogRelease } from "@/lib/changelog";

type DashboardStats = {
  totalAttempts: number;
  unlockedCount: number;
  directUnlockCount: number;
  highestLevel: number;
  avgMessagesToUnlock: number;
  latestUnlockAt: string | null;
};

export type DashboardBlogListItem = {
  title: string;
  description: string;
  publishedAt: string;
  slug: string;
  draft: boolean;
  tags: string[];
};

export type DashboardBlogDetail = DashboardBlogListItem & {
  body: string;
};

export type DashboardBlogPayload = {
  page: number;
  pageSize: number;
  total: number;
  pages: number;
  hasPrev: boolean;
  hasNext: boolean;
  items: DashboardBlogListItem[];
  selectedPost: DashboardBlogDetail | null;
  requestedPost: string | null;
};

type TechnologyCategoryId =
  | "frontend"
  | "backendDb"
  | "infrastructure"
  | "tools";

type TechnologyItemId =
  | "nextjs"
  | "tailwindCss"
  | "postgresql"
  | "sqlite"
  | "vercel"
  | "cloudflare"
  | "github"
  | "resend"
  | "vsCode"
  | "kiloCode";

type DashboardCopy = {
  badge: string;
  title: string;
  description: string;
  navigationLabel: string;
  navOverview: string;
  navProjects: string;
  navTech: string;
  navBlog: string;
  navChangelog: string;
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
  directUnlockCountLabel: string;
  avgMessagesToUnlockLabel: string;
  latestUnlockLabel: string;
  latestUnlockNever: string;
  sourceOffline: string;
  contactTitle: string;
  contactDescription: string;
  contactAvailabilityEyebrow: string;
  contactAvailabilityStatus: string;
  contactConnectLabel: string;
  contactGithubLabel: string;
  contactLinkedinLabel: string;
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
  techCategories: Record<TechnologyCategoryId, string>;
  techItems: Record<
    TechnologyItemId,
    {
      name: string;
      description: string;
    }
  >;
  blogPlaceholder: string;
  blogPreviousLabel: string;
  blogNextLabel: string;
  blogBackLabel: string;
  blogPageLabel: string;
  blogNoPostsLabel: string;
  blogMissingPostLabel: string;
  blogShareLabel: string;
  blogShareCopiedLabel: string;
  blogShareErrorLabel: string;
  blogDraftBadgeLabel: string;
  settingsLanguageTitle: string;
  settingsLanguageDescription: string;
  settingsCookieTitle: string;
  settingsCookieDescription: string;
  settingsThemeTitle: string;
  settingsThemeDescription: string;
  privacyPolicyLink: string;
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
  changelogTitle: string;
  changelogLead: string;
  changelogEmptyLabel: string;
  changelogTypeAdded: string;
  changelogTypeChanged: string;
  changelogTypeFixed: string;
  changelogTypeRemoved: string;
  navAiPulse: string;
  aiPulseTitle: string;
  aiPulseDescription: string;
  aiPulseTrendsTitle: string;
  aiPulseStocksTitle: string;
  aiPulseTickerLabel: string;
  aiPulseNoTrendsLabel: string;
  aiPulseLastUpdatedLabel: string;
  aiPulseSourceLabel: string;
  aiPulseLoadingLabel: string;
};

type UnlockedDashboardProps = {
  locale: string;
  copy: DashboardCopy;
  stats: DashboardStats | null;
  blog: DashboardBlogPayload;
  changelog: ChangelogRelease[];
  aiPulse: AiPulseData | null;
  initialView?: DashboardView;
};

type DashboardView =
  | "overview"
  | "projects"
  | "tech"
  | "blog"
  | "changelog"
  | "settings"
  | "ai-pulse";
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

function updateDashboardQueryParams(
  searchParams: { toString: () => string },
  updates: Record<string, string | null>,
) {
  const next = new URLSearchParams(searchParams.toString());

  for (const [key, value] of Object.entries(updates)) {
    if (!value) {
      next.delete(key);
      continue;
    }

    next.set(key, value);
  }

  return next;
}

function renderMarkdownToSafeHtml(source: string) {
  const rendered = marked.parse(source, {
    breaks: true,
    gfm: true,
  });

  const html = typeof rendered === "string" ? rendered : "";

  return sanitizeHtml(html, {
    allowedTags: sanitizeHtml.defaults.allowedTags.concat(["img"]),
    allowedAttributes: {
      ...sanitizeHtml.defaults.allowedAttributes,
      a: ["href", "name", "target", "rel"],
      img: ["src", "alt", "title", "width", "height", "style", "align"],
    },
    allowedStyles: {
      img: {
        width: [/^\d+(px|%)$/],
        height: [/^\d+(px|%)$/],
        margin: [/^[\d\sa-z%.-]+$/i],
        display: [/^(inline|block)$/],
        float: [/^(left|right|none)$/],
        "max-width": [/^\d+(px|%)$/],
      },
    },
    transformTags: {
      a: sanitizeHtml.simpleTransform("a", {
        target: "_blank",
        rel: "noreferrer noopener",
      }),
    },
  });
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
    { id: "changelog", label: copy.navChangelog },
    { id: "ai-pulse", label: copy.navAiPulse },
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

function OverviewView({
  locale,
  copy,
  stats,
}: {
  locale: string;
  copy: DashboardCopy;
  stats: DashboardStats | null;
}) {
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
            <h2 className="text-xl font-semibold text-foreground">
              {copy.aboutTitle}
            </h2>
            <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
              {copy.aboutBody}
            </p>

            <div className="mt-5">
              <p className="text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
                {copy.toolkitTitle}
              </p>
              <div className="mt-3 flex flex-wrap gap-2">
                <span className="rounded-full border border-border bg-muted px-3 py-1 text-xs text-foreground">
                  {copy.codingModelsLabel}
                </span>
                <span className="rounded-full border border-border bg-muted px-3 py-1 text-xs text-foreground">
                  {copy.chatbotModelLabel}
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="space-y-3">
        <h3 className="text-lg font-semibold text-foreground">
          {copy.analyticsTitle}
        </h3>

        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">
              {copy.totalAttemptsLabel}
            </p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {stats?.totalAttempts ?? 0}
            </p>
          </article>

          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">
              {copy.unlockedCountLabel}
            </p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {stats?.unlockedCount ?? 0}
            </p>
          </article>

          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">
              {copy.directUnlockCountLabel}
            </p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {stats?.directUnlockCount ?? 0}
            </p>
          </article>

          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">
              {copy.avgMessagesToUnlockLabel}
            </p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {Number((stats?.avgMessagesToUnlock ?? 0).toFixed(1))}
            </p>
          </article>

          <article className="rounded-xl border border-border bg-card p-4">
            <p className="text-xs text-muted-foreground">
              {copy.latestUnlockLabel}
            </p>
            <p className="mt-2 text-sm font-medium text-foreground">
              {latestUnlockText}
            </p>
          </article>
        </div>

        {!stats ? (
          <p className="text-xs text-destructive">{copy.sourceOffline}</p>
        ) : null}
      </section>

      <section className="rounded-xl border border-border bg-card p-6">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,320px)] lg:items-start">
          <div>
            <h3 className="text-lg font-semibold text-foreground">
              {copy.contactTitle}
            </h3>
            <p className="mt-2 text-sm text-muted-foreground">
              {copy.contactDescription}
            </p>
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
          </div>

          <ContactCompanionPanel copy={copy} />
        </div>
      </section>
    </div>
  );
}

function GithubBrandIcon({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
      className={className}
    >
      <path d="M12 .5C5.65.5.5 5.65.5 12a11.5 11.5 0 0 0 7.86 10.94c.58.11.79-.25.79-.56v-2c-3.2.7-3.88-1.37-3.88-1.37-.52-1.32-1.28-1.67-1.28-1.67-1.05-.72.08-.7.08-.7 1.16.08 1.77 1.2 1.77 1.2 1.03 1.77 2.7 1.26 3.36.96.1-.75.4-1.26.73-1.55-2.56-.29-5.25-1.28-5.25-5.7 0-1.26.45-2.29 1.19-3.1-.12-.29-.51-1.47.11-3.06 0 0 .97-.31 3.18 1.18a11 11 0 0 1 5.78 0c2.2-1.49 3.17-1.18 3.17-1.18.63 1.59.24 2.77.12 3.06.74.81 1.18 1.84 1.18 3.1 0 4.43-2.7 5.4-5.27 5.69.41.36.77 1.06.77 2.14v3.17c0 .31.21.68.8.56A11.5 11.5 0 0 0 23.5 12C23.5 5.65 18.35.5 12 .5Z" />
    </svg>
  );
}

function LinkedinBrandIcon({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
      className={className}
    >
      <path d="M20.45 20.45h-3.56v-5.57c0-1.33-.03-3.04-1.85-3.04-1.85 0-2.14 1.44-2.14 2.94v5.67H9.34V9h3.41v1.56h.05c.48-.9 1.64-1.85 3.38-1.85 3.61 0 4.27 2.38 4.27 5.47v6.27ZM5.34 7.43a2.06 2.06 0 1 1 0-4.12 2.06 2.06 0 0 1 0 4.12ZM7.12 20.45H3.56V9h3.56v11.45ZM22.22 0H1.77C.79 0 0 .77 0 1.72v20.56C0 23.23.79 24 1.77 24h20.45c.98 0 1.78-.77 1.78-1.72V1.72C24 .77 23.2 0 22.22 0Z" />
    </svg>
  );
}

function ContactCompanionPanel({ copy }: { copy: DashboardCopy }) {
  return (
    <aside className="rounded-lg border border-border bg-background/40 p-5">
      <p className="text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
        {copy.contactAvailabilityEyebrow}
      </p>
      <p className="mt-2 text-sm leading-relaxed text-foreground">
        {copy.contactAvailabilityStatus}
      </p>

      <div className="my-5 h-px bg-border" aria-hidden="true" />

      <p className="text-xs font-semibold tracking-[0.14em] text-muted-foreground uppercase">
        {copy.contactConnectLabel}
      </p>
      <ul className="mt-3 space-y-2">
        <li>
          <Link
            href="https://github.com/karot00/karotammela"
            target="_blank"
            rel="noreferrer noopener"
            className="group flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            <GithubBrandIcon className="size-4 text-muted-foreground transition-colors group-hover:text-foreground" />
            <span>{copy.contactGithubLabel}</span>
          </Link>
        </li>
        <li>
          <Link
            href="https://www.linkedin.com/in/karo-tammela/"
            target="_blank"
            rel="noreferrer noopener"
            className="group flex items-center gap-2 text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            <LinkedinBrandIcon className="size-4 text-muted-foreground transition-colors group-hover:text-foreground" />
            <span>{copy.contactLinkedinLabel}</span>
          </Link>
        </li>
      </ul>
    </aside>
  );
}

function ProjectsView({ copy }: { copy: DashboardCopy }) {
  const projects = [
    {
      title: copy.projectOneTitle,
      description: copy.projectOneDescription,
      href: "https://levifinland.fi",
    },
    {
      title: copy.projectTwoTitle,
      description: copy.projectTwoDescription,
      href: "https://greenfee.levifinland.fi",
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
      <h2 className="text-xl font-semibold text-foreground">
        {copy.projectsTitle}
      </h2>

      <div className="grid gap-4 sm:grid-cols-2">
        {projects.map((project) => {
          const cardContent = (
            <>
              <h3 className="font-semibold text-foreground">{project.title}</h3>
              <p className="mt-2 text-sm text-muted-foreground">
                {project.description}
              </p>

              {project.githubHref ? (
                <Button asChild size="sm" variant="outline" className="mt-4">
                  <Link
                    href={project.githubHref}
                    target="_blank"
                    rel="noreferrer"
                  >
                    {project.githubLabel}
                  </Link>
                </Button>
              ) : null}
            </>
          );

          if (project.href) {
            return (
              <Link
                key={project.title}
                href={project.href}
                target="_blank"
                rel="noreferrer noopener"
                className="group block rounded-xl border border-border bg-card p-5 transition-all hover:-translate-y-0.5 hover:border-foreground/30 hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {cardContent}
              </Link>
            );
          }

          return (
            <article
              key={project.title}
              className="rounded-xl border border-border bg-card p-5"
            >
              {cardContent}
            </article>
          );
        })}
      </div>
    </section>
  );
}

function TechView({ copy }: { copy: DashboardCopy }) {
  type TechnologyIconName =
    | "Layout"
    | "Palette"
    | "Database"
    | "HardDrive"
    | "Cloud"
    | "Shield"
    | "GitFork"
    | "Mail"
    | "Code2"
    | "Cpu";

  type Technology = {
    name: string;
    category: TechnologyCategoryId;
    description: string;
    icon: TechnologyIconName;
  };

  const technologyIconMap: Record<TechnologyIconName, LucideIcon> = {
    Layout,
    Palette,
    Database,
    HardDrive,
    Cloud,
    Shield,
    GitFork,
    Mail,
    Code2,
    Cpu,
  };

  const categoryOrder: TechnologyCategoryId[] = [
    "frontend",
    "backendDb",
    "infrastructure",
    "tools",
  ];

  const technologyConfig: Array<{
    id: TechnologyItemId;
    category: TechnologyCategoryId;
    icon: TechnologyIconName;
  }> = [
    { id: "nextjs", category: "frontend", icon: "Layout" },
    { id: "tailwindCss", category: "frontend", icon: "Palette" },
    { id: "postgresql", category: "backendDb", icon: "Database" },
    { id: "sqlite", category: "backendDb", icon: "HardDrive" },
    { id: "vercel", category: "infrastructure", icon: "Cloud" },
    { id: "cloudflare", category: "infrastructure", icon: "Shield" },
    { id: "github", category: "infrastructure", icon: "GitFork" },
    { id: "resend", category: "tools", icon: "Mail" },
    { id: "vsCode", category: "tools", icon: "Code2" },
    { id: "kiloCode", category: "tools", icon: "Cpu" },
  ];

  const technologies: Technology[] = technologyConfig.map((item) => ({
    category: item.category,
    icon: item.icon,
    name: copy.techItems[item.id].name,
    description: copy.techItems[item.id].description,
  }));

  const technologiesByCategory = categoryOrder.map((category) => ({
    category,
    label: copy.techCategories[category],
    items: technologies.filter(
      (technology) => technology.category === category,
    ),
  }));

  return (
    <section className="mx-auto w-full max-w-6xl space-y-10">
      <h2 className="text-xl font-semibold text-foreground">
        {copy.techTitle}
      </h2>

      {technologiesByCategory.map((group) => (
        <section key={group.category} className="space-y-6">
          <h3 className="flex items-center gap-2 text-xs font-semibold tracking-[0.16em] text-muted-foreground uppercase after:h-px after:flex-1 after:bg-border">
            {group.label}
          </h3>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {group.items.map((technology) => {
              const Icon = technologyIconMap[technology.icon];

              return (
                <article
                  key={technology.name}
                  className="group flex items-start gap-4 rounded-xl border border-border bg-card p-5 text-foreground transition-colors hover:bg-muted/40"
                >
                  <div className="rounded-lg bg-muted p-2.5 text-muted-foreground transition-colors group-hover:text-foreground">
                    <Icon className="size-5" aria-hidden="true" />
                  </div>

                  <div>
                    <h4 className="text-base font-medium text-foreground">
                      {technology.name}
                    </h4>
                    <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                      {technology.description}
                    </p>
                  </div>
                </article>
              );
            })}
          </div>
        </section>
      ))}
    </section>
  );
}

function BlogView({
  copy,
  blog,
  locale,
  onPageChange,
  onOpenPost,
}: {
  copy: DashboardCopy;
  blog: DashboardBlogPayload;
  locale: string;
  onPageChange: (page: number) => void;
  onOpenPost: (slug: string) => void;
}) {
  const localeCode = locale === "fi" ? "fi-FI" : "en-US";

  if (blog.total === 0) {
    return (
      <section className="flex min-h-[320px] items-center justify-center rounded-xl border border-border bg-card p-6">
        <p className="text-base text-muted-foreground">
          {copy.blogNoPostsLabel}
        </p>
      </section>
    );
  }

  return (
    <section className="grid gap-6 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
      <div className="rounded-xl border border-border bg-card p-4">
        <div className="space-y-3">
          {blog.items.map((post) => {
            const publishedAt = new Intl.DateTimeFormat(localeCode, {
              dateStyle: "medium",
            }).format(new Date(post.publishedAt));

            return (
              <article
                key={post.slug}
                className="rounded-lg border border-border bg-background p-3"
              >
                <button
                  type="button"
                  onClick={() => onOpenPost(post.slug)}
                  className="w-full text-left"
                >
                  <p className="text-sm font-semibold text-foreground">
                    {post.title}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {publishedAt}
                  </p>
                  <p className="mt-2 text-sm text-muted-foreground">
                    {post.description}
                  </p>
                  {post.draft ? (
                    <span className="mt-2 inline-block rounded border border-border bg-muted px-2 py-0.5 text-[11px] text-foreground">
                      {copy.blogDraftBadgeLabel}
                    </span>
                  ) : null}
                </button>
              </article>
            );
          })}
        </div>

        <div className="mt-4 flex items-center justify-between gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!blog.hasPrev}
            onClick={() => onPageChange(blog.page - 1)}
          >
            {copy.blogPreviousLabel}
          </Button>
          <p className="text-xs text-muted-foreground">
            {copy.blogPageLabel} {blog.page} / {blog.pages}
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={!blog.hasNext}
            onClick={() => onPageChange(blog.page + 1)}
          >
            {copy.blogNextLabel}
          </Button>
        </div>
      </div>

      <article className="rounded-xl border border-border bg-card p-6">
        {blog.selectedPost ? (
          <div className="mx-auto w-full max-w-3xl">
            <div className="flex items-center gap-3">
              <h2 className="text-xl font-semibold text-foreground">
                {blog.selectedPost.title}
              </h2>
            </div>
            <ShareButton
              url={`/${locale}/blog/${blog.selectedPost.slug}`}
              title={blog.selectedPost.title}
              text={blog.selectedPost.description}
              label={copy.blogShareLabel}
              copiedLabel={copy.blogShareCopiedLabel}
              errorLabel={copy.blogShareErrorLabel}
              className="mt-3"
            />
            <p className="mt-2 text-sm text-muted-foreground">
              {new Intl.DateTimeFormat(localeCode, {
                dateStyle: "medium",
              }).format(new Date(blog.selectedPost.publishedAt))}
            </p>
            <div
              className="blog-prose mt-4"
              dangerouslySetInnerHTML={{
                __html: renderMarkdownToSafeHtml(blog.selectedPost.body),
              }}
            />
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            {blog.requestedPost
              ? copy.blogMissingPostLabel
              : copy.blogPlaceholder}
          </p>
        )}
      </article>
    </section>
  );
}

function ChangelogView({
  copy,
  locale,
  changelog,
}: {
  copy: DashboardCopy;
  locale: string;
  changelog: ChangelogRelease[];
}) {
  const localeCode = locale === "fi" ? "fi-FI" : "en-US";

  const typeLabels: Record<ChangelogChangeType, string> = {
    added: copy.changelogTypeAdded,
    changed: copy.changelogTypeChanged,
    fixed: copy.changelogTypeFixed,
    removed: copy.changelogTypeRemoved,
  };

  const typeStyles: Record<ChangelogChangeType, string> = {
    added: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    changed: "bg-sky-500/10 text-sky-400 border-sky-500/20",
    fixed: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    removed: "bg-rose-500/10 text-rose-400 border-rose-500/20",
  };

  const formatReleaseDate = (value: string) => {
    const parsed = new Date(value);

    if (Number.isNaN(parsed.getTime())) {
      return value;
    }

    return new Intl.DateTimeFormat(localeCode, {
      dateStyle: "medium",
    }).format(parsed);
  };

  return (
    <section className="mx-auto w-full max-w-4xl space-y-6">
      <div>
        <h2 className="text-xl font-semibold text-foreground">
          {copy.changelogTitle}
        </h2>
        <p className="mt-2 text-sm text-muted-foreground">
          {copy.changelogLead}
        </p>
      </div>

      {changelog.length === 0 ? (
        <p className="rounded-xl border border-border bg-card p-6 text-sm text-muted-foreground">
          {copy.changelogEmptyLabel}
        </p>
      ) : (
        <div className="space-y-4">
          {changelog.map((release) => (
            <article
              key={release.version}
              className="rounded-xl border border-border bg-card p-5"
            >
              <div className="grid gap-4 md:grid-cols-[140px_1fr] md:items-start">
                <div>
                  <span className="inline-flex items-center rounded-full border border-border bg-muted px-2.5 py-0.5 text-xs font-semibold text-foreground">
                    {release.version}
                  </span>
                  <p className="mt-2 text-xs text-muted-foreground">
                    {formatReleaseDate(release.date)}
                  </p>
                </div>

                <div>
                  {release.title ? (
                    <h3 className="text-base font-medium text-foreground">
                      {release.title}
                    </h3>
                  ) : null}

                  <ul className="mt-3 space-y-2">
                    {release.changes.map((change, index) => (
                      <li
                        key={`${release.version}-${index}`}
                        className="flex items-start gap-2 text-sm leading-relaxed text-muted-foreground"
                      >
                        <span
                          className={`mt-0.5 inline-flex shrink-0 items-center rounded-full border px-2 py-0.5 text-[11px] font-semibold tracking-wide uppercase ${typeStyles[change.type]}`}
                        >
                          {typeLabels[change.type]}
                        </span>
                        <span>{change.text}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

function SettingsView({
  locale,
  copy,
  themeMode,
  onThemeChange,
}: {
  locale: string;
  copy: DashboardCopy;
  themeMode: ThemeMode;
  onThemeChange: (mode: ThemeMode) => void;
}) {
  return (
    <section className="space-y-6">
      <article className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold text-foreground">
          {copy.settingsLanguageTitle}
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {copy.settingsLanguageDescription}
        </p>
        <div className="mt-4">
          <LocaleSwitcher />
        </div>
      </article>

      <article className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold text-foreground">
          {copy.settingsThemeTitle}
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {copy.settingsThemeDescription}
        </p>

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

      <article className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold text-foreground">
          {copy.settingsCookieTitle}
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          {copy.settingsCookieDescription}
        </p>
        <div className="mt-4">
          <CookieConsentSettingsTrigger />
        </div>
      </article>

      <article className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-lg font-semibold text-foreground">
          {copy.privacyPolicyLink}
        </h2>
        <div className="mt-4">
          <Link
            href={`/${locale}/privacy?returnTo=${encodeURIComponent(
              `/${locale}/dashboard?view=settings`,
            )}`}
            className="text-sm text-primary underline-offset-4 hover:underline"
          >
            {copy.privacyPolicyLink}
          </Link>
        </div>
      </article>
    </section>
  );
}

export function UnlockedDashboard(props: UnlockedDashboardProps) {
  const {
    locale,
    copy,
    stats,
    blog,
    changelog,
    aiPulse,
    initialView = "overview",
  } = props;
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [localView, setLocalView] = useState<DashboardView>(initialView);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [themeMode, setThemeMode] = useState<ThemeMode>(() => {
    if (typeof window === "undefined") {
      return "dark";
    }

    return window.localStorage.getItem("karot-theme") === "light"
      ? "light"
      : "dark";
  });

  useEffect(() => {
    document.documentElement.classList.toggle("light", themeMode === "light");
  }, [themeMode]);

  const onThemeChange = (mode: ThemeMode) => {
    setThemeMode(mode);
    window.localStorage.setItem("karot-theme", mode);
    document.documentElement.classList.toggle("light", mode === "light");
  };

  const replaceQuery = (updates: Record<string, string | null>) => {
    const next = updateDashboardQueryParams(searchParams, updates);
    const queryString = next.toString();

    router.replace(queryString ? `${pathname}?${queryString}` : pathname, {
      scroll: false,
    });
  };

  const onSelectView = (view: DashboardView) => {
    setLocalView(view);

    if (view === "blog") {
      replaceQuery({ view: "blog" });
      return;
    }

    if (view === "changelog") {
      replaceQuery({ view: "changelog", page: null, post: null });
      return;
    }

    if (view === "ai-pulse") {
      replaceQuery({ view: "ai-pulse", page: null, post: null });
      return;
    }

    replaceQuery({ view: null, page: null, post: null });
  };

  const onBlogPageChange = (page: number) => {
    replaceQuery({ view: "blog", page: String(page), post: null });
  };

  const onBlogOpenPost = (slug: string) => {
    replaceQuery({ view: "blog", post: slug });
  };

  const viewParam = searchParams.get("view");
  const activeView: DashboardView =
    viewParam === "blog" ||
    viewParam === "changelog" ||
    viewParam === "ai-pulse"
      ? (viewParam as DashboardView)
      : localView;

  let view = <OverviewView locale={locale} copy={copy} stats={stats} />;
  if (activeView === "projects") view = <ProjectsView copy={copy} />;
  if (activeView === "tech") view = <TechView copy={copy} />;
  if (activeView === "blog") {
    view = (
      <BlogView
        copy={copy}
        blog={blog}
        locale={locale}
        onPageChange={onBlogPageChange}
        onOpenPost={onBlogOpenPost}
      />
    );
  }
  if (activeView === "changelog") {
    view = <ChangelogView copy={copy} locale={locale} changelog={changelog} />;
  }
  if (activeView === "settings")
    view = (
      <SettingsView
        locale={locale}
        copy={copy}
        themeMode={themeMode}
        onThemeChange={onThemeChange}
      />
    );
  if (activeView === "ai-pulse" && aiPulse) {
    view = (
      <AiPulseShell
        data={aiPulse}
        locale={locale}
        copy={{
          aiPulseTitle: copy.aiPulseTitle,
          aiPulseDescription: copy.aiPulseDescription,
          aiPulseTrendsTitle: copy.aiPulseTrendsTitle,
          aiPulseStocksTitle: copy.aiPulseStocksTitle,
          aiPulseTickerLabel: copy.aiPulseTickerLabel,
          aiPulseNoTrendsLabel: copy.aiPulseNoTrendsLabel,
          aiPulseLastUpdatedLabel: copy.aiPulseLastUpdatedLabel,
          aiPulseSourceLabel: copy.aiPulseSourceLabel,
          aiPulseLoadingLabel: copy.aiPulseLoadingLabel,
        }}
      />
    );
  }

  return (
    <main className="flex min-h-screen bg-background">
      <aside className="hidden w-72 border-r border-sidebar-border bg-sidebar px-4 py-6 md:block">
        <p className="px-3 text-xs font-semibold tracking-[0.16em] text-sidebar-foreground/70 uppercase">
          {copy.navigationLabel}
        </p>
        <div className="mt-4">
          <SidebarNav
            activeView={activeView}
            copy={copy}
            onSelect={onSelectView}
          />
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 border-b border-border bg-background/95 px-4 py-3 backdrop-blur supports-[backdrop-filter]:bg-background/85 sm:px-6">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2 md:hidden">
              <Button
                type="button"
                variant="outline"
                size="icon"
                onClick={() => setMobileMenuOpen((open) => !open)}
              >
                <Menu className="size-4" />
              </Button>
              <span className="text-sm font-medium text-foreground">
                {copy.menuLabel}
              </span>
            </div>

            <div className="min-w-0">
              <p className="text-xs tracking-[0.12em] text-muted-foreground uppercase">
                {copy.badge}
              </p>
              <h1 className="truncate text-lg font-semibold text-foreground sm:text-xl">
                {copy.title}
              </h1>
              {copy.description ? (
                <p className="hidden text-xs text-muted-foreground sm:block">
                  {copy.description}
                </p>
              ) : null}
            </div>

            <Link
              href={`/${locale}`}
              className="text-xs font-medium text-primary hover:text-primary/80"
            >
              {copy.homeLinkLabel}
            </Link>
          </div>

          {mobileMenuOpen ? (
            <div className="mt-3 rounded-xl border border-border bg-card p-3 md:hidden">
              <SidebarNav
                activeView={activeView}
                copy={copy}
                onSelect={onSelectView}
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
