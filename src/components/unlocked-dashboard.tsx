import { ContactForm } from "@/components/contact-form";

type DashboardStats = {
  totalAttempts: number;
  unlockedCount: number;
  highestLevel: number;
  latestUnlockAt: Date | null;
};

type DashboardCopy = {
  badge: string;
  title: string;
  description: string;
  livePulseTitle: string;
  totalAttemptsLabel: string;
  unlockedCountLabel: string;
  highestLevelLabel: string;
  latestUnlockLabel: string;
  latestUnlockNever: string;
  sourceOffline: string;
  labTitle: string;
  labProjectOneTitle: string;
  labProjectOneDescription: string;
  labProjectTwoTitle: string;
  labProjectTwoDescription: string;
  labProjectThreeTitle: string;
  labProjectThreeDescription: string;
  dossierTitle: string;
  dossierBody: string;
  dossierLocationLabel: string;
  dossierLocation: string;
  dossierContactLabel: string;
  dossierContact: string;
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

function formatDate(value: Date, locale: string) {
  const localeCode = locale === "fi" ? "fi-FI" : "en-US";

  return new Intl.DateTimeFormat(localeCode, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(value);
}

export function UnlockedDashboard({ locale, copy, stats }: UnlockedDashboardProps) {
  const latestUnlockText = stats?.latestUnlockAt
    ? formatDate(stats.latestUnlockAt, locale)
    : copy.latestUnlockNever;

  return (
    <main className="flex flex-1 px-6 py-10 sm:py-16">
      <section className="mx-auto flex w-full max-w-6xl flex-col gap-8">
        <header className="rounded-3xl border border-border/70 bg-card/50 p-7 backdrop-blur-md sm:p-10">
          <p className="text-xs font-semibold tracking-[0.22em] text-accent uppercase">{copy.badge}</p>
          <h1 className="mt-4 text-3xl font-semibold text-foreground sm:text-4xl">{copy.title}</h1>
          <p className="mt-4 max-w-3xl text-muted-foreground">{copy.description}</p>
        </header>

        <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
          <article className="rounded-3xl border border-border/70 bg-popover/60 p-6 backdrop-blur-md">
            <h2 className="text-sm font-semibold tracking-[0.18em] text-primary uppercase">{copy.livePulseTitle}</h2>

            <div className="mt-5 space-y-3 text-sm">
              <div className="flex items-center justify-between rounded-lg bg-background/70 px-3 py-2">
                <span className="text-muted-foreground">{copy.totalAttemptsLabel}</span>
                <strong>{stats?.totalAttempts ?? 0}</strong>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-background/70 px-3 py-2">
                <span className="text-muted-foreground">{copy.unlockedCountLabel}</span>
                <strong>{stats?.unlockedCount ?? 0}</strong>
              </div>
              <div className="flex items-center justify-between rounded-lg bg-background/70 px-3 py-2">
                <span className="text-muted-foreground">{copy.highestLevelLabel}</span>
                <strong>{stats?.highestLevel ?? 0}%</strong>
              </div>
              <div className="rounded-lg bg-background/70 px-3 py-2">
                <p className="text-muted-foreground">{copy.latestUnlockLabel}</p>
                <p className="mt-1 font-medium">{latestUnlockText}</p>
              </div>
            </div>

            {!stats ? <p className="mt-4 text-xs text-destructive">{copy.sourceOffline}</p> : null}
          </article>

          <article className="rounded-3xl border border-border/70 bg-popover/60 p-6 backdrop-blur-md">
            <h2 className="text-sm font-semibold tracking-[0.18em] text-primary uppercase">{copy.labTitle}</h2>

            <div className="mt-5 space-y-3">
              <div className="rounded-xl border border-border/70 bg-background/60 p-4">
                <p className="font-semibold">{copy.labProjectOneTitle}</p>
                <p className="mt-1 text-sm text-muted-foreground">{copy.labProjectOneDescription}</p>
              </div>

              <div className="rounded-xl border border-border/70 bg-background/60 p-4">
                <p className="font-semibold">{copy.labProjectTwoTitle}</p>
                <p className="mt-1 text-sm text-muted-foreground">{copy.labProjectTwoDescription}</p>
              </div>

              <div className="rounded-xl border border-border/70 bg-background/60 p-4">
                <p className="font-semibold">{copy.labProjectThreeTitle}</p>
                <p className="mt-1 text-sm text-muted-foreground">{copy.labProjectThreeDescription}</p>
              </div>
            </div>
          </article>

          <article className="rounded-3xl border border-border/70 bg-popover/60 p-6 backdrop-blur-md">
            <h2 className="text-sm font-semibold tracking-[0.18em] text-primary uppercase">{copy.dossierTitle}</h2>

            <p className="mt-5 text-sm leading-relaxed text-muted-foreground">{copy.dossierBody}</p>

            <div className="mt-6 space-y-2 text-sm">
              <p>
                <span className="text-muted-foreground">{copy.dossierLocationLabel}: </span>
                <span>{copy.dossierLocation}</span>
              </p>
              <p>
                <span className="text-muted-foreground">{copy.dossierContactLabel}: </span>
                <a className="text-primary hover:text-primary/80" href={`mailto:${copy.dossierContact}`}>
                  {copy.dossierContact}
                </a>
              </p>
            </div>

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
          </article>
        </div>
      </section>
    </main>
  );
}
