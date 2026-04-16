import { NextResponse } from "next/server";
import { z } from "zod";

import { getCachedDashboardStats } from "@/lib/db/stats-cache";

const querySchema = z.object({
  locale: z.enum(["en", "fi"]).default("fi"),
});

const responseSchema = z.object({
  totalAttempts: z.number().int().nonnegative(),
  unlockedCount: z.number().int().nonnegative(),
  highestLevel: z.number().int().min(0).max(100),
  avgMessagesToUnlock: z.number().nonnegative(),
  latestUnlockAt: z.union([z.date(), z.null()]),
});

export const runtime = "nodejs";

export async function GET(request: Request) {
  if (!process.env.TURSO_DATABASE_URL) {
    console.warn("[api/stats] TURSO_DATABASE_URL is missing.");
    return NextResponse.json(
      { error: "Stats database is not configured." },
      { status: 503 },
    );
  }

  const url = new URL(request.url);
  const parsed = querySchema.safeParse({
    locale: url.searchParams.get("locale") ?? "fi",
  });

  if (!parsed.success) {
    return NextResponse.json({ error: "Invalid locale." }, { status: 400 });
  }

  try {
    const stats = await getCachedDashboardStats();
    const payload = responseSchema.parse(stats);

    return NextResponse.json(payload);
  } catch (error) {
    console.error("[api/stats] Failed to read stats", error);
    return NextResponse.json(
      { error: "Failed to read stats." },
      { status: 500 },
    );
  }
}
