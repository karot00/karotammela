import { asc, desc, eq, sql } from "drizzle-orm";

import { getDb } from "@/lib/db/client";
import { insertLogSchema, insertSessionSchema } from "@/lib/db/validation";
import { getAccessCode } from "@/lib/ai/sentinel";
import {
  aiStocks,
  aiTrends,
  logs,
  sessions,
  type AiStock,
  type AiTrend,
  type NewAiStock,
  type NewAiTrend,
} from "@/lib/db/schema";

type PersistSentinelTurnInput = {
  sessionId: string;
  locale: "en" | "fi";
  userInput: string;
  assistantOutput: string;
  levelReached: number;
  success: boolean;
};

export async function persistSentinelTurn(input: PersistSentinelTurnInput) {
  const db = getDb();

  const sessionPayload = insertSessionSchema.parse({
    sessionId: input.sessionId,
    locale: input.locale,
    lastLevel: input.levelReached,
    unlocked: input.success,
    updatedAt: new Date(),
  });

  await db
    .insert(sessions)
    .values(sessionPayload)
    .onConflictDoUpdate({
      target: sessions.sessionId,
      set: {
        locale: input.locale,
        lastLevel: input.levelReached,
        unlocked: input.success,
        updatedAt: new Date(),
      },
    });

  const logPayload = insertLogSchema.parse({
    id: crypto.randomUUID(),
    sessionId: input.sessionId,
    locale: input.locale,
    userInput: input.userInput,
    assistantOutput: input.assistantOutput,
    levelReached: input.levelReached,
    success: input.success,
  });

  await db.insert(logs).values(logPayload);
}

export async function getLatestTrends(date?: string): Promise<AiTrend[]> {
  const db = getDb();
  const targetDate = date ?? new Date().toISOString().slice(0, 10);

  const rows = await db
    .select()
    .from(aiTrends)
    .where(eq(aiTrends.date, targetDate))
    .orderBy(asc(aiTrends.createdAt));

  if (rows.length > 0) {
    return rows;
  }

  // Fallback: return the most recent day that has rows
  const [latest] = await db
    .select({ date: aiTrends.date })
    .from(aiTrends)
    .orderBy(desc(aiTrends.date))
    .limit(1);

  if (!latest) {
    return [];
  }

  return db
    .select()
    .from(aiTrends)
    .where(eq(aiTrends.date, latest.date))
    .orderBy(asc(aiTrends.createdAt));
}

export async function getStockHistory(
  ticker: string,
  days = 365,
): Promise<AiStock[]> {
  const db = getDb();

  const cutoff = new Date();
  cutoff.setDate(cutoff.getDate() - days);
  const cutoffStr = cutoff.toISOString().slice(0, 10);

  return db
    .select()
    .from(aiStocks)
    .where(
      sql`${aiStocks.ticker} = ${ticker} AND ${aiStocks.date} >= ${cutoffStr}`,
    )
    .orderBy(asc(aiStocks.date));
}

export async function getAvailableTickers(): Promise<string[]> {
  const db = getDb();

  const rows = await db
    .selectDistinct({ ticker: aiStocks.ticker })
    .from(aiStocks)
    .orderBy(asc(aiStocks.ticker));

  return rows.map((r) => r.ticker);
}

export async function upsertTrends(rows: NewAiTrend[]): Promise<void> {
  if (rows.length === 0) return;
  const db = getDb();

  await db.insert(aiTrends).values(rows).onConflictDoNothing();
}

export async function upsertStocks(rows: NewAiStock[]): Promise<void> {
  if (rows.length === 0) return;
  const db = getDb();

  await db.insert(aiStocks).values(rows).onConflictDoNothing();
}

export async function getDashboardStats() {
  const db = getDb();
  const accessCode = getAccessCode();

  const [totals] = await db
    .select({
      totalAttempts: sql<number>`count(*)`,
      unlockedCount: sql<number>`sum(case when ${logs.success} = 1 then 1 else 0 end)`,
      directUnlockCount: sql<number>`sum(case when ${logs.success} = 1 and upper(trim(${logs.userInput})) = ${accessCode} then 1 else 0 end)`,
      highestLevel: sql<number>`max(${logs.levelReached})`,
    })
    .from(logs);

  const latest = await db
    .select({
      timestamp: logs.timestamp,
      levelReached: logs.levelReached,
      success: logs.success,
    })
    .from(logs)
    .where(eq(logs.success, true))
    .orderBy(desc(logs.timestamp))
    .limit(1);

  const successfulSessionCounts = db
    .select({
      sessionId: logs.sessionId,
      messageCount: sql<number>`count(*)`.as("message_count"),
    })
    .from(logs)
    .groupBy(logs.sessionId)
    .having(sql`sum(case when ${logs.success} = 1 then 1 else 0 end) > 0`)
    .as("successful_session_counts");

  const [averages] = await db
    .select({
      avgMessagesToUnlock: sql<number>`coalesce(avg(${successfulSessionCounts.messageCount}), 0)`,
    })
    .from(successfulSessionCounts);

  return {
    totalAttempts: totals?.totalAttempts ?? 0,
    unlockedCount: totals?.unlockedCount ?? 0,
    directUnlockCount: totals?.directUnlockCount ?? 0,
    highestLevel: totals?.highestLevel ?? 0,
    avgMessagesToUnlock: Number(averages?.avgMessagesToUnlock ?? 0),
    latestUnlockAt: latest[0]?.timestamp ?? null,
  };
}
