---
title: "Modernin, hakukoneoptimoidun blogin rakentaminen - Markdown vai MDX"
description: "Suorituskykyisten ja lokalisoitujen blogien rakentamiseen Next.js:llä ja arkkitehtuurivalinta MDX:n tai puhtaan Markdownin välillä."
publishedAt: "2026-06-06"
slug: "moderni-seo-blogi-markdown-vai-mdx"
draft: false
tags: ["nextjs", "markdown", "mdx", "seo", "webdev"]
---

Agenttisen tekoälyn aikakaudella blogin rakentaminen ei vaadi monimutkaisia julkaisujärjestelmiä, raskaita tietokantoja tai loputonta tietoturvapaikkailua. Tässä oppaassa käyn läpi, miten rakennat erittäin suorituskykyisen, hakukoneoptimoidun ja monikielisen blogin Next.js:llä ja staattisilla tiedostoilla – hyödyntäen oikeita arkkitehtuureja, jotka olen vienyt tuotantoon [levifinland.fi-sivustolla](https://www.levifinland.fi) ja täällä omassa blogissani.

Ja mikä parasta? Tämä arkkitehtuuri käsittelee sisältöä koodina, mikä tekee siitä täydellisen optimoidun niin ihmiskehittäjille kuin tekoälyagenteillekin.

### Pragmaattinen arkkitehti: MDX vai puhdas Markdown

Ennen kuin sukellan koodiin, meidän on käsiteltävä tärkeä arkkitehtuurivalinta: tarvitsetko todella MDX:ää, vai onko puhdas Markdown (`.md`) parempi valinta?

Ylläpidän kahta eri blogialustaa tällä staattisella arkkitehtuurilla, ja valitsin kummallekin eri työkalut niiden erityistarpeiden perusteella:

#### MDX levifinland.fi:ssä

Kaupallisella, runsasliikenteisellä sivustolla kuten levifinland.fi, konversio ja käyttäjän vuorovaikutus ovat tärkeitä mittareita. Blogin täytyy tehdä muutakin kuin näyttää vain tekstiä. Sen pitää konvertoida varauksia green fee -pelilippuihin Levi Golfiin sekä mökkivuokraukseen.

Tähän MDX on kätevä valinta. Se mahdollistaa monimutkaisten, interaktiivisten React-komponenttien (kuten dynaamisten varauswidgettien, interaktiivisten reittikarttojen tai tehokkaiden toimintakehotepainikkeiden) upottamisen suoraan tavalliseen tekstisisältöön.

#### Puhdas Markdown (.md) tässä blogissa

Henkilökohtaisella portfoliollani ensisijainen tavoitteeni oli yksinkertaisuus. Halusin keskittyä täysin kirjoittamiseen ilman huolta monimutkaisten React-komponenttien kääntämisestä tai rikkinäisten JSX-tagien debuggaamisesta sisällössäni. Haluan tässä kohtaa huomauttaa, että Markdown oli minulle vielä muutama kuukausi sitten täysin vieras. Keskustellen tekoälyn kanssa ja opiskellen sen antamien ideoiden pohjalta, minusta on tullut täysin .md-uskovainen. 

Valitsemalla puhtaan Markdownin (`.md`) pidän kirjoittamiseni täysin irrallaan frontend-frameworkista. Jos päätän viiden vuoden kuluttua siirtää tämän sivuston Next.js:stä Astroon, Hugoon tai itse tehtyyn Rust-generaattoriin, `.md`-tiedostoni jäsentyvät vaivattomasti missä tahansa - out-of-the-box. [Astro](https://astro.build/) on jälleen yksi uusi teknologia, johon haluan syventyä tarkemmin seuraavina kuukausina.

Saan silti kauniin typografian pienellä omalla `.blog-prose`-tyylitiedostolla (sen sijaan että ottaisin käyttöön Tailwind Typography -lisäosan), mutta ilman selaimen puolen JavaScript-kuormaa ja äärimmäisen helpolla ylläpidettävyydellä.

### Arkkitehtuurin vertailu: perinteinen CMS vs. staattinen MD/MDX

| Ominaisuus | Perinteinen tietokanta / CMS | Markdown/MDX |
| :--- | :--- | :--- |
| **Hostauskustannus** | Palvelin- + tietokantamaksut ($/kk) | 0 € (Hostaus Vercelissä) |
| **Ylläpito** | Tietoturvapaikat, liitännäispäivitykset | Ei ylläpitoa (pelkkiä tiedostoja ja Gitiä) |
| **Tekoälyvalmius** | Monimutkaiset API-skeemat tai UI-klikkailu | Täydellinen. LLM:t lukevat/kirjoittavat Markdownia natiivisti |
| **Suorituskyky** | Tietokantakyselyt ja palvelinpuolen renderöinti | Salamannopea esirenderöity staattinen HTML |

### Mikä MDX (ja MD) on?

Markdown on kevyt merkintäkieli, jossa on tekstipohjainen muotoilusyntaksi. MDX on tehokas Markdownin laajennos, joka antaa sinun kirjoittaa JSX:ää (React-komponentteja) suoraan sisältösi sisään.

Tyypillinen MDX-tiedosto mukautetuilla elementeillä näyttää tältä:

```mdx
---
title: "Levin golfkausi on alkanut!"
description: "Varaa green fee -lippusi kaudelle 2026."
date: "2026-05-30"
author: "Levi Finland"
image: "/images/blog/levi-golf-2026/cover.jpg"
tags: ["golf", "kesä"]
draft: false
---

# Golfkausi on täällä!

Vihdoin on aika suunnata viheriölle. [Varaa peliaikasi nyt](https://greenfee.levifinland.fi).

<CTAButton href="https://greenfee.levifinland.fi">Varaa green fee</CTAButton>
```

Ylhäällä `---`-merkkien sisään käärittyä metatietoa kutsutaan frontmatteriksi. Jos käytän puhdasta Markdownia (`.md`) sivustolla kuten karotammela.fi, kirjoitan täsmälleen saman frontmatterin ja Markdownin, mutta jätän pois mukautetut React-elementit kuten `<CTAButton />` pitääkseni homman yksinkertaisena.

### Teknologiapino

Rakentaakseni moottorin näiden staattisten tiedostojen ympärille käytän joukkoa eri kirjastoja:

- **Next.js (App Router):** React-framework, joka hoitaa tiedostopohjaisen reitityksen, kuvaoptimoinnin ja staattisen renderöinnin.
- **gray-matter:** Jäsennin, joka erottaa frontmatter-metatiedon leipätekstistä.
- **Zod:** TypeScript-ensimmäinen skeemavalidointikirjasto, jolla taataan frontmatterin eheys käännösaikana. Zod on käytössä lähes kaikissa projekteissani.
- **next-mdx-remote (tai marked puhtaalle .md:lle):** Kevyt apuväline tiedostojen turvalliseen kääntämiseen HTML-elementeiksi.
- **next-intl:** Vankka lokalisointireitityskirjasto saumattomaan monikielisyyteen.

### Miten se toimii

#### 1. Tiedostorakenne

Erotan raa'at sisältötiedostot Next.js:n reitityslogiikasta. Sisältö asuu juurihakemistossa, järjestettynä slugin ja kielen mukaan:

```text
content/blog/
  levi-golf-2026/
    fi.md (tai fi.mdx)
    en.md (tai en.mdx)
  sarkitunturi-hike/
    fi.md
```

Varsinainen dynaaminen reititys tapahtuu Next.js App Routerin sisällä. Luon dynaamisen kansiorakenteen polkuun `app/[locale]/blog/[slug]/page.tsx`. Tämä tiedosto toimii käännösaikaisena kuorena, joka hakee, validoi ja renderöi tiedostot URL-parametrien perusteella.

#### 2. Sisällön lukeminen ja validointi Zodilla

Varmistaakseni etteivät ihmiskirjoittajat tai tekoälykirjoittajat riko tuotantokäännöstä unohtamalla pakollisen kentän (kuten puuttuvan kansikuvan tai julkaisupäivän), pakotan tiukan skeematarkistuksen ajonaikaisesti Zodilla.

Näin datanhakuapuväline (`src/lib/blog.ts`) hoitaa lukemisen ja validoinnin:

```typescript
import fs from 'fs';
import matter from 'gray-matter';
import { z } from 'zod';

// Määritellään tiukat säännöt sille, mitä blogipostauksen metatiedon ON sisällettävä
const BlogFrontmatterSchema = z.object({
  title: z.string().min(1),
  description: z.string().max(160), // Optimoitu SEO-metakuvauksiin
  date: z.string(),
  author: z.string(),
  image: z.string(),
  tags: z.array(z.string()),
  draft: z.boolean().default(false),
});

export function getPost(slug: string, locale: string) {
  // Tarkistetaan sekä .md- että .mdx-päätteet
  const baseDir = `content/blog/${slug}`;
  const mdxPath = `${baseDir}/${locale}.mdx`;
  const mdPath = `${baseDir}/${locale}.md`;

  const filePath = fs.existsSync(mdxPath) ? mdxPath : fs.existsSync(mdPath) ? mdPath : null;

  if (!filePath) {
    return null;
  }

  const fileContents = fs.readFileSync(filePath, 'utf8');
  const { data, content } = matter(fileContents);

  // Jäsennetään ja validoidaan frontmatter turvallisesti skeemaa vasten
  const validatedFrontmatter = BlogFrontmatterSchema.parse(data);

  return {
    frontmatter: validatedFrontmatter,
    content,
    isMdx: filePath.endsWith('.mdx'),
  };
}
```

#### 3. Staattinen sivugenerointi (SSG)

Saadakseni täydelliset suorituskykypisteet haluan Next.js:n esirenderöivän jokaisen blogipostausyhdistelmän raa'aksi HTML:ksi käännösaikana. Saavutan tämän käyttämällä `generateStaticParams`-funktiota.

Sen sijaan että kovakoodaisin tuetut kielet, tuon aktiiviset kielet suoraan `next-intl`-reititysasetuksista yhden totuuden lähteen ylläpitämiseksi:

```typescript
import { routing } from '@/i18n/routing';
import { getPostSlugs } from '@/lib/blog';

export async function generateStaticParams() {
  const slugs = getPostSlugs(); // Palauttaa esim. ['levi-golf-2026', 'sarkitunturi-hike']
  const locales = routing.locales; // Palauttaa esim. ['en', 'fi']

  // Muotoutuu: [{ slug: 'levi-golf-2026', locale: 'en' }, { slug: 'levi-golf-2026', locale: 'fi' }, ...]
  return slugs.flatMap((slug) =>
    locales.map((locale) => ({ slug, locale }))
  );
}
```

Kun vierailija tulee blogiin, edge-verkko tarjoilee kevyen, täysin valmiin HTML-tiedoston välittömästi. Ei tietokantahakuja, ei kylmäkäynnistyksiä.

### SEO ja GEO (generatiivisten hakukoneiden optimointi)

Tekoälyhakukoneille (kuten Perplexityn ja Googlen AI Overviews'n) selkeä semantiikka ja eksplisiittinen datan kartoitus ovat elintärkeitä.

#### Strukturoitu data (JSON-LD)

Tehdäkseni indeksoinnin helpoksi verkkohakuroboteille ja tekoälylle, injektoin semanttisen `BlogPosting`-skeeman suoraan dokumentin headiin.

```tsx
export default async function BlogPostPage({ params }: Props) {
  // Next.js 16:ssa reitin `params` on Promise, joka pitää awaitata
  const { slug, locale } = await params;
  const post = await getPost(slug, locale);
  if (!post) return notFound();

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    "headline": post.frontmatter.title,
    "description": post.frontmatter.description,
    "datePublished": post.frontmatter.date,
    "image": post.frontmatter.image,
    "author": {
      "@type": "Person",
      "name": post.frontmatter.author,
      "url": "https://karotammela.fi"
    }
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      {/* Artikkelin sisältö */}
    </>
  );
}
```

#### Hreflang ja kansainvälistäminen

Hakukonerobotit rankaisevat huonosti kartoitetusta lokalisoidusta sisällöstä. Käyttämällä Next.js:n metatietogenerointia ilmoitan eksplisiittiset kielivaihtoehdot ja varmistan, että oikea versio tarjoillaan käyttäjälle.

```typescript
export async function generateMetadata({ params }: Props) {
  const baseUrl = 'https://karotammela.fi';
  // Next.js 16:ssa reitin `params` on Promise, joka pitää awaitata
  const { slug } = await params;

  return {
    alternates: {
      languages: {
        fi: `${baseUrl}/fi/blog/${slug}`,
        en: `${baseUrl}/en/blog/${slug}`,
        'x-default': `${baseUrl}/en/blog/${slug}`
      }
    }
  };
}
```

### Miten renderöin sisällön: MDX vs. puhdas MD

Riippuen siitä, onko tiedosto `.md`- vai `.mdx`-tiedosto, käännän sisällön eri tavalla:

#### MDX:n konepellin alla (käytössä Levi Finlandilla)

MDX on säädetty tunnistamaan erityiset koodinpätkät, joten kirjoittajat voivat lisätä vaikkapa <CTAButton>-painikkeen suoraan tekstin joukkoon.":

```tsx
import Link from 'next/link';
import { MDXRemote } from 'next-mdx-remote/rsc';

const customComponents = {
  a: ({ href, ...props }: any) => (
    <Link className="text-blue-600 underline hover:text-blue-800" href={href} {...props}/>
  ),
  CTAButton: ({ href, children }: { href: string; children: React.ReactNode }) => (
    <Link className="inline-block bg-blue-600 hover:bg-blue-700 text-white font-semibold px-8 py-4 rounded-xl transition-all transform hover:-translate-y-0.5 shadow-lg" href={href}>
      {children}
    </Link>
  ),
};

export default function PostBody({ content }: { content: string }) {
  return <MDXRemote components={customComponents} source={content}/>;
}
```

#### Puhdas .md konepellin alla (käytössä karotammela.fi:llä)

Pelkkiä tekstitiedostoja käytettäessä en tarvitse raskaita taustatoimintoja tai monimutkaista JavaScript-koodia. Jäsennän tiedoston yksinkertaisella markdown-jäsentimellä (`marked`) ja ajan tulosteen sitten `sanitize-html`:n läpi selkeällä sallittujen tagien listalla ennen kuin injektoin sen mukautettujen `.blog-prose`-tyylien sisään. Vaikka sisältö on kirjoittajalle luotettavaa, sanitointi on totuuden lähde, joka pitää merkkauksen turvallisena ja ennustettavana:

```tsx
import { marked } from 'marked';
import sanitizeHtml from 'sanitize-html';

export default function PostBody({ content }: { content: string }) {
  // marked.parse voi palauttaa string | Promise<string>, joten varmistetaan tyyppi
  const rendered = marked.parse(content, { breaks: true, gfm: true });
  const html = typeof rendered === 'string' ? rendered : '';

  const safeHtml = sanitizeHtml(html, {
    allowedTags: sanitizeHtml.defaults.allowedTags.concat(['img', 'pre', 'code']),
  });

  return (
    <div
      className="blog-prose max-w-none"
      dangerouslySetInnerHTML={{ __html: safeHtml }}
    />
  );
}
```

Tämä on elegantti, nopea, ilman JS:n tuomaa kuormaa ja nojaa web-standardeihin.

### Tekoälyn kanssa kirjoittamisen työnkulku

Valitsetpa Markdownin tai MDX:n, siirtyminen graafisesta tietokantanäkymästä tiedostopohjaiseen rakenteeseen parantaa radikaalisti sitä, miten kirjoitat koodia ja sisältöä samanaikaisesti. Näin tekoälyavustaja kiihdyttää tätä asetusta:

- **Automaattinen pohjakoodi:** Pyydä LLM:ää generoimaan monimutkaisia Zod-skeemarakenteita tai tyyppimäärittelyjä sekunneissa.
- **Saumattomat lokalisoinnit:** Ohjeista tekoälyagenttia lukemaan `content/blog/post-slug/en.md` ja tuottamaan täydellisen idiomaattisen, frontmatter-yhteensopivan `fi.md`-käännöksen täsmälleen samaan hakemistorakenteeseen.
- **Markdown on tekoälyn äidinkieli:** Suuret kielimallit on koulutettu natiivisti Markdown-syntaksilla. Tekoälykirjoittaja ei koskaan tee syntaksivirheitä generoidessaan standardia `.md`:tä, kun taas se voi joskus kompastua sulkemattomiin JSX-tageihin monimutkaisissa MDX-ratkaisuissa.

### Yhteenveto: valitse se, mikä sopii tavoitteeseesi

Jos rakennat kaupallista tuotetta tai sivustoa, joka vaatii erittäin mukautettuja interaktiivisia suppiloita, toimintakehotteita tai interaktiivisia kaavioita, MDX on valintasi. Tämä antaa sinulle erinomaisen tasapainon staattisen sisällön ja dynaamisten sovellusten välillä.

Mutta jos rakennat henkilökohtaista sivustoa tai teknistä päiväkirjaa, älä tee liian hienoa. Puhdas Markdown on kehittäjän ja tekoälyn paras ystävä. Se antaa sinulle helpon siirrettävyyden, yksinkertaisuuden ja erinomaisen suorituskyvyn ilman monimutkaisten käännösketjujen painolastia. Tekoäly rakastaa Markdownia ja se ymmärtää sitä paremmin kuin MDX:ää.
