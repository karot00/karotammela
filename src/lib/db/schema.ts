import { sql } from "drizzle-orm";
import { check, integer, sqliteTable, text } from "drizzle-orm/sqlite-core";

export const sessions = sqliteTable("sessions", {
  sessionId: text("session_id").primaryKey(),
  locale: text("locale", { enum: ["en", "fi"] }).notNull(),
  createdAt: integer("created_at", { mode: "timestamp_ms" })
    .notNull()
    .default(sql`(unixepoch() * 1000)`),
  updatedAt: integer("updated_at", { mode: "timestamp_ms" })
    .notNull()
    .default(sql`(unixepoch() * 1000)`),
  lastLevel: integer("last_level").notNull().default(0),
  unlocked: integer("unlocked", { mode: "boolean" }).notNull().default(false),
});

export const logs = sqliteTable(
  "logs",
  {
    id: text("id").primaryKey(),
    timestamp: integer("timestamp", { mode: "timestamp_ms" })
      .notNull()
      .default(sql`(unixepoch() * 1000)`),
    sessionId: text("session_id")
      .notNull()
      .references(() => sessions.sessionId, { onDelete: "cascade" }),
    locale: text("locale", { enum: ["en", "fi"] }).notNull(),
    userInput: text("user_input").notNull(),
    assistantOutput: text("assistant_output"),
    levelReached: integer("level_reached").notNull(),
    success: integer("success", { mode: "boolean" }).notNull().default(false),
  },
  (table) => [
    check("logs_level_range", sql`${table.levelReached} >= 0 and ${table.levelReached} <= 100`),
  ],
);

export type Session = typeof sessions.$inferSelect;
export type NewSession = typeof sessions.$inferInsert;
export type Log = typeof logs.$inferSelect;
export type NewLog = typeof logs.$inferInsert;
