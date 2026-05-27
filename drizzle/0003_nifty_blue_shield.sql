CREATE TABLE `ai_repos` (
	`id` text PRIMARY KEY NOT NULL,
	`date` text NOT NULL,
	`repo_full_name` text NOT NULL,
	`url` text NOT NULL,
	`description` text,
	`description_fi` text,
	`language` text,
	`stars` integer NOT NULL,
	`stars_today` integer NOT NULL,
	`source` text DEFAULT 'github-trending' NOT NULL,
	`created_at` integer DEFAULT (unixepoch() * 1000) NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `ai_repos_date_repo_idx` ON `ai_repos` (`date`,`repo_full_name`);--> statement-breakpoint
CREATE INDEX `ai_repos_date_idx` ON `ai_repos` (`date`);