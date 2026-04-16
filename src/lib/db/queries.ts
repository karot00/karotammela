import { desc, eq, sql } from "drizzle-orm";

import { getDb } from "@/lib/db/client";
import { insertLogSchema, insertSessionSchema } from "@/lib/db/validation";
import { logs, sessions } from "@/lib/db/schema";

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

export async function getDashboardStats() {
  const db = getDb();

  const [totals] = await db
    .select({
      totalAttempts: sql<number>`count(*)`,
      unlockedCount: sql<number>`sum(case when ${logs.success} = 1 then 1 else 0 end)`,
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
    highestLevel: totals?.highestLevel ?? 0,
    avgMessagesToUnlock: Number(averages?.avgMessagesToUnlock ?? 0),
    latestUnlockAt: latest[0]?.timestamp ?? null,
  };
}
