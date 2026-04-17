import enChangelog from "@/content/changelog/en.json";
import fiChangelog from "@/content/changelog/fi.json";

export type ChangelogChangeType = "added" | "changed" | "fixed" | "removed";

export type ChangelogChange = {
  type: ChangelogChangeType;
  text: string;
};

export type ChangelogRelease = {
  version: string;
  date: string;
  title?: string;
  changes: ChangelogChange[];
};

type ChangelogFile = {
  releases: ChangelogRelease[];
};

const enReleases = (enChangelog as ChangelogFile).releases;
const fiReleases = (fiChangelog as ChangelogFile).releases;

export function getChangelog(locale: "fi" | "en"): ChangelogRelease[] {
  return locale === "fi" ? fiReleases : enReleases;
}
