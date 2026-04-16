import type { Metadata } from "next";
import { cookies } from "next/headers";
import { Geist_Mono, Plus_Jakarta_Sans } from "next/font/google";

import { UmamiScript } from "@/components/umami-script";
import {
  CONSENT_COOKIE_NAME,
  CookieConsentProvider,
  parseConsentFromServerCookie,
} from "@/modules/cookie-consent";

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
  title: "karotammela.fi",
  description: "Agentic AI architect portfolio and Sentinel challenge",
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
      lang="en"
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
        <CookieConsentProvider initialConsent={initialConsent}>
          {children}
          <UmamiScript />
        </CookieConsentProvider>
      </body>
    </html>
  );
}
