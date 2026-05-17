---
title: "Suomalaiset jupisevat korpimajoissaan, kun Piilaakso koodaa agenttien avulla ja on jo ratkaissut ymmärrysvelan ongelman"
description: "VTT:n tutkijoiden blogi väittää, että agenttien käyttö aiheuttaa ymmärrysvelkaa. Minä väitän päinvastaista."
publishedAt: "2026-05-17"
slug: "vastaus-vtt-tutkijoiden-blogiin"
draft: false
tags: ["koodaus", "ai", "tekoäly", "agentit", "automaatio"]
---

VTT:n tutkijat Maaria Nuutinen ja Arto Wallin nostivat hiljattain esiin kriittisen huolen "ymmärrysvelasta" (comprehension debt) agenttipohjaisessa ohjelmistokehityksessä. Huoli on validi: riski siitä, että tekoälyagentit tuottavat koodia nopeammin kuin tiimit ehtivät sitä sisäistää, on todellinen. Olen osittain samaa mieltä tästä, mutta mielestäni tutkijat ovat jättäneet tärkeitä näkökulmia huomiotta.  

VTT:n blogi kuvaa mielestäni pitkälti niin kutsuttua "vibe codingia", ei harkittua ja ammattimaista agenttikehitystä. Tietenkin ihmiset koodaavat nyt vibeltämällä ja ymmärtämättä riviäkään koodia. Se on osa oppimispolkua ja nostaa miljoonien ihmisten ymmärrystä ohjelmistojen rakentamisesta, jota he eivät ilman laajojen kielimallien kyvykkyyksiä olisi koskaan edes halunneet alkaa opetella. Ymmärrysvelka ei ole mielestäni tekoälyn aiheuttama ongelma. Se on vanha ongelma koodauksessa. Siihen tekoäly on nimenomaan ratkaisu. 

Ohjelmistoalalla tunnetaan useita käsitteitä, jotka kuvaavat tätä ilmiötä jo vuosikymmenten takaa:

- **"Programming by Coincidence"** (Hunt & Thomas, *The Pragmatic Programmer*, 1999): kehittäjät kirjoittavat koodia, joka sattuu toimimaan, ilman että ymmärtävät miksi. Kirja vertaa tilannetta sotilaaseen, joka kävelee läpi miinakentän — ensimmäiset askeleet eivät räjähtäneet, joten hän päättää, että kenttä on turvallinen.

- **"Copy and paste is a design error"** (Steve McConnell, *Code Complete*, 2004): koodin monistaminen ilman ymmärrystä on suunnitteluvirhe.

Empiirinen data vahvistaa riskin: Fischer ym. (*IEEE Symposium on Security and Privacy*, 2017) analysoivat 1,3 miljoonaa Android-sovellusta ja havaitsivat, että 15,4 % niistä sisälsi Stack Overflow -koodinpätkiä. Näistä snippetteistä **97,9 %** sisälsi vähintään yhden tietoturvahaavoittuvan koodinpätkän.

Ymmärrysvelka on ollut olemassa ennen tekoälyä. Koodin kopioiminen ilman ymmärrystä on ollut riski jo kauan. Ero on, että nyt prosessi on nopeampi. Nyt se on kuitenkin helpommin havaittavissa, hallittavissa ja korjattavissa, kun hyödynnämme tekoälyagentteja ja käytetämme oikeanlaista työnkulkua niiden kanssa.

Olen rakentanut kymmeniä sovelluksia hyödyntäen agenttipohjaista koodausta kymmenillä eri kielimalleilla. Se on luonut minulle ymmärryksen rakenteellisesta ja vaiheistetusta työnkulusta, joka säilyttää minun ymmärryksen prosessin jokaisessa vaiheessa, jos niin haluan. Haluan vielä tähdentää, että ohjelmointi ei ole minun leipätyöni. Olen oppinut itsenäisesti tekemällä, käymällä avoimia ohjelmointikursseja, syventämällä taitojani kysymällä generatiiviselta tekoälyltä, sekä ehdottomasti tärkeimpänä: agenttipohjaisen koodauksen kautta. 

## 1. Kaikki koodi ei ole samanarvoista

Ennen kuin alan kirjoittamaan koodia, teen nykyään arvion projektin eri osa-alueista: mikä on minkäkin tehtävän kriittisyys? VTT:n tutkijoiden pelko ymmärryksen katoamisesta pitäisi mielestäni olla suoraan verrannollinen siihen, kuinka vakavat seuraukset virheellä voi olla. Tämä seuraava riskianalyysi ei ole mikään validoitu tai tutkittu tapa toteuttaa riskiarviota, mutta kuvaa karkeasti omaa tapaani lähestyä asiaa.

**Taso 1:** Avustettu luovuus (Matala riski)

- Esimerkki: Staattiset sivut, tyylittely (CSS), markkinointisisällöt, mockupit.
- Agentin rooli: Autonominen tekijä. Saa generoida kokonaisia tiedostoja.
- Ihmisen rooli: Kuraattori ja varmistaa, että esimerkiksi tyylittely on toteutettu siten, että se on helposti muokattavissa myöhemmin.

Autonomisuus: Agentilla on 90 % valta toteuttaa muutokset.

**Taso 2:** Operatiivinen tuki (Keskisuuri riski)

- Esimerkki: Sisäiset työkalut, datan muunnokset, testien generointi.
- Agentin rooli: Toteuttaja. Kirjoittaa edelleen koodia hyvin itsenäisesti, mutta tarkka suunnitelma ohjaa sitä.
- Ihmisen rooli: Validoija. Tarkistaa logiikan ja suorittaa automaattiset testit.

Autonomisuus: Agentilla on 60 % valta; ihminen hyväksyy jokaisen pull requestin.

**Taso 3:** Strateginen komponentti (Korkea riski)

- Esimerkki: Ulkoiset integraatiot, monimutkainen business-logiikka, API-rajapinnat.
- Agentin rooli: Apulaissuunnittelija. Ei kirjoita koodia ennen kuin on selittänyt suunnitelman.
- Ihmisen rooli: Arkkitehti. Ohjaa agenttia vaihe kerrallaan ja vaatii EXPLANATION.md-tiedoston päivityksen jokaisesta muutoksesta. Tarkistaa koodin.

Autonomisuus: Agentilla on 30 % valta. Koodi syntyy tiiviissä dialogissa. Se on paloiteltu pienempiin kokonaisuuksiin ja se tarkistetaan ja testataan tiukasti.

**Taso 4:** Sensitiivinen ydin (Kriittinen riski)

- Esimerkki: Maksuliikenne, tietoturvakriittinen koodi, GDPR-arkaluonteinen data.
- Agentin rooli: Syväasiantuntija. Tuottaa ehdotuksia ja koodia, mutta joutuu perustelemaan jokaisen valinnan ja hylätyn vaihtoehdon.
- Ihmisen rooli: Kriittinen tarkastaja. Ei hyväksy koodia ennen kuin on haastanut agentin ("Miksi et käyttänyt kirjastoa X?", "Miten tämä käsittelee kilpailutilanteen?"). Ihminen vastaa järjestelmän logiikan hallinnasta.

Autonomisuus: Agentilla on 15 % valta. Se tekee raskaan työn, mutta ihminen pitää ohjat tiukasti käsissään jatkuvan dialogin ja koodin yksityiskohtaisen tarkastelun kautta.

## 2. Vibe Coding vs. järjestelmällinen agenttikoodauksen hyödyntäminen

VTT:n artikkeli hyvin pitkälle olettaa, että agenttipohjainen kehitys tarkoittaa: *"anna agentille tehtävä, paina enteriä, saa valmis koodi."* Tämä on vibe codingia — tekoälyn antamista suorittaa vapaana ilman kunnollista valvontaa. Tämä vibeltäminen kuuluu myös agenttipohjaisen koodauksen opiskeluun. Tätä minäkin toteutin aluksi, kun aloin käyttämään agenttipohjaista koodausta projekteissa. Katsoin kun agentit rakensivat sivuston muutamassa minuutissa ja lähes kaikki toimi kuten pitikin. Nälkäni tietenkin kasvoi ja kun testasin isompia kokonaisuuksia, niin kävi kuten oletettua... projekti ei edes rakentunut ja kaatui virheeseen heti ensimmäisessä mutkassa. Siitä aloinkin kehittämään omaa ymmärrystäni siitä, että mitkä ovat eri kielimallien rajat koodauksessa, mihin ne kykenevät, missä ne loistavat ja mikä on ihmisen eli minun itseni rooli tässä kokonaisuudessa. Nyt, noin 2 miljardia tokenia myöhemmin, ymmärrykseni tästä työnkulusta on aivan toinen kuin aloittaessani tämän parissa työskentelyn/harjoittelun.

Väitän, että työnkulkuni on nykyään vibeltämisen täysi vastakohta:

1. **Projekti- Arkkitehtitason suunnitelma** — Määritän projektin tavoitteet, kriittiset työnkulut, teknologia stackin, MVP-vaiheen valmiuden, turvallisuustarkistukset, selkeät vaiheet ja jokaiselle vaiheelle läpäisyvaatimukset. Käytän tässäkin vaiheessa tekoälyä apuna ja yleensä ensimmäisen vaiheen teen Gemini 3.1 pron kanssa keskustellen.
2. **Dialogi useamman mallin kanssa** — Käyn läpi lähestymistavan ja rakennan projektin vaiheistusta pidemmälle 2-3 kyvykkään mallin (Opus 4.6 tai 4.7 ja Codex 5.3) kanssa ennen kuin aloitan varsinaisen projektin.
3. **Suojakaiteet + AGENT.md** — Annan eksplisiittiset ohjeet, rajoitteet ja formaatit joka vaiheelle erikseen, ei kerran projektin alussa.
4. **Yhteenveto + katselmointi** — Agentti tiivistää jokaisen vaiheen tuloksen; minä katselmoin koodin ennen kuin annan "vihreää valoa" seuraavalle vaiheelle.
5. **Muistidokumentaatio** — Jokaisessa vaiheessa agentti tallentaa edistymisen, päätökset ja hylätyt vaihtoehdot MEMORY.md-tiedostoon.

Noudatan tätä melko orjallisesti merkittävissä projekteissa. Pienemmissä tai vain omaksi iloksi tehdyissä hommissa voin oikoa mutkia suoriksi.

Väitän, että tällä työnkululla ymmärrykseni koodista on korkeammalla tasolla kuin perinteisellä koodauksella ja dokumentaatio on huikeasti parempaa kuin siinä tapauksessa, että koodari on kirjoittanut hampaat irvessä pari riviä dokumentaatiota viikon päätteeksi. En usko, että dokumentaation kirjoittaminen on yhdenkään ammattidevaajan huippuhetkiä työviikossa, vaan pakollinen paha, jonka tärkeys ymmärretään, mutta joka on todella tylsää puuhaa. 


## 3. MEMORY.md — Organisaation jaetut aivot

Suurin lääke ymmärrysvelkaan on **jatkuvasti päivittyvä MEMORY.md**. Se tallentaa sen, mikä muuten katoaa, kun tiimin jäsen vaihtuu tai projekti jatkuu kuukausia myöhemmin.

MEMORY.md:ssä on kolme kriittistä osaa:

- **Projektin eteneminen:** Mitkä vaiheet projektista on suoritettu?
- **Konteksti:** Miksi valitsimme tämän lähestymistavan, tämän kirjaston, tämän arkkitehtuurin?
- **Vaihtoehdot:** Mitä muita tapoja harkittiin ja miksi ne hylättiin?
- **Rajoitteet:** Mitkä ovat järjestelmän tunnetut heikkoudet ja missä tilanteissa se voi mennä pieleen?

Tämä tiedosto on kultaa: jos uusi kehittäjä tulisi tiimiin tai kun palaan koodiin pitkän ajan kuluttua, historia ei ole kadonnut tekoälyn chatti-ikkunaan. Tämä on tapa, jolla voitan ymmärrysvelan — ei vähentämällä tekoälyn käyttöä, vaan rakentamalla sen käytölle selkeät säännöt ja ymmärtämällä sen mahdollisuudet ja heikkoudet. Muistitiedosto myös täydentyy siinä vaiheessa, kun ohjelmasta löytyy se tuotantoon päässyt bugi ja se korjataan - tekoälyn avustamana. Projektille luodaan tietenkin myös tekninen dokumentaatio, mutta se on korkeamman tason dokumentointi kuin projektin muisti.

## 6. FoSW-projektisivuston koodi on elävä varoitusmerkki tutkijoiden pelkäämästä ymmärrysvelasta

Tutkimusryhmällä on projektilleen sivusto: [futuresofsoftwarework.github.io/FoSW](https://futuresofsoftwarework.github.io/FoSW/).

Tarkastelin sivustoa ja sen lähdekoodia GitHub-repositoriosta — ja löysin aika paljon näyttöä siitä, että sivusto on rakennettu vibe-koodaamalla:

- **AI-generoidut commit-viestit:** `"feat: implement metrics dashboard brainstorming and design documentation with supporting assets"` — ihmisen kirjoittama viesti olisi tyypillisesti lyhyempi ja suorempi.
- **Tailwind-luokkien valtavat määrät:** Esim. yksittäinen `<h1>`-elementti sisältää yli 10 Tailwind-luokkaa, mukaan lukien arbitrary values (`drop-shadow-[0_0_15px_rgba(245,158,11,0.5)]`).
- **Inter-fontti:** klassinen AI-fontti, joka tulee AI:lta lähes oletuksena, kun rakennetaan jotain tech-tyyppistä sivustoa.
- **Custom "AI-värit":** `neon-gold`, `hologram-cyan`, `electric-blue`, `midnight` — klassisia tekoälykäyttöliittymien värejä.
- **Kehittäjä + Claude:** Kaikki commitit ovat Arton tai Arton ja Claude coden tekemiä.
- **Pelkkä silmäys riittää:** Sivusto on hyvin pitkälti juuri sen näköistä, mitä AI tuottaa. Eikä siinä mitään, jos se on tavoiteltua. Omilla sivuillani tämä ei haittaa, koska se ei sinänsä ole mikään virhe. En ole itsekään UI designer + haluan, että AI myös näkyy ja tuntuu omalla sivustollani.

Toisaalta nämä kaikki ovat ns. matalan riskin asioita ja projektin kuvauksessahan myös mainitaan, että kokeilut kuuluvat tähän tutkimusprojektiin. Sivusto ei kerää tietoa käyttäjistä, se ei kerää evästeitä, ei käsittele maksuliikennettä tai muutakaan arkaluontoista tietoa. Jos sivun on tarkoitus pysyä tällaisena matalan riskin projektina, niin sehän on ihan ok :) 

Ehkä läpinäkyvyyden vuoksi voisi AI-Signaleista mainita, että ne on tekoälyautomaation luomia .json-tiedostoja, jotka mahdollisesti sivuston ylläpitäjä on tarkistanut ja muuttanut niiden statuksen draftista --> published -muotoon tai ainakin hän on joutunut ne committoimaan, jolloin ne vasta deployautuu GitHub pagesiin. Julkisen rahoituksen projekteissa olisi mielestäni kohtuullista, että lukijalle kerrotaan tämä kaikki hyvin eksplisiittisesti heti sivuston alussa.

## Huomautus "tutkimusmenetelmästä"

Tämän blogin taustatyö — mukaan lukien FoSW-sivuston analyysi, lähdeviitteiden tarkistus — tehtiin hyödyntäen agentic AI -työkalua. Tällä kertaa käytin runsaasti OpenClaw:ta, sillä sen avulla sain suoraan haettua tietoa sivustoon liittyvästä GitHub repositoriosta ja lähdekoodista ilman, että minun itseni tarvitsi alkaa käymään kaikkea läpi, etsimään tietojen sijainta tai kahlaamaan julkista repoa alusta loppuun läpi.

Määritin kyselyt, tarkistin tulokset, ja rakentelin kokonaiskuvan vaihe vaiheelta yhdessä agentin kanssa. Tarkistin jokaisen lähteen alkuperäisestä materiaalista. AI teki virheitä matkan varrella, mutta keskustellen sen kanssa ja tarkistaen tulokset "käsin" nämä puutteet löydettiin. 

Ilman tekoälyä en olisi jaksanut — enkä ehtinyt — perehtyä näin perinpohjaisesti aiheeseen. Ymmärrykseni olisi jäänyt tästäkin asiasta vajaaksi. **Juuri tämä on pointtini:** kun tekoälyä käytetään rakenteellisessa työnkulussa, se ei vähennä ymmärrystä — se laajentaa sitä.

## Yhteenveto ja loppusanat

Tekoälyagentit eivät ole ongelma. Ongelma syntyy, jos niiden käyttöönotto nähdään vain **tuottavuushyppynä** eikä muutoksena organisaation (tai henkilön ja agenttien välisessä) osaamisjärjestelmässä.

Vastaus löytyy työnkulusta, joka luo ymmärrystä koodin mukana. Kriittisissä järjestelmissä meillä ei ole varaa pelkkään "vibe-koodaukseen". Me tarvitsemme rakenteellista agenttikehitystä, joka kasvattaa ihmisen ymmärrystä projektista. Vastuuta emme voi ulkoistaa tekoälylle tai koodin alkuperäiselle kirjoittajalle, jolta olemme sen kopioineet. 

Tekoäly pitää pystyä Suomessakin näkemään mahdollisuutena. Virheellistä ja haavoittuvuuksia sisältävää koodia on aina kopioitu projekteihin eikä kukaan tilaajan päässä ole kysynyt, että mistä tämä skripti on tullut ja pyytänyt selittämään sen juurta jaksaen. Ihminen ei todellakaan ole vastaus virheiden välttämiseen vaan uuden teknologian täysimittainen hyödyntäminen. 

Kritisoin vielä lyhyesti sitä, että Suomessa keskitytään mielestäni liikaa tekoälyn negatiivisiin asioihin ja ollaan lähtökohtaisesti kriittisiä sen sijaan, että voisimme tarkastella sen mahdollisuuksia. Ehkä me olemme Suomessa kuitenkin tällainen kehityksen takapajula, että kun emme saa itse kehitettyä skaalautuvaa teknologiaa, niin meidän rooliksemme jää mutista korpimajoissamme ja huudella taustalta tekoälyn haitallisuudesta ihmisille tai esimerkiksi ympäristölle.

What an amazing time to be alive! (Let's not ruin it)