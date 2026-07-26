---
title: "Go + HTMX vs. Next.js: Miksi rakensin kopion sivuistani toisella ohjelmointikielellä?"
description: "Rakensin identtisen Go+HTMX-toteutuksen Next.js-sivustoni rinnalle, mitä opin VPS:n pystyttämisestä Hetzneriin ja SQLiten pyörittämisestä omalla palvelimella."
publishedAt: "2026-07-25"
slug: "go-htmx-nextjs-vertailu-hetzner"
draft: false
tags: ["koodaus", "go", "htmx", "devops", "sqlite"]
---

Tämä sivusto, jota luet juuri nyt, on rakennettu go+HTMX pinolla ja hostattu Hetznerin pilvipalvelimessa. Mutta jos vaihdat oikeasta alakulmasta löytyvästä kirjaimesta toiseen näkymään, pääset tarkastelemaan toista toteutusta. Se on tehty next.js:llä ja hostattu Vercelissä. Tietokantana on Turso. Molemmat ovat sama sivusto – sama etusivu, sama AI-portsari, sama dashboard – mutta kaksi täysin eri teknologiavalintaa samassa paketissa. Pieniä eroja joissain toteutuksissa sivuston designin osalta, mutta ei mitään merkittäviä tai sivuston suoristuskykyihin vaikuttavia.

### Miksi lähdin rakentamaan koko sivuston uudelleen

Kolme syytä kolahti minulla yhtä aikaa. Ensiksi minua kiinnosti tietää, mihin todellisuudessa Go+HTMX pystyy verrattuna Next.js:ään, jolla olen useimmat projektini toteuttanut. Halusin rakentaa saman websovelluksen kahteen kertaan ja mitata eroa livenä käyttäjän selaimesta. Toiseksi halusin opetella VPS:n pystyttämisen Hetzneriin nollasta: SSH, palomuuri, systemd, Caddy ja Let's Encrypt käsin, ei. Kolmanneksi halusin kokeilla, miltä SQLite tuntuu tuotannossa omalla levyllä sen sijaan, että nojaan Tursoon – WAL-tila, migraatiot, varmuuskopiot ja restore-drillit ilman kolmannen osapuolen hallintapaneelia.

Ja kuten varmaan jo sivustoni sisällöstä päättelitkin, niin kaikki on rakennettu hyödyntämällä agenttipohjaista koodausta. Työkaluna minulla on Kilo Code. Tämän projektin kielimalleina hyödynsin **GPT-5.6 Sol**, **Sonnet 5**, **Kimi K3** ja **Hy3**:sta. Taisin jossain vaiheessa hairahtua käyttämään **Gemini 3.6 Flashia**. Rinnalla kävin keskustelua Geminin kanssa eri toteutusmalleista, jota projektin edetessä tuli vastaan ja validoin suunnitelmat sekä toteutukset myös sen avulla.

Sivuhuomatuksena kielimalleista. Tykästyin todella paljon tämän projektin aikana Tencentin hy3 malliin. Totesin myös että Kimi K3 hoitaa homman maaliin, mutta se lörpöttelee ja sen toteutukset kestävät ihan älyttömän pitkään verrattuna esim. Sonnet 5:een tai GPT-5.6:een. 

### 13 vaihetta, yksi binääri

Projektin laajuudeksi kasvoi äkkiä, kuten agenttipohjaisessa koodauksessa usein tuppaa käymään, koko julkinen sivusto: etusivu, SENTINEL-7-portsari, blogi, tietosuoja, ja täysi lukittu dashboard ynnä sen kaikki näkymät (overview, projektit, tech, blogireader, changelog, AI Pulse, asetukset). Ainoa tietoinen rajaus oli postikorttigeneraattori – se jäi Next.js:n yksinoikeudeksi, koska sen porttaaminen ei olisi tuonut mitään opittavaa lisäarvoa vertailuun. Rakensin toteutuksen vaihe kerrallaan (13 vaihetta + yksi ylimääräinen standalone-gate), ja jokaisen jälkeen kirjasin muistiin, mitä tehtiin ja mikä meni rikki. Lopputulos on yksi itsenäinen Go-binääri, jossa templaatit, staattiset assetit, media ja migraatiot ovat kaikki embedattuna – `go build` ja siinä se on, ei erillistä `node_modules`-riippuvuutta ajon aikana.

### Kovin pähkinä: kahden stackin pitäminen synkassa

Helpoin osa oli UI:n kopioiminen – html/template + Alpine.js + Tailwind toistavat React-komponentit yllättävän suoraviivaisesti. Kielimallit gpt5.6 Sol ja Tencentin hy3 hoitivat sen melko kivuttomasti. Vaikeampaa oli varmistaa, että kaksi täysin eri kieltä tuottavat *bittitasolla* saman lopputuloksen kriittisissä kohdissa. Unlock-eväste on HMAC-allekirjoitettu jaetulla salaisuudella, ja sen täytyy validoitua identtisesti riippumatta siitä, kumman stackin se loi – vaihdoin stackia kesken session eikä käyttäjän pitäisi huomata mitään. Todistin tämän Node↔Go golden-vector-testeillä, en silmämääräisesti. SENTINEL-7-portsarin taso- ja lukituslogiikka on portattu rivi riviltä, ja tilastosemantiikka (kuinka `unlockedCount` tai `avgMessagesToUnlock` lasketaan) kävi läpi oman erillisen korjauskierroksen, kun huomasin Go-version laskevan istuntoja siinä missä Next.js laskee rivejä.

Isoin todenne rakennettiin vasta lopuksi: standalone-drilli, joka ajaa 40 väitettä ja pakottaa Go:n toimimaan täysin ilman Next.js:ää – Next.js-loopback tapetaan tarkoituksella, ja drilli tarkistaa, että AI Pulsen HN/GitHub/Yahoo-syötteet, Gemini-yhteenvedot, palvelimen uudelleenkäynnistys kesken päivityksen ja jokainen julkinen reitti selviävät siitä siististi. Tämä ajetaan nykyään joka ainoassa deployssa CI:ssä ennen kuin mitään viedään tuotantoon.

### VPS Hetzneriin ja SQLite tuotannossa

Hetzneer toteutus: CX23 pilvipalvelin Helsingissä, Ubuntu, avainpohjainen SSH ja Cloud Firewall. Caddy hoitaa TLS:n automaattisesti Let's Encryptilta – ja tässä sain ensimmäisen oikean opetuksen: ensimmäinen sertifikaatti tuli vahingossa staging-CA:sta, jonka selaimet hylkäävät. Korjaus oli yksinkertainen (oikea `acme_ca`-osoite + vanhan sertifikaatin poisto + uudelleenkäynnistys), mutta se opetti, että TLS:ää ei pidä julistaa valmiiksi tarkistamatta issueria `openssl`:llä. SQLite pyörii `/var/lib/goth/goth.db`:ssä WAL-tilassa, ja tietokanta varmuuskopioituu paikallisesti sekä Cloudflare R2:een `rclone`-cronilla joka yö – round-trip-tarkistus (`sha256sum` paikallinen vs. R2) ja restore-drilli kokonaisuudessaan tehtynä ennen kuin mitään DNS:ää kosketettiin. `deploy.sh` tarkistaa checksummin, ajaa migraatiot, vaihtaa symlinkin atomisesti ja rollbackkaa itse, jos health check ei mene läpi.

### Next.js sai omat verkkotunnuksensa ja vertailusta tuli reilu

Alun perin Next.js oli tarkoitus viedä samalle CX23:lle Caddyn evästereitityksen takana. Päädyin toiseen ratkaisuun: Next.js sai oman subdomaininsa (`next.karotammela.fi`) ja pysyi Vercelissä, kun Go otti apex-domainin. Näin vertailu mittaa oikeasti kahta erilaista deploy-mallia – "self-hosted Go VPS Helsingissä" vastaan "Vercel Edge" – eikä keinotekoista viivettä, jonka Caddy-proxy olisi lisännyt yhdelle puolelle. Tech Switcher (se pyöreä nappi oikeassa alakulmassa) tekee nyt cross-origin-uudelleenohjauksen polun ja kielen säilyttäen, ja Live Performance -widget pingaa molempien `/api/ping`-reittejä suoraan selaimesta ja näyttää mediaani-TTFB:n – ei palvelinpuolen arviota, vaan sen, mitä sinun selaimesi todella mittasi. Ensimmäisillä mittauksilla Go+HTMX vastasi noin 50 ms:ssa, Next.js Vercelistä noin 190 ms:ssa – ero näkyy suoraan palkkien pituudessa. Tällä sivustollahan tuolla viiveellä ei ole mitään merkitystä, mutta nopeuden kyllä huomaa esimerkiksi siinä, miten nopeasti AI-portsari palauttaa vastauksen tai miten nopeasti eri sivut latautuvat selaimella. 

### Viimeinen palanen: deploy yhdellä pushilla

Julkaisu oli tähän asti manuaalinen: `make release`, `scp` palvelimelle, `sudo ./deploy.sh`. Viimeistelin sen tänään automaattiseksi GitHub Actions -putkeksi: erillinen, salasanaton deploy-avain vain CI:tä varten, kolme repo-secretiä, ja workflow jossa `test` (vet, yksikkötestit, standalone-drilli) → `release` (identtinen build kuin manuaalinen) → `deploy` (scp + `deploy.sh` + tuotannon health check) ajetaan peräkkäin joka kerta, kun `goth/`-kansioon pushataan `main`-haaraan. Ensimmäinen automaattiajo onnistui suoraan – testit, build ja deploy alle neljässä minuutissa, ja tuotanto vastasi terveenä heti perään.

### Mitä opin

Go+HTMX yllätti siinä, kuinka vähällä koodilla pääsee pitkälle, kun ei tarvitse React-hydraatiota tai bundlerin optimointeja – yksi 20-30 megan binääri hoitaa kaiken, TTFB on luokkaa neljäsosa Vercel-edgestä mitattuna omalta koneelta Helsingissä. Hinta tästä pitää toki maksettu käsityönä: monet jutut Next.js:ssä tulevat suoraan kehyksestä. Kumpikaan stack ei ole absoluuttisesti "parempi" – mutta nyt tiedän, mistä ero tulee, koska rakensin molemmat itse ja mittasin sen omalla selaimellani. Taidan siirtää useammat kevyet sivut Go+HTMX:lle, sillä ne latautuvat merkittävästi nopeammin ja vaativat vähemmän resursseja. CX23 instanssiin myös mahtuu useampi projekti noin 5 € kuukausihintaan. Saan tiputettua webhotellikulut minimiin tämän avulla, sillä perustin myös toista projektiani varten Purelymail-tilin, jolla saa sähköpostit pelaamaan 10€ vuosikustannuksella. 
