import { google } from "@ai-sdk/google";
import { streamText } from "ai";
import { NextResponse } from "next/server";
import { z } from "zod";

import {
  applyInputAdjustment,
  buildSentinelSystemPrompt,
  extractLevelTag,
  getAccessCode,
  isUnlockTriggered,
  resolveNextLevel,
  stripLevelTag,
} from "@/lib/ai/sentinel";
import { createUnlockCookieValue } from "@/lib/security/unlock-cookie";
import { runMigrationsIfNeeded } from "@/lib/db/migrate";
import { persistSentinelTurn } from "@/lib/db/queries";
import { enforceRateLimit, getClientIp } from "@/lib/security/rate-limit";
import { trackServerEvent } from "@/lib/telemetry/events";

const requestSchema = z.object({
  locale: z.enum(["en", "fi"]),
  sessionId: z.string().min(4).max(128),
  currentLevel: z.number().int().min(0).max(100),
  messages: z
    .array(
      z.object({
        role: z.enum(["user", "assistant"]),
        content: z.string().min(1).max(4000),
      }),
    )
    .min(1)
    .max(30),
});

export const runtime = "nodejs";
export const maxDuration = 30;

function getModelName() {
  return process.env.AI_MODEL || "gemini-3.1-flash-lite-preview";
}

export async function POST(request: Request) {
  const ip = getClientIp(request);
  const ipRate = enforceRateLimit({
    scope: "sentinel-ip",
    key: ip,
    limit: 40,
    windowMs: 60_000,
  });

  if (!ipRate.allowed) {
    trackServerEvent("sentinel.rate_limited", {
      mode: "ip",
      ip,
    });

    return NextResponse.json(
      { error: "Rate limit exceeded. Retry shortly." },
      {
        status: 429,
        headers: {
          "Retry-After": String(Math.ceil(ipRate.retryAfterMs / 1000)),
        },
      },
    );
  }

  const parsed = requestSchema.safeParse(await request.json());
  if (!parsed.success) {
    return NextResponse.json(
      { error: "Invalid request payload." },
      { status: 400 },
    );
  }

  if (!process.env.GOOGLE_GENERATIVE_AI_API_KEY) {
    return NextResponse.json(
      { error: "AI provider key is missing." },
      { status: 500 },
    );
  }

  const { locale, messages, sessionId, currentLevel } = parsed.data;

  const sessionRate = enforceRateLimit({
    scope: "sentinel-session",
    key: sessionId,
    limit: 16,
    windowMs: 5 * 60_000,
  });

  if (!sessionRate.allowed) {
    trackServerEvent("sentinel.rate_limited", {
      mode: "session",
      sessionId,
    });

    return NextResponse.json(
      { error: "Session turn limit reached. Please reset and try again." },
      {
        status: 429,
        headers: {
          "Retry-After": String(Math.ceil(sessionRate.retryAfterMs / 1000)),
        },
      },
    );
  }

  const latestUserInput =
    [...messages].reverse().find((message) => message.role === "user")
      ?.content ?? "";

  try {
    const result = streamText({
      model: google(getModelName()),
      system: buildSentinelSystemPrompt(locale),
      messages,
      temperature: 0.7,
      maxOutputTokens: 300,
    });

    let fullText = "";
    for await (const chunk of result.textStream) {
      fullText += chunk;
    }

    const parsedLevel = extractLevelTag(fullText);
    const resolvedLevel = resolveNextLevel(currentLevel, parsedLevel);
    const level = applyInputAdjustment(resolvedLevel, latestUserInput);
    const cleanText = stripLevelTag(fullText);
    const unlocked = isUnlockTriggered(fullText, level);

    trackServerEvent("sentinel.turn", {
      locale,
      sessionId,
      previousLevel: currentLevel,
      parsedLevel: parsedLevel ?? null,
      finalLevel: level,
      unlocked,
    });

    const response = NextResponse.json({
      message: cleanText,
      level,
      unlocked,
      accessCode: unlocked ? getAccessCode() : null,
    });

    if (unlocked && process.env.UNLOCK_COOKIE_SECRET) {
      const cookieValue = createUnlockCookieValue(
        { sessionId, locale, unlockedAt: Date.now() },
        process.env.UNLOCK_COOKIE_SECRET,
      );

      response.cookies.set({
        name: "karot_unlock",
        value: cookieValue,
        httpOnly: true,
        sameSite: "lax",
        secure: process.env.NODE_ENV === "production",
        path: "/",
        maxAge: 60 * 60 * 24 * 14,
      });
    }

    if (process.env.TURSO_DATABASE_URL) {
      try {
        await runMigrationsIfNeeded();
        await persistSentinelTurn({
          sessionId,
          locale,
          userInput: latestUserInput,
          assistantOutput: cleanText,
          levelReached: level,
          success: unlocked,
        });
      } catch (error) {
        console.error("[api/sentinel] Failed to persist sentinel turn", error);
        // Keep API resilient when persistence layer is unavailable.
      }
    }

    return response;
  } catch {
    return NextResponse.json(
      { error: "Sentinel route failed." },
      { status: 500 },
    );
  }
}
