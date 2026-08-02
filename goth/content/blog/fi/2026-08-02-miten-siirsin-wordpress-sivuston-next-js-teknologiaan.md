---
title: "Miten siirtää WordPress-sivusto Next.js-aikaan ilman että SEO-näkyvyys kärsii?"
description: "Miten siirrän WordPress-sivuston uuteen teknologiaan yhdessä työpäivässä? Katso, miten tekoälyavusteinen web-kehitys nopeuttaa migraatiota ja suojaa Google-sijoitukset."
publishedAt: "2026-08-02"
slug: "miten-siirsin-wordpress-sivuston-next-js-teknologiaan"
draft: false
tags: ["Next.js", "WordPress", "Tekoäly", "Web-kehitys", "SEO"]
---

# Miten siirtää WordPress-sivusto Next.js-aikaan ilman että SEO kärsii?

Monen yrityksen verkkosivu saavuttaa pisteen, jossa ylläpito kankeutuu ja sivusto hidastuu, mutta vanha Google-näkyvyys halutaan säilyttää entisellään. Tässä casessa 10 vuotta vanha WordPress/Elementor-sivusto ([Hiihtogreeni.fi](https://hiihtogreeni.fi/)) siirrettiin moderniin Next.js 16 -arkkitehtuuriin yhdessä työpäivässä tekoälyä hyödyntäen.

Tämä blogikirjoitus on tiivis katsaus migraation tärkeimpiin ratkaisuihin. Olen koonnut koko toteutuksen yksityiskohtaiseksi PDF-oppaaksi, jossa käyn läpi myös koodiesimerkit, tekoälyagenttien hallinnan ja suorituskykyoptimoinnin vaiheet.

## TL;DR: Projektin ydin pähkinänkuoressa

- **Lähtötilanne:** Hiihtogreeni.fi-mökkisivusto toimi kankealla WordPress/Elementor-yhdistelmällä.
- **Tavoite:** Siirtää sivusto täysin SEO-turvallisesti moderniin Next.js 16 -ympäristöön.
- **Toteutus:** Tekoälyavusteinen koodaus agentteja hyödyntäen (Kilo Code/Kiloclaw), staattinen generointi (SSG) sekä millintarkat 301- ja 410-ohjaukset.
- **Lopputulos:** Salamannopea ja ylläpitovapaa sivusto yhdessä työpäivässä, noin kahdeksassa tunnissa, ilman hakukonesijoitusten menetystä. Lomakeroskapostit putosivat heti nollaan.

![Hiihtogreeni.fi siirrettiin WordPressistä Next.js-aikaan](/media/from_wordpress_to_next_js_hiihtogreeni.png)

## Voiko WordPress-sivuston siirtää ilman SEO-riskiä?

Suurin virhe sivustouudistuksessa on sokea koodaus ilman vanhan sivuston auditointia. Projekti aloitettiin scraping-menetelmällä vanha sivusto Kiloclaw-palvelulla, jotta kaikki piilotetut ja vanhat indeksoidut URL-osoitteet saatiin kartoitettua ennen koodin kirjoittamista.

Auditoinnissa löytyi myös vanha englanninkielinen galleria-URL, joka oli edelleen julkinen ja Googlen indeksoima. Ilman auditointia tämä sisältö olisi voinut kadota uudistuksen yhteydessä.

## Miten Google-näkyvyys säilytetään sivustouudistuksessa?

- WordPressin kuva- ja PDF-polut, kuten `/wp-content/uploads/`, tuotiin sellaisenaan uuden Next.js-projektin `public/`-kansioon. Vanhat linkit palauttavat aidon `200 OK` -vastauksen.
- Muuttuneet osoitteet ohjattiin pysyvästi `301`-uudelleenohjauksilla.
- Poistuneille WordPress-järjestelmäpoluille, kuten `/wp-admin/`, `/feed/` ja `/wp-json/`, asetettiin `410 Gone` etusivulle ohjaamisen sijaan.

![WordPress-sivuston PageSpeed-analyysi ennen migraatiota](/media/pagespeed_insights_hiihtogreeni.fi_wordpress.png)

## Miksi Next.js-sivusto on WordPress-sivustoa nopeampi?

Sivusto toteutettiin staattisella sivujen generoinnilla (SSG) React 19:n, Tailwind CSS 4:n ja TypeScriptin avulla. Mobiilin PageSpeed-pisteet nostettiin 72:sta 95:een luomalla hero-kuville rinnakkainen LCP-latausputki AVIF- ja WebP-formaateilla.

![Next.js-sivuston PageSpeed-analyysi migraation jälkeen](/media/excellent_page_speed_with_next_js.png)

## Yhteydenottolomake ilman roskapostia

Uudessa Next.js-toteutuksessa yhteydenotot ohjataan Resend-palveluun Server Actions- tai Route-rajapinnan kautta ilman selaimessa näkyviä API-avaimia. Tämä pudotti roskapostien määrän heti nollaan.

![Roskapostilta suojattu Next.js-yhteydenottolomake](/media/contact_form_with_no_spam.png)

## Paljonko tekoälyavusteinen migraatio maksaa?

Projektin aktiivinen työaika oli noin kahdeksan tuntia. Käytetyt työkalut olivat Kiloclaw vanhan sivuston datan keräämiseen, Qwen3.7 max arkkitehtuurin pohjatöihin, GPT-5.6 Sol migraatiostrategiaan ja auditointeihin sekä Hy3 toistuviin käyttöliittymä- ja redirect-tehtäviin.

Tekoälyn API-kustannukset olivat yhteensä noin **9,55 €**. Lisäksi Purelymail-sähköpostipalvelu maksaa 9 € vuodessa.

## Haluatko tietää täsmälleen, miten migraatio tehtiin?

Olen kirjoittanut aiheesta yksityiskohtaisen oppaan ja teknisen projektipäiväkirjan. PDF-opas sisältää koodiesimerkit, tekoälyagenttien hallintamallit (`AGENTS.md` ja `memory.md`) sekä suorituskykyoptimoinnin vaiheet.

> **[Lataa koko PDF-opas WordPress-sivuston siirtämisestä Next.js-aikaan](/media/Wordpress-sivusto%20next.js-aikaan.pdf)**
