import Script from "next/script";

export function UmamiScript() {
  const websiteId = process.env.NEXT_PUBLIC_UMAMI_WEBSITE_ID;
  const scriptUrl = process.env.NEXT_PUBLIC_UMAMI_SCRIPT_URL;

  if (!websiteId || !scriptUrl) {
    return null;
  }

  return (
    <Script
      src={scriptUrl}
      strategy="afterInteractive"
      data-website-id={websiteId}
      data-auto-track="true"
    />
  );
}
