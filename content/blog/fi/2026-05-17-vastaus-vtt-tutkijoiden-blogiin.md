---
title: "Suomalaiset jupisevat korpimajoissaan, kun Piilaakso koodaa agenttien avulla ja on jo ratkaissut ymmärrysvelan ongelman"
description: "VTT:n tutkijoiden blogi väittää, että agenttien käyttö aiheuttaa ymmärrysvelkaa. Minä väitän päinvastaista."
publishedAt: "2026-05-17"
slug: "vastaus-vtt-tutkijoiden-blogiin"
draft: true
tags: ["koodaus", "ai", "tekoäly", "agentit", "automaatio"]
---

VTT:n tutkijat Maaria Nuutinen ja Arto Wallin nostivat hiljattain esiin kriittisen huolen "ymmärrysvelasta" (comprehension debt) agenttipohjaisessa ohjelmistokehityksessä. Huoli on validi: riski siitä, että tekoälyagentit tuottavat koodia nopeammin kuin tiimit ehtivät sitä sisäistää, on todellinen. Olen osittain samaa mieltä tästä, mutta mielestäni tutkijat ovat jättäneet tärkeitä näkökulmia huomiotta. Ja itse asiassa heidän hahmottelemansa uudet työtavat ovat agenttikoodauksessa jo arkipäivää.

 Tietenkin ihmiset koodaavat tällä hetkellä paljon ymmärtämättä riviäkään koodia. Se on osa oppimispolkua ja nostaa miljoonien ihmisten ymmärrystä ohjelmistojen rakentamisesta, jota he eivät ilman laajojen kielimallien kyvykkyyksiä olisi koskaan edes halunneet alkaa opetella. Tekoälylle ja agenteille annetaan tarkoituksenmukaisesti niin paljon autonomiaa kuin mahdollista, kunnes todetaan että nyt homma ei enää toimi. Tämä on täysin luonnollista oppimista matkalla kohti kyvykkäämpiä agenttiarmeijoita. Siinä kohtaa koodari tai tiimi tarkastelee, mikä meni koodissa pieleen, mikä agenttien ohjeistuksessa aiheutti ongelman ja sen jälkeen korjaavat asian ja jatkavat. 

Tutkijat kysyvätkin osuvasti, onko tekoäly ajattelun tuki vai ajattelun korvike. Jos se on korvike, ymmärrysvelka kasvaa. Ymmärrysvelka ei ole kuitenkaan tekoälyn aiheuttama ongelma. Se on vanha ongelma koodauksessa. Siihen tekoäly on nimenomaan ratkaisu. 

Ohjelmistoalalla tunnetaan useita käsitteitä, jotka kuvaavat tätä ilmiötä jo vuosikymmenten takaa:

- **"Programming by Coincidence"** (Hunt & Thomas, *The Pragmatic Programmer*, 1999): kehittäjät kirjoittavat koodia, joka sattuu toimimaan, ilman että ymmärtävät miksi. Kirja vertaa tilannetta sotilaaseen, joka kävelee läpi miinakentän — ensimmäiset askeleet eivät räjähtäneet, joten hän päättää, että kenttä on turvallinen.

- **"Copy and paste is a design error"** (Steve McConnell, *Code Complete*, 2004): koodin monistaminen ilman ymmärrystä on suunnitteluvirhe.

Empiirinen data vahvistaa riskin: Fischer ym. (*IEEE Symposium on Security and Privacy*, 2017) analysoivat 1,3 miljoonaa Android-sovellusta ja havaitsivat, että 15,4 % niistä sisälsi Stack Overflow -koodinpätkiä. Näistä snippetteistä **97,9 %** sisälsi vähintään yhden tietoturvahaavoittuvan koodinpätkän.

Ymmärrysvelka on siis ollut olemassa jo pitkään ennen tekoälyä. Koodin kopioiminen ilman ymmärrystä on ollut riski jo kauan. Ero on, että agenttien kanssa prosessi on nopeampi ja siihen on tullut uusia elementtejä. Nyt virheet ovat kuitenkin helpommin havaittavissa, hallittavissa ja korjattavissa, kun hyödynnämme tekoälyagentteja ja käytämme oikeanlaisia työnkulkuja niiden kanssa.

Olen rakentanut kymmeniä sovelluksia hyödyntäen agenttipohjaista koodausta kymmenillä eri kielimalleilla. Se on luonut minulle ymmärryksen rakenteellisesta ja vaiheistetusta työnkulusta, joka säilyttää minun ymmärryksen prosessin jokaisessa vaiheessa, jos niin haluan. 

Haluan vielä tähdentää, että ohjelmointi ei ole minun leipätyöni. Olen oppinut itsenäisesti tekemällä, käymällä avoimia ohjelmointikursseja, syventämällä taitojani kysymällä generatiiviselta tekoälyltä, sekä ehdottomasti tärkeimpänä: agenttipohjaisen koodauksen kautta. 

## 1. Kaikki koodi ei ole samanarvoista

Ennen kuin alan nykyään kirjoittamaan koodia agenttiavusteisesti, teen arvion projektin eri osa-alueista: mikä on minkäkin tehtävän kriittisyys? VTT:n tutkijoiden pelko ymmärryksen katoamisesta pitäisi mielestäni olla suoraan verrannollinen siihen, kuinka vakavat seuraukset virheellä voi olla.  Tämä seuraava riskianalyysi ei ole mikään validoitu tai tutkittu tapa toteuttaa riskiarviota, mutta kuvaa karkeasti omaa tapaani lähestyä asiaa.

**Taso 1:** Avustettu luovuus (Matala riski)

- Esimerkki: Staattiset sivut, tyylittely (CSS), markkinointisisällöt, mockupit.
- Agentin rooli: Autonominen tekijä. Saa generoida kokonaisia tiedostoja.
- Ihmisen rooli: Kuraattori ja varmistaa, että esimerkiksi tyylittely on toteutettu siten, että se on helposti muokattavissa myöhemmin.

Autonomisuus: Agentilla on 90-100 % valta toteuttaa muutokset.

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
- Ihmisen rooli: Kriittinen tarkastaja. Ei hyväksy koodia ennen kuin on haastanut agentin ("Miksi et käyttänyt kirjastoa X?", "Miten tämä käsittelee kilpailutilanteen?"). Ihminen vastaa järjestelmän logiikan hallinnasta. Tarve koodin ymmärtämiseksi on korkea.

Autonomisuus: Agentilla on 15 % valta. Se tekee raskaan työn, mutta ihminen pitää ohjat tiukasti käsissään jatkuvan dialogin ja koodin yksityiskohtaisen tarkastelun kautta.

Korostan, että tämä "riskiarvio" on minulle matkan varrella sisäänrakentunut tapa toimia. Monissa tapauksissa agenteille voi antaa paljon autonomisuutta, kunhan vain tunnistaa ne kohdat, joissa se on mahdollista. 

Antaa siis agenttien touhuta ja koodata, jos se ei ole vaarana aiheuttaa merkittävää vahinkoa. Itsenäisillä agenteilla saattaa kuitenkin olla merkittävä vaikutus tuottavuuteen. Jos tavaraa osuu tuulettimeen, niin se on erinomainen oppimisen paikka. 

## 2. Järjestelmällinen agenttikoodauksen hyödyntäminen

VTT:n artikkelissa puhutaan siitä, että lihaa ja verta olevien koodareiden sijaan ollaan siirtymässä vähitellen agenttien omiin suljettuihin silmukoihin, joissa ne koodaavat itsenäisesti, tarkistavat koodin, muokkaavat ja parantavat sitä. Jokainen joka on agenttipohjaisen koodaamisen kanssa tehnyt pidempään työtä tai kokeiluja, tietää, että se on täysin mahdollista, mutta myös sen, että se on vaatii huomattavan määrän ymmärrystä kielimallien kyvyistä, rajoitteista, projektien logiikasta ja ihmisen roolista tässä kokonaisuudessa. 

Kun itse aloin käyttämään agenttipohjaista koodausta projekteissa, näin, miten agentit rakensivat sivuston muutamassa minuutissa ja lähes kaikki toimi kuten pitikin. Nälkäni tietenkin kasvoi ja kun testasin isompia kokonaisuuksia, niin kävi kuten oletettua... projekti ei edes rakentunut ja kaatui virheeseen heti ensimmäisessä mutkassa. 

Siitä aloinkin kehittämään omaa ymmärrystäni, mitkä ovat eri kielimallien rajat koodauksessa, mihin ne kykenevät, missä ne loistavat ja mikä on ihmisen eli minun itseni rooli tässä kokonaisuudessa. 

Nyt, noin 2 miljardia tokenia myöhemmin, ymmärrykseni tästä työnkulusta on aivan toinen kuin aloittaessani tämän parissa työskentelyn/harjoittelun:

1. **Projekti- Arkkitehtitason suunnitelma** — Määritän projektin tavoitteet, kriittiset työnkulut, teknologia stackin, MVP-vaiheen valmiuden, turvallisuustarkistukset, selkeät vaiheet ja jokaiselle vaiheelle läpäisyvaatimukset. Käytän tässäkin vaiheessa tekoälyä apuna ja yleensä ensimmäisen vaiheen teen Gemini 3.1 pron kanssa keskustellen.
2. **Dialogi useamman mallin kanssa** — Käyn läpi lähestymistavan ja rakennan projektin vaiheistusta pidemmälle 2-3 kyvykkään mallin (Opus 4.6 tai 4.7 ja Codex 5.3) kanssa ennen kuin aloitan varsinaisen projektin.
3. **Suojakaiteet + AGENT.md** — Annan eksplisiittiset ohjeet, rajoitteet ja formaatit joka vaiheelle erikseen, ei kerran projektin alussa.
4. **Yhteenveto + katselmointi** — Agentti tiivistää jokaisen vaiheen tuloksen; minä katselmoin koodin ennen kuin annan "vihreää valoa" seuraavalle vaiheelle.
5. **Muistidokumentaatio** — Jokaisessa vaiheessa agentti tallentaa edistymisen, päätökset ja hylätyt vaihtoehdot MEMORY.md-tiedostoon.

Noudatan tätä melko orjallisesti merkittävissä projekteissa. Pienemmissä tai vain omaksi iloksi tehdyissä hommissa voin oikoa mutkia suoriksi.

Väitän, että tällä työnkululla ymmärrykseni koodista on korkeammalla tasolla kuin perinteisellä koodauksella ja dokumentaatio on huikeasti parempaa kuin siinä tapauksessa, että koodari on kirjoittanut hampaat irvessä pari riviä dokumentaatiota viikon tai sprintin päätteeksi. En usko, että dokumentaation kirjoittaminen on yhdenkään ammattidevaajan huippuhetkiä työviikossa, vaan pakollinen paha, jonka tärkeys ymmärretään, mutta joka on myös todella tylsää puuhaa.


## 3. MEMORY.md — Organisaation jaettu ymmärrys

Tutkijat ehdottavat viiden kohdan listassaan siirtymistä "näkymättömästä näkyvään osaamiseen" sekä "delegoinnista oppimiseen". Minun yksi keskeisimmistä ratkaisuista tähän ja parantava lääke ymmärrysvelkaan on **jatkuvasti päivittyvä MEMORY.md** -tiedosto, joka toimii projektin muistina. Se tallentaa sen, mikä muuten katoaa, kun tiimin jäsen vaihtuu tai projekti jatkuu kuukausia myöhemmin.

MEMORY.md:ssä on kolme kriittistä osaa:

- **Projektin eteneminen:** Mitkä vaiheet projektista on suoritettu ja miten ne toteutettu? 
- **Konteksti:** Miksi valitsimme tämän lähestymistavan, tämän kirjaston, tämän arkkitehtuurin, mitkä ovat riippuuvuudet?
- **Vaihtoehdot:** Mitä muita tapoja harkittiin ja miksi ne hylättiin?

Tämä tiedosto on kultaa: jos uusi kehittäjä tulisi tiimiin tai kun palaan koodiin pitkän ajan kuluttua, historia ei ole kadonnut tekoälyn chatti-ikkunaan. Tämä on tapa, jolla voitan ymmärrysvelan — ei vähentämällä tekoälyn käyttöä, vaan rakentamalla sen käytölle selkeät säännöt ja ymmärtämällä sen mahdollisuudet ja heikkoudet. Muistitiedosto täydentyy myös siinä vaiheessa, kun ohjelmasta löytyy se tuotantoon päässyt bugi ja se korjataan - tekoälyn avustamana. Projektille luodaan tietenkin myös tekninen dokumentaatio, mutta se on korkeamman tason dokumentointi kuin projektin muisti.

## 6. FoSW-projektisivuston koodi on elävä varoitusmerkki tutkijoiden pelkäämästä ymmärrysvelasta

Tutkimusryhmällä on projektilleen sivusto: [futuresofsoftwarework.github.io/FoSW](https://futuresofsoftwarework.github.io/FoSW/).

Tarkastelin sivustoa ja sen lähdekoodia GitHub-repositoriosta — ja löysin aika paljon näyttöä siitä, että sivusto on rakennettu vibe-koodaamalla:

- **AI-generoidut commit-viestit:** `"feat: implement metrics dashboard brainstorming and design documentation with supporting assets"` — ihmisen kirjoittama viesti olisi tyypillisesti lyhyempi ja suorempi.
- **Tailwind-luokkien valtavat määrät:** Esim. yksittäinen `<h1>`-elementti sisältää yli 10 Tailwind-luokkaa, mukaan lukien arbitrary values (`drop-shadow-[0_0_15px_rgba(245,158,11,0.5)]`).
- **Inter-fontti:** klassinen AI-fontti, joka tulee AI:lta lähes oletuksena, kun rakennetaan jotain tech-tyyppistä sivustoa.
- **Custom "AI-värit":** `neon-gold`, `hologram-cyan`, `electric-blue`, `midnight` — klassisia tekoälykäyttöliittymien värejä.
- **Kehittäjä + Claude:** Kaikki commitit ovat Arton tai Arton ja Claude coden tekemiä.
- **Pelkkä silmäys riittää:** Sivusto on hyvin pitkälti juuri sen näköistä, mitä AI tuottaa. Eikä siinä mitään, jos se on tavoiteltua. Omilla sivuillani tämä ei haittaa, koska se ei sinänsä ole mikään virhe. En ole itsekään UI designer + haluan, että AI myös näkyy ja tuntuu omalla sivustollani.

Toisaalta projektisivusto on matalan riskin sivusto ja projektin kuvauksessahan myös mainitaan, että kokeilut kuuluvat tähän tutkimusprojektiin. Sivusto ei kerää tietoa käyttäjistä, se ei kerää evästeitä, ei käsittele maksuliikennettä tai muutakaan arkaluontoista tietoa. Jos sivun on tarkoitus pysyä tällaisena matalan riskin projektina, niin sehän on ihan ok :) 

Ehkä läpinäkyvyyden vuoksi voisi AI-Signaleista mainita, että ne on tekoälyautomaation luomia .json-tiedostoja, jotka mahdollisesti sivuston ylläpitäjä on tarkistanut ja muuttanut niiden statuksen draftista --> published -muotoon tai ainakin hän on joutunut ne committoimaan, jolloin ne vasta deployautuu GitHub pagesiin. Julkisen rahoituksen projekteissa olisi mielestäni kohtuullista, että lukijalle kerrotaan tämä kaikki hyvin eksplisiittisesti heti sivuston alussa.

## Huomautus "tutkimusmenetelmästä"

Tämän blogin taustatyö — mukaan lukien FoSW-sivuston analyysi, lähdeviitteiden tarkistus — tehtiin hyödyntäen agentic AI -työkalua. Tällä kertaa käytin runsaasti OpenClaw:ta, sillä sen avulla sain suoraan haettua tietoa sivustoon liittyvästä GitHub repositoriosta ja lähdekoodista ilman, että minun itseni tarvitsi alkaa käymään kaikkea läpi, etsimään tietojen sijainta tai kahlaamaan julkista repoa alusta loppuun läpi.

Määritin kyselyt, tarkistin tulokset, ja rakentelin kokonaiskuvan vaihe vaiheelta yhdessä agentin kanssa. Tarkistin jokaisen lähteen alkuperäisestä materiaalista. AI teki virheitä matkan varrella, mutta keskustellen sen kanssa ja tarkistaen tulokset "käsin" nämä puutteet löydettiin. 

Ilman tekoälyä en olisi jaksanut — enkä ehtinyt — perehtyä näin perinpohjaisesti aiheeseen. Ymmärrykseni olisi jäänyt tästäkin asiasta vajaaksi. **Juuri tämä on pointtini:** kun tekoälyä käytetään rakenteellisessa työnkulussa, se ei vähennä ymmärrystä — se laajentaa sitä.

## Yhteenveto ja loppusanat

Tekoälyagentit eivät aiheuta ymmärrysvelkaa. Meillä ehtii tulla ymmärrysvelkaa agenttien hyödyntämisestä, jos emme lähde kokeilemaan ja testaamaan niitä ennakkoluulottomasti. Se tarkoittaa myös aluksi sitä, että emme välttämättä ymmärrä kaikkea, mitä agentit tekevät, mutta ei meidän tarvitsekaan. Agenttipohjaisen koodaamisen opettelu on melkoisen kaoottista ja sen kanssa tehdessä vastaan tulee erityisesti aluksi hengästyttävän paljon uutta tietoa, jotka tekee mieli ottaa haltuun. Oppimisprosessi sisältää paljon käytännön harjoittelua sekä jatkuvaa asioihin perehtymistä. Agenttipohjainen koodaus, ja kielimallit ylipäätään, ovat uutta teknologiaa, josta ei ole vielä juurikaan olemassa kattavia opetuskokonaisuuksia ja jos onkin, niin ne ovat luultavasti jo vanhentuneita. 

Tekoäly pitää pystyä Suomessakin näkemään mahdollisuutena. Virheellistä ja haavoittuvuuksia sisältävää koodia on aina kopioitu projekteihin eikä kukaan esimerkiksi tilaajan päässä ole kysynyt, että mistä tämä skripti on tullut ja pyytänyt selittämään sen juurta jaksaen. Agenttien luoman dokumentaation pohjalta, myös tilaajalla on mahdollisuus ymmärtää paremmin mitä koodi tekee. Ihminen ei todellakaan ole vastaus virheiden välttämiseen vaan uuden teknologian täysimittainen hyödyntäminen. Rakennetaan sellaisia työnkulkuja agenttien avulla, että ne nostavat meidän ymmärryksemme ihan uudelle tasolle.

Viimeisen parin vuoden aikana kielimallit ovat ottaneet huikeita harppauksia eteenpäin. Ne ovat nyt jo erittäin kyvykkäitä, eikä tälle kehitykselle ole vielä näköpiirissä suurta hidastumista. Kielimallit kuten Anthropicin Mythos löytävät ja korjaavat ohjelmistoissa olevia ihmisten sinne tekemiä bugeja vauhdilla, jotka ihmiselle olisivat mahdottomia.

Kritisoin vielä lyhyesti sitä, että Suomessa keskitytään mielestäni liikaa tekoälyn negatiivisiin asioihin ja ollaan lähtökohtaisesti kriittisiä sen sijaan, että voisimme tarkastella sen mahdollisuuksia. Nyt on aika polttaa tokeneita, tehdä virheitä ja oppia niistä. Ymmärrys tulee siinä mukana, kun huomaamme, että agentit eivät olekaan kaikkivoipaisia, vielä. 

What an amazing time to be alive! (Let's not ruin it)