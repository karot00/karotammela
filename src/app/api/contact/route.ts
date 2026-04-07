import { NextResponse } from "next/server";
import { Resend } from "resend";
import { z } from "zod";

import { enforceRateLimit, getClientIp } from "@/lib/security/rate-limit";
import { trackServerEvent } from "@/lib/telemetry/events";

const requestSchema = z.object({
  name: z.string().trim().min(2).max(80),
  email: z.string().trim().email().max(200),
  message: z.string().trim().min(10).max(4000),
  company: z.string().trim().max(120).optional(),
  website: z.string().trim().max(1).optional(),
});

export const runtime = "nodejs";

function getContactConfig() {
  const apiKey = process.env.RESEND_API_KEY;
  const from = process.env.CONTACT_FROM_EMAIL;
  const to = process.env.CONTACT_TO_EMAIL;

  if (!apiKey || !from || !to) {
    return null;
  }

  return { apiKey, from, to };
}

export async function POST(request: Request) {
  const ip = getClientIp(request);
  const ipRate = enforceRateLimit({
    scope: "contact-ip",
    key: ip,
    limit: 8,
    windowMs: 60_000,
  });

  if (!ipRate.allowed) {
    trackServerEvent("contact.rate_limited", { ip });

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
    return NextResponse.json({ error: "Invalid contact payload." }, { status: 400 });
  }

  if (parsed.data.website) {
    trackServerEvent("contact.honeypot_triggered", { ip });
    return NextResponse.json({ ok: true });
  }

  const config = getContactConfig();
  if (!config) {
    return NextResponse.json({ error: "Contact delivery is not configured." }, { status: 503 });
  }

  const resend = new Resend(config.apiKey);

  try {
    await resend.emails.send({
      from: config.from,
      to: [config.to],
      replyTo: parsed.data.email,
      subject: `karotammela.fi contact: ${parsed.data.name}`,
      text: [
        `Name: ${parsed.data.name}`,
        `Email: ${parsed.data.email}`,
        parsed.data.company ? `Company: ${parsed.data.company}` : null,
        "",
        parsed.data.message,
      ]
        .filter(Boolean)
        .join("\n"),
    });

    trackServerEvent("contact.submitted", {
      hasCompany: Boolean(parsed.data.company),
    });

    return NextResponse.json({ ok: true });
  } catch {
    return NextResponse.json({ error: "Failed to send contact email." }, { status: 500 });
  }
}
