CREATE TABLE `logs` (
	`id` text PRIMARY KEY NOT NULL,
	`timestamp` integer DEFAULT (unixepoch() * 1000) NOT NULL,
	`session_id` text NOT NULL,
	`locale` text NOT NULL,
	`user_input` text NOT NULL,
	`assistant_output` text,
	`level_reached` integer NOT NULL,
	`success` integer DEFAULT false NOT NULL,
	FOREIGN KEY (`session_id`) REFERENCES `sessions`(`session_id`) ON UPDATE no action ON DELETE cascade,
	CONSTRAINT "logs_level_range" CHECK("logs"."level_reached" >= 0 and "logs"."level_reached" <= 100)
);
--> statement-breakpoint
CREATE TABLE `sessions` (
	`session_id` text PRIMARY KEY NOT NULL,
	`locale` text NOT NULL,
	`created_at` integer DEFAULT (unixepoch() * 1000) NOT NULL,
	`updated_at` integer DEFAULT (unixepoch() * 1000) NOT NULL,
	`last_level` integer DEFAULT 0 NOT NULL,
	`unlocked` integer DEFAULT false NOT NULL
);
