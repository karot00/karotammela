import type { Metadata } from "next";
import { cookies } from "next/headers";
import { Geist_Mono, Plus_Jakarta_Sans } from "next/font/google";

import { UmamiScript } from "@/components/umami-script";
import {
  CONSENT_COOKIE_NAME,
  CookieConsentProvider,
  parseConsentFromServerCookie,
} from "@/modules/cookie-consent";
import {
  SITE_DESCRIPTION,
  SITE_NAME,
  SITE_OWNER_NAME,
  getDefaultSocialImage,
} from "@/lib/seo";
import { getSiteUrl, toAbsoluteUrl } from "@/lib/site-url";

import "./globals.css";

const plusJakartaSans = Plus_Jakarta_Sans({
  variable: "--font-plus-jakarta-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL(getSiteUrl()),
  icons: {
    icon: "/karo-tammela-tammenterho.svg",
  },
  title: {
    default: SITE_NAME,
    template: `%s | ${SITE_NAME}`,
  },
  description: SITE_DESCRIPTION,
  openGraph: {
    type: "website",
    siteName: SITE_NAME,
    title: SITE_NAME,
    description: SITE_DESCRIPTION,
    url: toAbsoluteUrl("/"),
    images: [getDefaultSocialImage()],
  },
  twitter: {
    card: "summary_large_image",
    title: SITE_NAME,
    description: SITE_DESCRIPTION,
    images: [getDefaultSocialImage().url],
  },
  alternates: {
    canonical: toAbsoluteUrl("/"),
  },
  authors: [{ name: SITE_OWNER_NAME }],
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  const consentCookieValue = cookieStore.get(CONSENT_COOKIE_NAME)?.value;
  const initialConsent = parseConsentFromServerCookie(consentCookieValue);

  return (
    <html
      className={`${plusJakartaSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <head>
        <link
          rel="preconnect"
          href="https://api-gateway.umami.dev"
          crossOrigin="anonymous"
        />
      </head>
      <body className="min-h-full flex flex-col">
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "WebSite",
              name: SITE_NAME,
              url: toAbsoluteUrl("/"),
              inLanguage: ["fi-FI", "en-US"],
              publisher: {
                "@type": "Person",
                name: SITE_OWNER_NAME,
                url: toAbsoluteUrl("/"),
              },
            }),
          }}
        />
        <CookieConsentProvider initialConsent={initialConsent}>
          {children}
          <UmamiScript />
        </CookieConsentProvider>
      </body>
    </html>
  );
}
