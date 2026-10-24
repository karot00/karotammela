---
title: "llms.txt, GEO ja agenttinen verkko: Näin valmistelet sivustosi tekoälyaikakauteen"
description: "Mitä llms.txt, llms-full.txt ja GEO oikeasti tekevät, mitä ne eivät tee ja miten agenttinen kaupankäynti muuttaa verkkosivuja. Esimerkkinä oma UTM-työkaluni Kanavoi.com."
publishedAt: "2026-10-24"
slug: "agenttinen-shoppailu-ja-tekoalyn-nettiselaus"
draft: false
tags: ["tekoäly", "llms.txt", "GEO", "SEO", "agentit", "verkkokauppa"]
---

Internetin informaatioarkkitehtuuri on kokenut historiansa aikana suuria murroksia. Ensimmäisessä aallossa verkkosivut rakennettiin ihmisten selaimille ja visuaaliseen selailuun. Toisessa aallossa ne optimoitiin Googlen kaltaisille hakuroboteille ja kymmenen sinisen linkin hakutulossivuille (SEO).

Nyt elämme kolmatta, ehkä historian nopeinta murrosta: **siirtymää ihmisten verkosta kohti kielimallien ja tekoälyagenttien verkkoa**.

Microsoft Advertisingin vuonna 2026 julkaisema opas [*Prepare for the Age of AI Shopping*](https://about.ads.microsoft.com/content/dam/sites/msa-about/global/common/content-lib/pdf/prepare-for-the-age-of-ai-shop.pdf) kokoaa yhteen tuoreita lukuja:

- HUMAN Securityn vuoden 2026 raportin mukaan **automatisoitu liikenne kasvaa verkossa kahdeksan kertaa ihmisliikennettä nopeammin**, ja tekoälyn ohjaamat istunnot lähes kolminkertaistuivat vuonna 2025.
- Adobe Analyticsin (Q1 2026) mukaan **tekoälyn kautta saapuvat ostajat konvertoivat 42 % paremmin** kuin muu liikenne.
- Microsoftin yhdysvaltalaisessa kuluttajakyselyssä (n = 1 028) **yli 72 % vastaajista odotti brändeiltä tekoälypohjaisia ostokokemuksia vuoden sisällä**.

Tekoälyt eivät enää vain hae linkkejä. Ne lukevat, vertailevat, tekevät päätöksiä ja yhä useammin myös suorittavat toimintoja ihmisten puolesta. Siksi verkkosivujen kehittäjien ja markkinoijien on kysyttävä uusi kysymys: ymmärtääkö tekoäly sivustoni sisällön oikein?

Käytän tässä kirjoituksessa esimerkkinä omaa projektiani, [Kanavoi.comia](https://kanavoi.com). Kanavoi on suomalainen UTM-työkalu markkinointitiimeille ja -toimistoille. Se pitää kampanjalinkkien nimeämisen yhtenäisenä ja näyttää jo ennen julkaisua, missä Google Analytics 4:n kanavassa (esimerkiksi Paid Social tai Email) liikenne raportoidaan. Kanavoi on sopiva esimerkki, koska se on aito, pieni ja tuore sivusto. Kun aloitin tämän tekstin, siitä puuttui vielä osa tekoälyajan perusasioista, joten toteutin ne kirjoittamisen lomassa. Kaikki alla olevat esimerkit ovat Kanavoin oikeita, tuotannossa olevia tiedostoja.

Tein Kanavoin ilmaisella UTM-generaattorilla linkin tähän kirjoitukseen LinkedIn-julkaisua varten. Kampanjan nimi siistiytyy automaattisesti (lokakuu 2026 → lokakuu_2026), ja GA4-kanava (Organic Social) näkyy jo ennen kuin linkki lähtee jakoon:

![Kanavoin UTM-generaattori: linkki tähän blogikirjoitukseen LinkedIn-julkaisua varten, GA4-kanavana Organic Social](/media/kanavoi-utm-generaattori-ga4-kanava.png)

Tekoälyaikakauden keskusteluun on noussut kolme käsitettä: avoin standardiehdotus **`llms.txt`**, sen rinnalle vakiintunut **`llms-full.txt`** ja laajempi optimointioppi **GEO (Generative Engine Optimization)**. Ne sekoitetaan usein toisiinsa, joten erotellaan ne tässä kunnolla.

---

### 1. Iso kuva: Verkon kolme aikakautta

Microsoftin opas jakaa verkossa asioinnin kolmeen rinnakkaiseen tapaan:

```
+-----------------------------------------------------------------------------------+
| 1. THE HUMAN WEB           2. THE LLM WEB              3. THE AGENTIC WEB         |
| "Help me find it"    -->   "Help me choose"      -->   "Do it for me"             |
|                                                                                   |
| Perinteinen selain,        Kielimallit tiivistävät,    Ihminen asettaa tavoitteen |
| välilehdet, hakukoneen     vertailevat vaihtoehtoja    ja agentti etsii, arvioi   |
| linkit ja manuaalinen      ja kuratoivat suosituksen.  ja vie ostoksen tai        |
| tutkimustyö.               Ihminen tekee päätöksen.    tehtävän loppuun.          |
+-----------------------------------------------------------------------------------+
```

Kaikki kolme ovat käytössä samaan aikaan. Ihmisten verkko tuo yhä valtaosan liikenteestä, mutta suurimmat uudet mahdollisuudet ovat kielimallien ja agenttien verkossa.

#### Jokaisessa kohtaamisessa on nyt kolmas osapuoli

Aiemmin markkinoijan ja verkkosivukehittäjän tavoitteena oli vakuuttaa **ihminen**: panostettiin näyttäviin kuviin, visuaaliseen tarinankerrontaan, animaatioihin ja taitavasti sijoiteltuihin toimintakehotteisiin (CTA).

Nyt mukana on usein **kolmas osapuoli: tekoälyassistentti tai -agentti**.

- Tekoälyä eivät vakuuta CSS-animaatiot, liukuvärit tai brändivideot.
- Kielimallit ja agentit arvioivat tuotteita ja palveluja ensisijaisesti **rakenteisen, koneluettavan ja ajantasaisen tiedon** perusteella.
- Agentit osaavat kyllä lukea tavallisia verkkosivuja, mutta silloin niiden on pääteltävä tiedot ihmisille tehdystä sisällöstä. Se johtaa helposti virheisiin. Jos tieto on puutteellista tai vaikeasti tulkittavaa, **on suuri riski, että tuotteesi ei päädy suositukseen lainkaan**.

---

### 2. Verkkosivujen "kohinaongelma"

Nykyaikainen verkkosivusto on tekoälymallille kallis, hidas ja riskialtis luettava:

1. **HTML-kuona ja skriptit:** Sivustot ovat täynnä JavaScript-koodia, seurantapikseleitä, evästebannereita, ponnahdusikkunoita ja navigointipalkkeja.
2. **Konteksti-ikkunan tuhlaus:** Kielimallien käsittelemä tieto mitataan tokeneina. Raskaasti skriptatulla sivulla suuri osa ladatusta datasta on esitys- tai seurantakoodia, jolla ei ole mallille mitään arvoa.
3. **Virhetulkintojen riski:** Kun asiasisältö hukkuu mainosten ja suosituslinkkien sekaan, kasvaa riski, että malli poimii väärän hinnan, vanhentuneen ehdon tai keksii puuttuvan tiedon itse (hallusinoi).

Tähän ongelmaan on tarjottu kahta eri lähestymistapaa:

- **`llms.txt`:** Answer.AI:n Jeremy Howard ehdotti syyskuussa 2024 Markdown-muotoista tiedostoa sivuston juureen. Se on kielimallille tarkoitettu tiivis reittiopas sivuston tärkeimpään sisältöön. Ehdotus on julkaistu osoitteessa [llmstxt.org](https://llmstxt.org).
- **GEO (Generative Engine Optimization):** Termi on peräisin vuonna 2023 julkaistusta [tutkimusartikkelista](https://arxiv.org/abs/2311.09735) (Aggarwal ym., mm. Princeton ja IIT Delhi). GEO ei ole tiedostomuoto, vaan laajempi oppi siitä, miten sisältö kirjoitetaan ja rakennetaan niin, että generatiiviset hakukoneet nostavat sen vastauksiinsa ja viittaavat siihen. Tutkimuksen mukaan näkyvyyttä paransivat esimerkiksi lähdeviitteet, lainaukset ja konkreettiset tilastot.

`llms.txt` on siis yksi GEO:n työkalu, ei sen synonyymi.

---

### 3. Auttavatko nämä tiedostot Claudea, ChatGPT:tä ja Geminiä?

Tähän kannattaa vastata rehellisesti, koska aiheesta liikkuu paljon liioittelua. Tekoäly hyödyntää verkkosisältöä kolmella tasolla:

| Käyttötilanne | Miten tekoäly toimii | llms.txt:n vaikutus |
| :--- | :--- | :--- |
| **Agentit ja koodaustyökalut** (esim. Claude Code, Cursor, GitHub Copilot, OpenAI Codex) | Hakevat dokumentaatiota ja lähdemateriaalia URL-osoitteesta, kun käyttäjä tai agentti ohjaa ne sinne. | **Selvin hyöty.** Kun agentti ohjataan sivustolle, `llms.txt` kertoo heti, mitkä sivut kannattaa lukea, ja säästää tokeneita. Siksi dokumentaatiosivustot julkaisevat sen yleisimmin. |
| **Reaaliaikainen haku chateissa** (ChatGPT search, Gemini, Microsoft Copilot, Perplexity) | Hakevat tuloksia hakuindeksistä ja lukevat löytyneet sivut, yleensä HTML-muodossa. | **Välillinen.** Suuret palvelut eivät ole kertoneet käyttävänsä `llms.txt`:tä, ja Google on todennut, ettei Google-haku huomioi sitä. Ratkaisevaa on, että itse sivujen sisältö on selkeää ja helposti poimittavaa. |
| **Mallien koulutus** | Crawlerit, kuten GPTBot ja ClaudeBot, haravoivat verkkoa massana. | **Ei todennettua vaikutusta.** Pääsyä säädellään edelleen `robots.txt`-tiedostolla. |

Lokakuussa 2026 tilanne on tämä: yksikään suurista tekoälyyhtiöistä ei ole sitoutunut lukemaan `llms.txt`-tiedostoja automaattisesti, vaikka monet, Anthropic mukaan lukien, julkaisevat sellaisen omasta dokumentaatiostaan. Riippumattomissa mittauksissa tiedoston olemassaololla ei ole havaittu yhteyttä siihen, kuinka usein kielimallit viittaavat sivustoon.

Miksi se kannattaa silti tehdä? Tiedosto on halpa toteuttaa, siitä on konkreettista hyötyä agenteille ja kehittäjätyökaluille, ja ennen kaikkea **sen kirjoittaminen pakottaa tiivistämään oman palvelun ydinfaktat**: mitä teet, kenelle, mihin hintaan ja millä ehdoilla. Juuri tätä samaa tietoa tekoäly etsii myös tavallisilta sivuiltasi.

---

### 4. Tiedostojen roolit: llms.txt vs. llms-full.txt

Molemmat tiedostot sijoitetaan **verkkosivuston juureen**, samaan paikkaan kuin tutut `robots.txt` ja `sitemap.xml`. Kanavoissa ne näyttävät tältä:

```
https://kanavoi.com/robots.txt               <- Crawlerien pääsysäännöt
https://kanavoi.com/sitemap.xml              <- Hakukoneiden sivukartta
https://kanavoi.com/llms.txt                 <- Tekoälyn sisällysluettelo
https://kanavoi.com/llms-full.txt            <- Keskeinen sisältö yhtenä tiedostona
https://kanavoi.com/ohjeet/utm-parametrit.md <- Yksittäinen ohje Markdownina
```

#### llms.txt: sivuston sisällysluettelo

`llms.txt` on kevyt tiedosto, joka antaa kielimallille nopean yleiskuvan:

- sivuston tai palvelun lyhyen ydinkuvauksen
- kuratoidun linkkilistan sivuston tärkeimpiin osioihin
- lyhyen kuvauksen kunkin linkin takaa löytyvästä sisällöstä.

**Kenelle:** Agenteille, joiden konteksti-ikkuna on rajallinen tai jotka etsivät täsmävastausta yhteen kysymykseen, kuten yhteen API-päätepisteeseen tai palautusehtoon.

#### llms-full.txt: kaikki keskeinen tieto kerralla

`llms-full.txt` kokoaa sivuston keskeisen asiasisällön yhdeksi pitkäksi Markdown-dokumentiksi. Se ei kuulu alkuperäiseen ehdotukseen, jossa vastaavaa tarkoitusta palvelevat työkalulla generoitavat `llms-ctx.txt` ja `llms-ctx-full.txt`. `llms-full.txt` on dokumentaatioalustojen, kuten Mintlifyn, yleistämä käytäntö, josta on tullut käytännön standardi.

Nykyisten kärkimallien, kuten Anthropicin Claude Opus 5.5:n sekä OpenAI:n ja Googlen uusimpien mallien, konteksti-ikkunaan mahtuu satoja tuhansia tokeneita, osaan jopa miljoona. Kun mallille tarjotaan `llms-full.txt`:

1. Sen ei tarvitse tehdä kymmeniä erillisiä HTTP-pyyntöjä sivu kerrallaan.
2. Se voi ladata koko tuotekatalogin tai dokumentaation kerralla työmuistiinsa.
3. Se voi tehdä ristiviittauksia sivujen välillä ilman riskiä siitä, että jokin alasivu jää lukematta.

Iso konteksti ei kuitenkaan ole ilmainen: jokainen token maksaa, ja hyvin pitkissä konteksteissa mallien tarkkuus heikkenee. Siksi `llms-full.txt` kannattaa pitää rajattuna olennaiseen.

#### robots.txt päättää edelleen pääsystä

`llms.txt` ei anna eikä estä pääsyä. Sen tekee `robots.txt`. Tekoälyyhtiöillä on nykyään useita eri crawlereita eri tarkoituksiin: esimerkiksi OpenAI:n `GPTBot` (koulutus), `OAI-SearchBot` (ChatGPT:n haku) ja `ChatGPT-User` (käyttäjän pyytämä haku) sekä Anthropicin `ClaudeBot`, `Claude-SearchBot` ja `Claude-User`. Googlen `Google-Extended`-tunnisteella voi kieltää sisällön käytön Gemini-mallien kehittämisessä ilman, että se vaikuttaa Google-hakuun.

Kanavoin `robots.txt` on tällä hetkellä yksinkertainen: kaikki crawlerit saavat lukea julkiset sivut. Kanavoi haluaa tulla löydetyksi tekoälyhauissa, joten se on tietoinen valinta. Tärkeintä on, ettei hakubotteja estetä vahingossa, jos haluaa näkyä tekoälyn vastauksissa.

---

### 5. Tiedostojen anatomia: Kanavoin llms.txt

Formaatti on tavallista Markdownia. Rakenne noudattaa llmstxt.org-ehdotuksen järjestystä:

1. **H1-otsikko (`#`):** Projektin tai sivuston nimi. Tämä on ainoa pakollinen osa.
2. **Lainauslohko (`>`):** Lyhyt tiivistelmä, joka sisältää tiedoston ymmärtämisen kannalta tärkeimmät tiedot.
3. **Vapaamuotoiset kappaleet tai listat:** Lisätietoa sivustosta (ei otsikoita).
4. **H2-osiot (`##`):** Linkkilistoja muodossa `- [Otsikko](url): Valinnainen kuvaus`.
5. **`## Optional`-osio:** Toissijainen aineisto, jonka agentti voi ohittaa, jos konteksti on tiukalla.

Ehdotus suosittelee lisäksi, että jokaisesta sivusta olisi tarjolla Markdown-versio samassa osoitteessa `.md`-päätteellä. Kanavoissa jokainen ohje on saatavilla näin, esimerkiksi [kanavoi.com/ohjeet/utm-parametrit.md](https://kanavoi.com/ohjeet/utm-parametrit.md), ja `llms.txt` linkittää suoraan näihin versioihin.

#### Kanavoin [llms.txt](https://kanavoi.com/llms.txt)

```markdown
# Kanavoi

> Kanavoi on suomalainen UTM-työkalu markkinointitiimeille ja -toimistoille. Se pitää UTM-linkkien nimeämisen yhtenäisenä ja näyttää jo ennen julkaisua, missä Google Analytics 4:n kanavassa (esim. Paid Social tai Email) liikenne raportoidaan.

Ilmaiset työkalut toimivat ilman kirjautumista. Palvelin on Helsingissä, ja varmuuskopiot ovat EU:ssa. Ei seurantaevästeitä. Sivuston kävijämäärät lasketaan ilman evästeitä.

## Ilmaiset työkalut

- [UTM-generaattori](https://kanavoi.com/#generaattori): Rakentaa UTM-linkin, siistii arvot (esim. Kevät Ale → kevat_ale) ja näyttää ennen julkaisua, missä GA4:n kanavassa liikenne raportoidaan.
- [GA4-kanavatarkistin](https://kanavoi.com/ga4-kanavatarkistin): Tarkista ilmaiseksi, mihin Google Analytics 4:n kanavaan utm_source- ja utm_medium-arvot ohjaavat liikenteen, ja korjaa Unassigned-liikenne.
- [QR-koodi](https://kanavoi.com/qr-koodi): Tee ilmainen QR-koodi linkille. Lisää UTM-tunnisteet ja lataa koodi SVG-muodossa painoon tai PNG-kuvana. Ei rekisteröitymistä.

## Ohjeet

- [UTM-parametrit: kattava opas suomeksi](https://kanavoi.com/ohjeet/utm-parametrit.md): Mitä viisi UTM-parametria tarkoittavat, mitkä niistä ovat välttämättömiä ja miten nimeät ne niin, että GA4:n raportit pysyvät siisteinä.
- [Miksi GA4 näyttää liikenteen Unassigned-kanavassa ja miten sen korjaa](https://kanavoi.com/ohjeet/ga4-unassigned.md): Unassigned tarkoittaa, ettei mikään GA4:n kanavasääntö täsmännyt. Näin löydät syyn ja korjaat sekä linkit että raportit.
- [Google Ads ja UTM: tarvitaanko niitä gclidin rinnalla?](https://kanavoi.com/ohjeet/google-ads-utm.md): Automaattinen merkintä (gclid) riittää GA4:lle. Näin päätät, tarvitaanko UTM-tunnisteita lisäksi, ja lisäät ne sotkematta raportteja.
- [UTM-tagit Meta-mainoksiin oikein](https://kanavoi.com/ohjeet/utm-meta-mainokset.md): Facebook- ja Instagram-mainosten UTM-tunnisteet: oikea lähde ja media, dynaamiset parametrit ja miksi lähde ei saa olla meta.
- [UTM-tagit uutiskirjeisiin](https://kanavoi.com/ohjeet/utm-uutiskirjeet.md): Uutiskirjeen linkkien UTM-tunnisteet: media email, kampanja jokaiselle lähetykselle ja sisältö jokaiselle linkille. Toimii kaikissa postituspalveluissa.

## Hinnat

- [Hinnat](https://kanavoi.com/#hinnat): Ilmainen 0 €, Tiimi 14 €/kk, Toimisto 49 €/kk. Hinnat ilman arvonlisäveroa.
- [Odotuslista](https://kanavoi.com/#odotuslista): Tilit avataan odotuslistalla oleville. Paketin voi valita liittyessä.

## Optional

- [Koko sisältö yhtenä tiedostona](https://kanavoi.com/llms-full.txt): Palvelun kuvaus, hinnat, usein kysytyt kysymykset ja kaikki ohjeet.
- [Tietosuojaseloste](https://kanavoi.com/tietosuoja): Mitä henkilötietoja Kanavoi käsittelee, miksi ja kuinka kauan.
- [Käyttöehdot](https://kanavoi.com/kayttoehdot): Kanavoi-palvelun käyttöehdot yritysasiakkaille.
```

Huomaa, että hinnat ja tietosuojalupaukset on kirjoitettu auki eikä vain linkitetty. Agentti, joka lukee pelkän `llms.txt`:n, saa silti oikeat perusfaktat.

#### Kanavoin [llms-full.txt](https://kanavoi.com/llms-full.txt)

`llms-full.txt` kokoaa saman tiedon varsinaisena tekstinä yhteen tiedostoon (noin 30 kt) ja erottaa osiot vaakaviivoilla (`---`). Siihen kuuluvat palvelun kuvaus, toimintaperiaate, hinnat, tietosuojalupaukset, usein kysytyt kysymykset ja kaikki ohjeet. Tiedoston alku:

````markdown
# Kanavoi: palvelun kuvaus, hinnat ja ohjeet

> Kanavoi on suomalainen UTM-työkalu markkinointitiimeille ja -toimistoille. Se pitää UTM-linkkien nimeämisen yhtenäisenä ja näyttää jo ennen julkaisua, missä Google Analytics 4:n kanavassa (esim. Paid Social tai Email) liikenne raportoidaan.

Tämä tiedosto kokoaa Kanavoin keskeisen sisällön yhdeksi tekstiksi kielimalleille. Ajantasaiset tiedot: https://kanavoi.com/

## Näin Kanavoi toimii

1. **Sovitte säännöt kerran.** Sallitut lähteet ja mediat sekä kampanjanimen muoto. Valmis pohja noudattaa GA4:n sääntöjä, joten aloittaminen on nopeaa.
2. **Jokainen tekee linkit samalla lomakkeella.** Väärää arvoa ei voi valita, ja kirjoitusasu siistitään automaattisesti: Kevät Ale → kevat_ale.
3. **Näet tuloksen ennen julkaisua.** Linkin vieressä näkyy, missä GA4:n kanavassa liikenne raportoidaan, ja varoitus korjausehdotuksineen, jos se päätyisi väärään paikkaan.

## Hinnat

Tilit avataan pian. Nämä ovat hinnat avautumisesta alkaen: kiinteä kuukausihinta koko tiimille, ei käyttäjäkohtaisia maksuja. Liity odotuslistalle, niin saat kutsun ensimmäisten joukossa. Hinnat ilman arvonlisäveroa.

- **Ilmainen, 0 €**
  - Generaattori, GA4-kanavan tarkistus ja QR-koodit
  - Yksi käyttäjä
  - 50 viimeisintä linkkiä tallessa
- **Tiimi, 14 €/kk**
  - Rajaton määrä käyttäjiä ja linkkejä
  - Nimeämissäännöt ja kampanjanimen tarkistus
  - Monta kanavaa kerralla, Meta- ja Google Ads -muodot
  - Linkin kunnon tarkistus, API
...
````

Kaksi yksityiskohtaa ovat tärkeämpiä kuin miltä ne näyttävät. Ensinnäkin hinnoista kerrotaan myös ehto: ne ovat voimassa tilien avautumisesta alkaen. Jos tämä puuttuisi, agentti voisi kertoa käyttäjälle, että palvelun voi ostaa heti. Toiseksi ohjeiden täsmätieto säilyy. Verkkosivulla GA4-kanava näytetään värillisenä laatikkona, mutta Markdown-versiossa se on tavallinen virke:

```markdown
GA4-kanava, kun lähde on uutiskirje ja media newsletter: **Unassigned**. Mikään GA4:n oletuskanavaryhmän sääntö ei täsmää, joten liikenne näkyy raporteissa kanavattomana.
```

Juuri tällaista tietoa kielimalli ei osaisi päätellä markkinointisivun mainoslauseista.

---

### 6. Miten tiedostot toteutetaan teknisesti?

Toteutustapa riippuu sivuston arkkitehtuurista.

#### Vaihtoehto A: Staattiset sivustot (Astro, Next.js, Hugo, VitePress)

Jos sisältö on jo Markdown-tiedostoina, buildin yhteydessä voi ajaa koontiskriptin:

```javascript
// scripts/generate-llms.mjs
import fs from "node:fs";
import path from "node:path";

const DOCS_DIR = "./content/docs";
const OUTPUT_FULL = "./public/llms-full.txt";

const files = fs
  .readdirSync(DOCS_DIR)
  .filter((file) => file.endsWith(".md"))
  .sort();

let fullContent = "# Sivuston nimi: täysi dokumentaatio\n\n";

for (const file of files) {
  const content = fs.readFileSync(path.join(DOCS_DIR, file), "utf-8");
  // Poista front matter, jotta tiedostoon jää vain asiasisältö.
  const body = content.replace(/^---\n[\s\S]*?\n---\n/, "");
  fullContent += `\n---\n\n${body.trim()}\n`;
}

fs.writeFileSync(OUTPUT_FULL, fullContent);
console.log("llms-full.txt generoitu.");
```

#### Vaihtoehto B: Dynaaminen reitti (Next.js App Router, WordPress, Shopify)

Jos sisältö on tietokannassa, palvelinpuolen reitti voi muuntaa sen Markdowniksi ja palauttaa sen pelkkänä tekstinä. Next.js:ssä kansio voi olla nimeltään `llms.txt`:

```typescript
// app/llms.txt/route.ts (Next.js App Router)
export const dynamic = "force-static";

export function GET() {
  const markdown = [
    "# Yritys Oy",
    "",
    "> Yrityksen yhden virkkeen kuvaus.",
    "",
    "## Sisältö",
    "- [Tietoa meistä](https://yritys.fi/tietoa): Profiili ja arvot.",
  ].join("\n");

  return new Response(markdown, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Cache-Control": "public, max-age=86400",
    },
  });
}
```

#### Vaihtoehto C: Go-palvelin (näin Kanavoissa)

Kanavoi on Go-monoliitti, jonka ohjeartikkelit ovat Markdown-tiedostoina binääriin upotettuina. Kaikki kolme tiedostotyyppiä kootaan käynnistyksen yhteydessä samoista lähteistä kuin HTML-sivut: käännöstiedoston teksteistä, hinnoista ja ohjeista. Sisältö ei siis pääse koskaan eriytymään, ja uusi ohje päätyy tiedostoihin itsestään.

Yksittäisen ohjeen Markdown-versio tarjoillaan näin:

```go
// handleDocsMarkdown serves /ohjeet/<slug>.md. The canonical link header keeps
// search engines on the HTML page.
func (a *App) handleDocsMarkdown(w http.ResponseWriter, r *http.Request, slug string) {
	art, ok := a.docs.Get(slug)
	if !ok {
		a.handleNotFound(w, r)
		return
	}
	w.Header().Set("Link", `<`+a.SiteURL+art.Path()+`>; rel="canonical"`)
	writeMarkdown(w, []byte(a.articleMarkdown(art)))
}
```

`Link: rel="canonical"` -otsake kannattaa muistaa. Ilman sitä hakukone voi pitää Markdown-versiota saman sisällön kaksoiskappaleena ja näyttää sen hakutuloksissa HTML-sivun sijaan. Lisäksi HTML-sivun `<head>`-osa kertoo Markdown-version osoitteen: `<link rel="alternate" type="text/markdown" href="…/utm-parametrit.md">`.

Kanavoin artikkeleissa on muutama oma komponentti (esimerkkilinkit, GA4-kanavat, vinkkilaatikot ja "Kokeile itse" -kortit). Koontivaiheessa ne muunnetaan tavalliseksi Markdowniksi, ja sivuston sisäiset linkit muutetaan absoluuttisiksi, jotta ne toimivat myös sivuston ulkopuolella. Muuten tiedostoon päätyisi merkintöjä, joita kielimalli ei ymmärrä.

> **Tekninen vinkki:** Varmista, että palvelin palauttaa tiedostoille HTTP-otsakkeen `Content-Type: text/plain; charset=utf-8` tai `Content-Type: text/markdown; charset=utf-8`. Silloin lukija tietää heti, ettei kyse ole selaimessa renderöitävästä HTML-sivusta.

---

### 7. Agenttisen kaupankäynnin standardit: UCP ja ostaminen keskustelussa

Dokumentaation lisäksi agenttinen murros koskettaa erityisen voimakkaasti verkkokauppaa. Microsoftin oppaan keskeinen viesti on, että pelkkä löydettävyys ei riitä: **ostopolku tiivistyy keskustelun sisään**.

#### Universal Commerce Protocol (UCP)

Universal Commerce Protocol on Googlen ja Shopifyn tammikuussa 2026 esittelemä avoin standardi, jonka kehitykseen osallistuvat myös esimerkiksi Etsy, Wayfair, Target ja Walmart. Sen avulla tekoälyassistentit ja kaupankäyntialustat keskustelevat keskenään ilman jokaiselle botille erikseen rakennettuja integraatioita. Myös Microsoft tukee UCP:tä Merchant Centerissä. Microsoftin oppaan mukaan UCP rakentuu olemassa olevan tuotesyötteen päälle ja rikastaa sitä agenteille tärkeillä tiedoilla:

- palautusehdot ja asiakastuen yhteystiedot (kaupan yleisasetuksissa)
- tuotekohtaiset varoitus- ja sääntelymerkinnät (`consumer_notice_type` ja `consumer_notice_message`), jotka ovat pakollisia esimerkiksi leluille ja säännellyille tuotteille
- yksilöllinen tuotetunniste (`merchant_item_id`), jolla tuote voidaan tunnistaa eri syötteissä ja järjestelmissä.

Perustana ovat edelleen tuotesyötteen perustiedot: tarkat otsikot, kuvaukset, GTIN/MPN-koodit, reaaliaikainen saatavuus, hinta ja tuotekategoria. Oppaan mukaan vanhentuneet hinnat ja varastotilanteet ovat yleisimmät syyt siihen, että tuotteet katoavat näkyvistä. Syöte kannattaa siksi päivittää vähintään kerran päivässä.

#### Ostaminen keskustelussa (Agentic Checkout)

Kun käyttäjä pyytää tekoälyä hoitamaan ostoksen puolestaan ("Do it for me"), tilaus tehdään poistumatta keskustelusta. Microsoftin **Copilot Checkout** toimii Copilot.comissa ja mobiilisovelluksessa. Se on käytettävissä Shopifyn (oikeutetut kauppiaat liitetään automaattisesti), PayPalin ja Stripen kautta. Yritys pysyy kauppiaana (*merchant of record*), ja maksut, toimitus sekä asiakassuhde pysyvät sen omissa järjestelmissä. OpenAI on vastaavasti rakentanut Stripen kanssa avoimen Agentic Commerce Protocolin (ACP) ChatGPT:n ostoksia varten.

**Suomalaisen kauppiaan kannalta tärkeä huomio:** sekä Microsoftin UCP-tuki että Copilot Checkout ovat toistaiseksi käytössä vain Yhdysvalloissa, ja Copilot Checkout on vielä pilottivaiheessa. Tuotedatan laatuun kannattaa silti panostaa jo nyt, koska siitä hyötyy kaikissa kanavissa.

#### Entä jos et myy tuotteita?

Kanavoi on SaaS-palvelu, ei verkkokauppa, joten tuotesyötettä sillä ei ole. Palvelulle agenttisen verkon vastine on rajapinta, jota agentti voi käyttää itse. Kanavoin tiekartalla on REST API ja MCP-palvelin (Model Context Protocol), jossa on työkalut `create_utm_link`, `list_allowed_values` ja `classify_ga4_channel`. Ajatus on, että tekoälyagentti voi luoda kampanjalinkit tiimin sääntöjen mukaan, mutta se ei voi keksiä uusia arvoja lukittuihin kenttiin. Agentti saa siis tehdä työn, mutta tiimin säännöt pitävät.

---

### 8. Kuinka tekoälynäkyvyyttä mitataan?

Perinteinen analytiikka kertoo, mitä sivustolla tapahtui, mutta harvoin sitä, **miksi tekoäly suositteli kilpailijaa**. Mittaamiseen on tullut vuoden 2026 aikana kaksi merkittävää uudistusta.

#### GA4:n oma AI Assistant -kanava

Google lisäsi toukokuussa 2026 GA4:n oletuskanavaryhmään oman **AI Assistant** -kanavan. Kun istunnon viittaaja tunnistetaan tekoälyavustajaksi (esimerkiksi ChatGPT, Gemini tai Claude), GA4 asettaa medium-arvoksi `ai-assistant` ja raportoi istunnon omana kanavanaan. Kolme asiaa kannattaa tietää:

- Muutos ei koske vanhaa dataa, joten vertailukelpoista historiaa ei ole.
- Jos tekoälysovellus ei välitä viittaajatietoa, käynti näkyy edelleen Direct-kanavassa.
- Omat suodattimet ja kanavaryhmät kannattaa tarkistaa. Tekoälystä tullut liikenne, joka ennen näkyi Referral-kanavassa, näkyy nyt omana kanavanaan medium-arvolla `ai-assistant`.

Kanavoin [GA4-kanavatarkistin](https://kanavoi.com/ga4-kanavatarkistin) tuntee jo tämän säännön. Se kertoo, että `chatgpt.com / ai-assistant` päätyy AI Assistant -kanavaan, mutta `chatgpt.com / referral` Referral-kanavaan. Kun tiimi tekee linkkejä esimerkiksi tekoälytyökalujen hakemistoihin tai kumppaneille, tarkistimesta näkee etukäteen, mihin raportin riviin liikenne päätyy.

#### Microsoft Clarityn AI Visibility -näkymät

Microsoftin oppaan mukaan Microsoft Clarityn tekoälynäkyvyyden työkaluissa seurataan kolmea asiaa:

1. **Löydettävyys (Bot activity, Citations):** Mitkä sivut tekoälybotit hakevat, kuinka usein sisältöön viitataan ja millä hakukyselyillä (*grounding queries*) se nousee vastauksiin.
2. **Näkyvyys (Topic Insights):** Missä aiheissa kilpailijoihin viitataan mutta sinuun ei (*content gaps*), sekä priorisoidut toimenpidesuositukset.
3. **Käyttäytyminen saapumisen jälkeen (AI-referred traffic):** Tekoälystä tulevat kävijät ovat usein vertailun loppusuoralla. Jos he poistuvat ostamatta, syy on usein puuttuva täsmätieto, kuten kokotaulukko tai toimitusaika. Heatmapit ja istuntotallenteet näyttävät, mitä kävijä etsi.

Clarity on ilmainen eikä korvaa GA4:ää, vaan täydentää sitä. Tietosuojasyistä kannattaa huomata, että istuntotallenteet ja heatmapit vaativat seurantakoodin ja yleensä myös evästesuostumuksen. Kanavoi on itse valinnut toisin: sen kävijätilasto lasketaan palvelimella ilman evästeitä.

---

### 9. Yhdistelmästrategia: Näin saavutat parhaan tekoälynäkyvyyden

`llms.txt` ei toimi tyhjiössä, vaan se on yksi kerros nelitasoisessa kokonaisuudessa:

```
+--------------------------------------------------------------+
|                     1. PERINTEINEN SEO                       |
|  - Sivuston tekninen nopeus, Core Web Vitals ja indeksointi  |
|  - Laadukkaat takaisinlinkit ja brändimaininnat verkossa     |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                    2. RAKENTEINEN DATA                       |
|  - Schema.org (JSON-LD): Organization, Product, Offer,       |
|    Article, FAQPage                                          |
|  - Auttaa hakukoneita ja tekoälyjä tunnistamaan entiteetit   |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|                   3. ANSWER-FIRST-SISÄLTÖ                    |
|  - Selkeät määritelmät, vertailutaulukot ja tiivistelmät     |
|  - Vastaus käyttäjän kysymykseen heti sivun alussa           |
+--------------------------------------------------------------+
                               |
                               v
+--------------------------------------------------------------+
|        4. LLMS.TXT, LLMS-FULL.TXT JA TUOTESYÖTE (UCP)        |
|  - Karsii HTML- ja skriptikohinan                            |
|  - Tarjoaa agenteille halvan ja suoran pääsyn faktoihin      |
+--------------------------------------------------------------+
```

#### Kanavoin tilanne

Rakennan Kanavoita julkisesti, joten tässä tilannekatsaus:

- ✅ **Tekninen perusta:** Kevyt Go- ja HTMX-sivusto ilman raskasta JavaScriptiä, `robots.txt` ja dynaaminen `sitemap.xml`.
- ✅ **Rakenteinen data ohjeissa:** Jokaisella ohjeartikkelilla on schema.org-merkinnät `Article` ja `BreadcrumbList`.
- ✅ **Answer-first-sisältö:** Ohjeet vastaavat suoraan yhteen kysymykseen, kuten "Miksi GA4 näyttää liikenteen Unassigned-kanavassa?".
- ✅ **Rakenteinen data etusivulla:** `Organization`, `WebSite`, `SoftwareApplication` (paketit `Offer`-tietoina) ja `FAQPage`.
- ✅ **llms.txt, llms-full.txt ja ohjeiden `.md`-versiot**, kaikki tuotettuina samoista lähteistä kuin sivut.

Kun muutokset olivat tuotannossa, ajoin etusivun Googlen [Rich Results Testin](https://search.google.com/test/rich-results) läpi. Se hyväksyi organisaation ja ohjelmistosovelluksen, ja testistä opin kolme asiaa:

1. **Googlen hyväksymä ei ole sama kuin schema.org:n sallima.** Ilmoitin hinnat ensin tarkemmalla `UnitPriceSpecification`-rakenteella (kuukausihinta, ALV 0). Schema.org sallii sen, mutta Googlen testi merkitsi sen ohjelmistosovelluksen tarjouksessa virheelliseksi. Hinta ja valuutta riittävät, joten kuukausijakso ja ALV-tieto siirtyivät tarjouksen kuvaukseen.
2. **Puuttuvaa arvosanaa ei korjata keksimällä.** Testi huomauttaa yhä, että `aggregateRating` puuttuu. Kanavoilla ei vielä ole aitoja asiakasarvioita, ja keksityt tähdet rikkoisivat Googlen ohjeita. Arvosana lisätään, kun arvioita on.
3. **Kaikkea ei näytetä hakutuloksissa, eikä tarvitsekaan.** `FAQPage` ei näy testissä lainkaan, koska Google näyttää usein kysytyt kysymykset hakutuloksissa vain viranomais- ja terveyssivustoille. Merkintä on silti hyödyllinen muille lukijoille, kuten tekoälyhauille.

#### Mitä tiedostoihin EI pidä laittaa?

1. **Evästekäytäntöjä tai yleistä lakijargonia:** Älä täytä mallin konteksti-ikkunaa toistuvilla sopimusehdoilla. Linkki `Optional`-osiossa riittää.
2. **Henkilötietoja tai sisäisiä avaimia:** Vaikka tiedosto on tarkoitettu koneille, se on julkinen verkkotiedosto.
3. **Heikkolaatuista arkistosisältöä:** Ota mukaan vain sisältöä, joka tuo lukijalle aidosti uutta tietoa (*information gain*).
4. **Tietoa, joka ei vastaa sivustoa:** Jos `llms.txt` lupaa eri hinnan kuin hinnastosivu, tekoäly ei tiedä kumpaan uskoa. Tuota tiedostot samasta lähteestä kuin sivut.

---

### 10. Yhteenveto

Siirtymä ihmisten verkosta kohti agenttista verkkoa ei tarkoita, että perinteinen verkkosivusto kuolisi, mutta sen rooli muuttuu. Verkkosivustosi ei ole enää vain näyteikkuna ihmisille, vaan myös tietolähde ja rajapinta tekoälyille.

`llms.txt` ei ole oikotie tekoälyhakujen kärkeen, eikä kukaan suurista toimijoista ole luvannut lukea sitä. Se on kuitenkin halpa ja hyödyllinen osa kokonaisuutta, jonka varsinainen ydin on tämä:

- **Ydinfaktat selkeästi ja yhdessä paikassa:** Mitä teet, kenelle, mihin hintaan ja millä ehdoilla.
- **Rakenteinen data ja ajantasainen tuotesyöte:** Agentti voi vahvistaa tiedot arvailematta, mikä vähentää virhetulkintoja.
- **Mittaaminen:** GA4:n AI Assistant -kanava ja tekoälynäkyvyyden työkalut kertovat, toimiiko tämä.

Organisaatiot, jotka laittavat koneluettavan perustansa kuntoon nyt, ovat valmiina, kun agenttinen kaupankäynti laajenee Yhdysvalloista muualle.

---

**Teetkö itse kampanjoita?** Jos haluat, että kampanjasi näkyvät GA4:ssä yhtenä rivinä oikeassa kanavassa, oli liikenne sitten maksettua somea, uutiskirjettä tai tekoälyavustajista, kokeile [Kanavoin ilmaista UTM-generaattoria](https://kanavoi.com/#generaattori) ja [GA4-kanavatarkistinta](https://kanavoi.com/ga4-kanavatarkistin). Ne toimivat ilman kirjautumista. Tiimi- ja Toimisto-paketteihin pääset mukaan [odotuslistan](https://kanavoi.com/#odotuslista) kautta, ja 50 ensimmäistä maksavaa asiakasta saa hinnasta pysyvän 50 %:n alennuksen. Ja jos lähetät oman tekoälyagenttisi tutustumaan palveluun, ohjaa se osoitteeseen [kanavoi.com/llms.txt](https://kanavoi.com/llms.txt).

#### Lähteet

- Microsoft Advertising: [Prepare for the Age of AI Shopping](https://about.ads.microsoft.com/content/dam/sites/msa-about/global/common/content-lib/pdf/prepare-for-the-age-of-ai-shop.pdf) (2026)
- Jeremy Howard / Answer.AI: [The /llms.txt file](https://llmstxt.org) (2024)
- Aggarwal ym.: [GEO: Generative Engine Optimization](https://arxiv.org/abs/2311.09735) (2023, KDD 2024)
