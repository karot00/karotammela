import { NextResponse } from "next/server";
import { Resend } from "resend";
import { z } from "zod";

import { enforceRateLimit, getClientIp } from "@/lib/security/rate-limit";
import { trackServerEvent } from "@/lib/telemetry/events";

const requestSchema = z.object({
  senderName: z.string().trim().min(2).max(80),
  recipientEmail: z.string().trim().email().max(200),
  greeting: z.string().trim().min(1).max(250),
  base64Image: z.string(),
  locale: z.enum(["fi", "en"]).optional().default("fi"),
  website: z.string().optional(),
});

export const runtime = "nodejs";

function getResendConfig() {
  const apiKey = process.env.RESEND_API_KEY;
  const from = process.env.CONTACT_FROM_EMAIL; // Use existing verified sender

  if (!apiKey || !from) {
    return null;
  }

  return { apiKey, from };
}

function escapeHtml(text: string) {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

export async function POST(request: Request) {
  const ip = getClientIp(request);

  // Rate limit: Max 5 card sends per IP per minute
  const ipRate = enforceRateLimit({
    scope: "send-card-ip",
    key: ip,
    limit: 5,
    windowMs: 60_000,
  });

  if (!ipRate.allowed) {
    trackServerEvent("postcard.rate_limited", { ip });

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

  // Daily IP Rate limit: Max 10 card sends per IP per 24 hours
  const ipDailyRate = enforceRateLimit({
    scope: "send-card-ip-daily",
    key: ip,
    limit: 10,
    windowMs: 24 * 60 * 60 * 1000,
  });

  if (!ipDailyRate.allowed) {
    trackServerEvent("postcard.daily_rate_limited", { ip });

    return NextResponse.json(
      { error: "Daily sending limit reached. Try again tomorrow." },
      {
        status: 429,
        headers: {
          "Retry-After": String(Math.ceil(ipDailyRate.retryAfterMs / 1000)),
        },
      },
    );
  }

  try {
    const body = await request.json();
    const parsed = requestSchema.safeParse(body);

    if (!parsed.success) {
      return NextResponse.json(
        { error: "Invalid parameters. Please check your inputs." },
        { status: 400 }
      );
    }

    const { senderName, recipientEmail, greeting, base64Image, locale, website } = parsed.data;

    // Honeypot check: If filled, silently mock success
    if (website && website.trim().length > 0) {
      trackServerEvent("postcard.honeypot_triggered", { ip });
      return NextResponse.json({ ok: true });
    }

    // Block links/URLs in name or greeting
    const urlPattern = /https?:\/\/[^\s]+|www\.[^\s]+|\.[a-z]{2,}\/[^\s]*/i;
    if (urlPattern.test(greeting) || urlPattern.test(senderName)) {
      trackServerEvent("postcard.spam_detected", { ip, reason: "url_in_content" });
      return NextResponse.json(
        { error: "Links or URLs are not allowed in the card content." },
        { status: 400 }
      );
    }

    // Recipient Rate limit: Max 3 card sends to the same email address per 24 hours
    const recipientRate = enforceRateLimit({
      scope: "send-card-recipient-daily",
      key: recipientEmail.toLowerCase().trim(),
      limit: 3,
      windowMs: 24 * 60 * 60 * 1000,
    });

    if (!recipientRate.allowed) {
      trackServerEvent("postcard.recipient_rate_limited", { ip, recipientEmail });

      return NextResponse.json(
        { error: "This recipient has received too many cards today. Try again tomorrow." },
        {
          status: 429,
          headers: {
            "Retry-After": String(Math.ceil(recipientRate.retryAfterMs / 1000)),
          },
        },
      );
    }

    const config = getResendConfig();
    if (!config) {
      return NextResponse.json(
        { error: "Email delivery is not configured." },
        { status: 503 }
      );
    }

    // Clean up base64 prefix if present
    const cleanBase64 = base64Image.includes(",")
      ? base64Image.split(",")[1]
      : base64Image;

    const resend = new Resend(config.apiKey);

    const subject =
      locale === "fi"
        ? `Digitaalinen postikortti lähettäjältä ${senderName}! ❤️`
        : `A digital postcard greeting from ${senderName}! ❤️`;

    const html =
      locale === "fi"
        ? `
        <div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 24px; color: #1f2937;">
          <p style="font-size: 16px; line-height: 1.6; margin-bottom: 24px;">
            Hei! <strong>${escapeHtml(senderName)}</strong> on luonut ja lähettänyt sinulle personoidun digitaalisen postikortin:
          </p>
          
          <div style="margin: 24px 0; text-align: center;">
            <img src="cid:postcard-image" alt="Postikortti" width="600" style="display: block; width: 100%; max-width: 100%; height: auto; border-radius: 12px; border: 1px solid #e5e7eb; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1);" />
          </div>

          <p style="font-size: 14px; color: #4b5563; line-height: 1.6; margin-top: 32px; border-top: 1px solid #e5e7eb; padding-top: 16px; text-align: center;">
            Tämä sähköposti lähetettiin osoitteesta <a href="https://karotammela.fi" style="color: #c2410c; text-decoration: none; font-weight: 500;">karotammela.fi</a>.
          </p>
        </div>
      `
        : `
        <div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 24px; color: #1f2937;">
          <p style="font-size: 16px; line-height: 1.6; margin-bottom: 24px;">
            Hello! <strong>${escapeHtml(senderName)}</strong> has created and sent a personalized digital postcard to you:
          </p>
          
          <div style="margin: 24px 0; text-align: center;">
            <img src="cid:postcard-image" alt="Postcard" width="600" style="display: block; width: 100%; max-width: 100%; height: auto; border-radius: 12px; border: 1px solid #e5e7eb; box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1);" />
          </div>

          <p style="font-size: 14px; color: #4b5563; line-height: 1.6; margin-top: 32px; border-top: 1px solid #e5e7eb; padding-top: 16px; text-align: center;">
            This email was sent from <a href="https://karotammela.fi" style="color: #c2410c; text-decoration: none; font-weight: 500;">karotammela.fi</a>.
          </p>
        </div>
      `;

    await resend.emails.send({
      from: `karotammela.fi Postikortti <${config.from}>`,
      to: [recipientEmail],
      replyTo: config.from,
      subject,
      html,
      attachments: [
        {
          filename: "postikortti.png",
          content: Buffer.from(cleanBase64, "base64"),
          contentId: "postcard-image",
        },
      ],
    });

    trackServerEvent("postcard.submitted", {
      senderName,
      recipientEmail,
      locale,
    });

    return NextResponse.json({ ok: true });
  } catch (error: unknown) {
    const errorMsg = error instanceof Error ? error.message : String(error);
    trackServerEvent("postcard.failed", { error: errorMsg });
    return NextResponse.json(
      { error: "Failed to process card delivery request." },
      { status: 500 }
    );
  }
}
