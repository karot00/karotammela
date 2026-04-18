CREATE TABLE `ai_stocks` (
	`id` text PRIMARY KEY NOT NULL,
	`date` text NOT NULL,
	`ticker` text NOT NULL,
	`company_name` text NOT NULL,
	`open` real NOT NULL,
	`high` real NOT NULL,
	`low` real NOT NULL,
	`close` real NOT NULL,
	`volume` integer,
	`created_at` integer DEFAULT (unixepoch() * 1000) NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `ai_stocks_ticker_date_idx` ON `ai_stocks` (`ticker`,`date`);--> statement-breakpoint
CREATE INDEX `ai_stocks_ticker_date_desc_idx` ON `ai_stocks` (`ticker`,`date`);--> statement-breakpoint
CREATE TABLE `ai_trends` (
	`id` text PRIMARY KEY NOT NULL,
	`date` text NOT NULL,
	`title` text NOT NULL,
	`summary` text NOT NULL,
	`url` text NOT NULL,
	`source` text,
	`created_at` integer DEFAULT (unixepoch() * 1000) NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `ai_trends_date_title_idx` ON `ai_trends` (`date`,`title`);--> statement-breakpoint
CREATE INDEX `ai_trends_date_idx` ON `ai_trends` (`date`);