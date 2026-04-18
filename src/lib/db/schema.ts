import { sql } from "drizzle-orm";
import {
  check,
  index,
  integer,
  real,
  sqliteTable,
  text,
  uniqueIndex,
} from "drizzle-orm/sqlite-core";

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
    check(
      "logs_level_range",
      sql`${table.levelReached} >= 0 and ${table.levelReached} <= 100`,
    ),
  ],
);

export type Session = typeof sessions.$inferSelect;
export type NewSession = typeof sessions.$inferInsert;
export type Log = typeof logs.$inferSelect;
export type NewLog = typeof logs.$inferInsert;

export const aiTrends = sqliteTable(
  "ai_trends",
  {
    id: text("id").primaryKey(),
    date: text("date").notNull(),
    title: text("title").notNull(),
    summary: text("summary").notNull(),
    summaryFi: text("summary_fi"),
    url: text("url").notNull(),
    source: text("source"),
    createdAt: integer("created_at", { mode: "timestamp_ms" })
      .notNull()
      .default(sql`(unixepoch() * 1000)`),
  },
  (table) => [
    uniqueIndex("ai_trends_date_title_idx").on(table.date, table.title),
    index("ai_trends_date_idx").on(table.date),
  ],
);

export const aiStocks = sqliteTable(
  "ai_stocks",
  {
    id: text("id").primaryKey(),
    date: text("date").notNull(),
    ticker: text("ticker").notNull(),
    companyName: text("company_name").notNull(),
    open: real("open").notNull(),
    high: real("high").notNull(),
    low: real("low").notNull(),
    close: real("close").notNull(),
    volume: integer("volume"),
    createdAt: integer("created_at", { mode: "timestamp_ms" })
      .notNull()
      .default(sql`(unixepoch() * 1000)`),
  },
  (table) => [
    uniqueIndex("ai_stocks_ticker_date_idx").on(table.ticker, table.date),
    index("ai_stocks_ticker_date_desc_idx").on(table.ticker, table.date),
  ],
);

export type AiTrend = typeof aiTrends.$inferSelect;
export type NewAiTrend = typeof aiTrends.$inferInsert;
export type AiStock = typeof aiStocks.$inferSelect;
export type NewAiStock = typeof aiStocks.$inferInsert;
