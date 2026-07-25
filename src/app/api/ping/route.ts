import { NextResponse } from "next/server";

import { comparisonCorsHeaders } from "@/lib/comparison";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

/**
 * GET /api/ping — comparison probe for the Tech Switcher performance widget.
 *
 * Identifies this stack as `next`, never caches, and does no DB or external
 * work so the browser measures pure TTFB. CORS allows the Go apex (and this
 * subdomain) so the widget can ping across origins.
 */
export function GET(request: Request) {
  const cors = comparisonCorsHeaders(request.headers.get("origin"));

  return NextResponse.json(
    { status: "ok", stack: "next" },
    {
      headers: {
        ...cors,
        "Cache-Control": "no-store",
      },
    },
  );
}

export function OPTIONS(request: Request) {
  const cors = comparisonCorsHeaders(request.headers.get("origin"));

  return new NextResponse(null, {
    status: 204,
    headers: {
      ...cors,
      "Cache-Control": "no-store",
    },
  });
}
