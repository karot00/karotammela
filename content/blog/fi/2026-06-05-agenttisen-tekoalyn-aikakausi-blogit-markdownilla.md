---
title: "Agenttivetoisen koodauksen aikakausi: näin kirjoitan ja julkaisen blogit Markdownilla"
description: "Miten siirtymä generatiivisista chateista autonomisiin agentteihin muuttaa sisällön hallintaa, paikallisia Markdown-tiedostoja ja tuotannon työnkulkuja vuonna 2026."
publishedAt: "2026-06-05"
slug: "agenttisen-tekoalyn-aikakausi-blogit-markdownilla"
draft: false
tags: ["AI", "Agentic-AI", "MDX", "Next.js", "Kilo-Code", "Web-Development"]
---

Olemme ylittäneet merkittävän rajapyykin siinä, miten ohjelmistoja, sivustoja ja sisältöä rakennetaan. Vuosi 2026 on jo aivan erinäköinen kuin 2025. Kehitys on huimaa niin kielimalleissa kuin työkaluissa. Olemme menneet yksinkertaisesta generatiivisesta tekoälystä kohti aitoa **agenttista tekoälyä**. Generatiivinen tekoäly vastaa kysymyksiin. Agenttinen tekoäly päättelee, suunnittelee ja tekee oikeita toimenpiteitä itsenäisesti.

Tämä sivusto pystytettiin ja julkaistiin muutamassa tunnissa. Kun projekti kasvoi, tarvitsin selkeän ja kehittäjäystävällisen tavan hallita sisältöä. Ratkaisu oli tiedostopohjainen Markdown. Raskaan tietokannan tai monimutkaisen headless-järjestelmän sijaan sisältö elää suoraan koodin vieressä.

![Kilo Code suorittaa tehtäviä työtilassa](/media/karo-tammela-agentic-ai.png)

### Muutos: generatiivinen vai agenttinen

Chatbottien aikaan pyysit tekoälyä kirjoittamaan kappaleen. Sitten kopioit sen, avasit editorin, muotoilit tekstin käsin ja teit commitin. Agenttisella aikakaudella kuvaat vain tavoitteen. Esimerkiksi: *"Kirjoita tekninen blogiteksti agenttisesta pinosta. Muotoile se projektin frontmatter-skeeman mukaan. Tarkista, että koodiesimerkit toimivat. Valmistele teksti commitia varten."*

Agentti ei ainoastaan kirjoita tekstiä. Se tutkii tiedostojärjestelmää. Se lukee validointiskeemat, kuten Zod-tiedostot ja gray-matter-lataajat. Se tarkistaa kuvapolut ja muokkaa työtilaa suoraan.

Vertaillaan ominaisuuksia:

| Ominaisuus | Chatbottien aika (generatiivinen) | Agenttinen aika (autonominen) |
| :--- | :--- | :--- |
| **Vuorovaikutus** | Yksittäiset kysymykset ja vastaukset | Tavoitelähtöinen ja toistuva silmukka |
| **Tehtävän vaativuus** | Perusluonnokset ja ehdotukset | Monivaiheinen päättely ja työnkulku |
| **Työkalut** | Ei mitään tai rajatut chat-liitännäiset | Suora tiedostojärjestelmä, paikallinen CLI ja verkkohaku |
| **Suoritus** | Ihmisen pakko kopioida ja liittää | Agentti kirjoittaa, testaa ja varmistaa koodin |

### Miksi tämä blogi toimii Markdownilla, ei MDX:llä

Käytän tätä sivustoa esimerkkinä, koska rakensin sen itse. Näin voin näyttää oikean toteutuksen yleisen oppaan sijaan. Kevyelle portfoliolle, henkilökohtaiselle sivustolle tai dokumentaatiolle tietokantapohjainen järjestelmä on liikaa. Kun jokainen teksti on pelkkä `.md`-tiedosto kansiossa `content/blog/fi/`, koko blogi pysyy gitissä. Sisältö on versioitu yhdessä koodin kanssa.

Valitsin tietoisesti puhtaan `.md`-muodon MDX:n sijaan. MDX on tehokas. Sen avulla tekstin sisään voi upottaa eläviä React-komponentteja. Tällä teholla on kuitenkin hintansa. Jokaisesta tekstistä tulee käytännössä suoritettavaa koodia, joka pitää kääntää ja johon pitää luottaa. Kirjoittamiseen keskittyvä blogi ei tarvitse vimpaimia keskelle lausetta. Tarvitsen nopeaa, ennustettavaa ja siirrettävää tekstiä. Puhdas Markdown pitää työn yksinkertaisena. Kirjoita teksti, lisää frontmatter ja tee commit. Tämän projektin lataaja lukee vain `.md`-tiedostot ja ohittaa kaiken muun.

Tämän eron näkee hyvin omissa projekteissani. Levi Finlandin blogissa oleva [Levi Golfin kausi 2026 -teksti](https://levifinland.fi/fi/blog/levi-golf-season-2026-has-started) käyttää MDX:ää, ei puhdasta Markdownia. Se on tarkoituksellinen valinta. Ne tekstit tarvitsevat selkeitä toimintapainikkeita. Rakensin siksi yhden uudelleenkäytettävän CTA-painikekomponentin. Uuteen tekstiin lisään painikkeen vain kutsumalla komponenttia ja muokkaamalla sen tekstin. Juuri tähän MDX on tehty. Tämä portfolioblogi ei tarvitse painikkeita, joten puhdas `.md` on täällä yksinkertaisempi ja turvallisempi valinta.

### Tekstitiedoston rakenne

Tyypillinen blogiteksti tällä sivustolla näyttää tältä. Se on pelkkä frontmatter-lohko ja sen perässä Markdown-teksti:

```markdown
---
title: "Agenttisen tekoälyn aikakausi"
description: "Miten agenttinen tekoäly muuttaa sisältötyötä vuonna 2026."
publishedAt: "2026-06-05"
slug: "agenttisen-tekoalyn-aikakausi-blogit-markdownilla"
draft: false
tags: ["AI", "Agentic-AI", "Markdown", "Next.js"]
---

Olemme ylittäneet merkittävän rajapyykin siinä,
miten ohjelmistoja ja sisältöä rakennetaan...
```

Miten Next.js-sovellus muuttaa tämän tiedoston turvalliseksi ja tyypitetyksi olioksi? Tässä on projektin oikea jäsennyslogiikka:

```typescript
// src/lib/blog.ts
import matter from "gray-matter";
import { z } from "zod";

const blogFrontmatterSchema = z.object({
  title: z.string().min(1),
  description: z.string().min(1),
  publishedAt: z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  slug: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/),
  draft: z.boolean().optional().default(false),
  tags: z.array(z.string().min(1)).optional().default([]),
});

function parsePostFile(filePath: string, source: string) {
  const { data, content } = matter(source);
  const parsed = blogFrontmatterSchema.parse(data);
  return {
    ...parsed,
    body: content.trim(),
  };
}
```

`gray-matter` erottaa frontmatterin tekstistä. Zod tarkistaa metatiedot tiukkaa skeemaa vasten. Jos päivämäärä on väärässä muodossa tai slug sisältää kiellettyjä merkkejä, build kaatuu heti. Näin rikkinäinen teksti ei pääse tuotantoon. Saan täyden hallinnan metatietoihin, kuten otsikkoon, julkaisupäivään ja tunnisteisiin. Samalla itse sisältö pysyy luettavana ja siirrettävänä Markdownina.

### Koodilohkojen visuaalinen tyylittely

Pelkkä tasalevyinen teksti toimii, mutta se ei ole houkutteleva. Halusin, että koodiesimerkit vetävät lukijan puoleensa. Siksi puin jokaisen koodilohkon näyttämään **terminaali-ikkunalta**. Siinä on tumma otsikkopalkki ja kolme tuttua macOS-tyylistä liikennevalopistettä. Käytin tähän pelkkää CSS:ää ilman yhtäkään lisäkirjastoa. Kaksi pseudoelementtiä kohdistuu valitsimeen `.blog-prose pre` globaalissa tyylitiedostossa:

```css
/* Terminaali-ikkunan ulkoasu blogin koodilohkoille */
.blog-prose pre {
  position: relative;
  background-color: #0d0f16;
  border: 1px solid var(--border);
  border-radius: 0.75rem;
  padding: 2.85rem 1.15rem 1.15rem; /* tilaa otsikkopalkille */
  overflow-x: auto;
}

/* Tumma otsikkopalkki */
.blog-prose pre::before {
  content: "";
  position: absolute;
  inset: 0 0 auto 0;
  height: 2.05rem;
  background: linear-gradient(180deg, #1b1f2a, #14171f);
}

/* Kolme liikennevalopistettä yhdellä box-shadow-arvolla */
.blog-prose pre::after {
  content: "";
  position: absolute;
  top: 0.72rem;
  left: 1.1rem;
  width: 0.66rem;
  height: 0.66rem;
  border-radius: 50%;
  background: #ff5f56;
  box-shadow: 1.15rem 0 0 #ffbd2e, 2.3rem 0 0 #27c93f;
}
```

Kikka on tehdä koko juttu pseudoelementeillä `::before` ja `::after`. Ensimmäinen maalaa tumman otsikkopalkin. Toinen piirtää punaisen pisteen. Keltainen ja vihreä piste loihditaan kahdella siirretyllä `box-shadow`-kopiolla. Lopputulos on korkeakontrastinen ja moderni terminaali-ilme. Se sopii insinöörimäisen näkymän tunnelmaan. Juuri tämä tyyli renderöi tämänkin tekstin koodilohkot.

### Agenttien rooli sisällön elinkaaressa

Vuonna 2026 sisältö ei ole staattista. Agenttinen tekoäly voi ajaa automaattisia tarkistuksia koko blogikirjastoon säännöllisesti. Tyypillinen ylläpitosilmukka näyttää tältä:

1. **Tarkista Markdownin rakenne.** Näin rikkinäiset linkit ja virheellinen frontmatter jäävät kiinni ennen julkaisua.
2. **Käännä luonnokset** kieleltä toiselle automaattisesti. Slugit, tunnisteet ja muotoilu säilyvät ennallaan.
3. **Varmista rajapintojen ajantasaisuus.** Agentti huomaa juuri päivittyneen kirjaston importin, tarkistaa muutoslokin, korjaa koodilohkon ja kääntää sen paikallisesti.

Tämä työnkulku muuttaa yksinkertaisen blogikansion eläväksi ja itseään korjaavaksi tietopankiksi. Kilo Coden kaltaisten työkalujen avulla kehittäjä voi keskittyä ominaisuuksien rakentamiseen. Autonominen silmukka hoitaa muotoilun, ulkoasun tarkistuksen ja repositorion hallinnan.
