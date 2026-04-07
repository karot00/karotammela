type EventPayload = Record<string, string | number | boolean | null | undefined>;

function sanitize(payload: EventPayload) {
  return Object.fromEntries(
    Object.entries(payload).filter(([, value]) => value !== undefined),
  );
}

export function trackServerEvent(event: string, payload: EventPayload) {
  const line = JSON.stringify({
    at: new Date().toISOString(),
    event,
    payload: sanitize(payload),
  });

  console.log(`[telemetry] ${line}`);
}
