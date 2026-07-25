# Hetzner-julkaisusuunnitelma — karotammela.fi

Tämä on **sinun** tehtävälistasi tuotantoon viemiseksi. Koodipuoli (backup,
Caddyfile, deploy.sh, env-kontrakti, runbook, savutesti) on valmis
`goth/`-repossa. Jäljellä on vain palvelimen pystytys ja julkaisu itse.
Yksityiskohtaiset komennot löytyvät `goth/docs/runbook.md`:stä — tässä
tiedostossa on **järjestys, valinnat ja mitä raportoida minulle** jokaisen
vaiheen jälkeen.

Merkinnät: **PÄÄTÄ** = sinun tehtävä valinta, **RAPORTOI** = kerro minulle
tulos/arvo seuraavaa vaihetta varten (älä koskaan liitä salaisuuksia itse
tekstiin — riittää "kyllä/täytetty").

**Kokonaistilanne (2026-07-25):** Vaiheet 1–8 valmiit ja merkitty ✅ alla.
Palvelin `ubuntu-4gb-hel1-1` (`135.181.27.78`, hel1/Helsinki) pystyssä,
SSH-avainkirjautuminen ja palomuuri kunnossa, Caddy asennettu ja validoitu,
salaisuudet täytetty (täsmäävät Vercelin tuotantoarvojen kanssa), Cloudflare
R2 -varmuuskopiokanava toimii testattuna (rclone + cron 04:00 UTC), ja
**karotammela.fi on julkinen Go+HTMX-tuotanto CX23:llä** (harmaa pilvi →
`135.181.27.78`, tuotanto-Let's Encrypt -TLS). **Vaihe 9:n koodiosuus valmis**
(Next.js + Go `/api/ping` + CORS, cross-origin Tech Switcher, client-side
perf-widget); jäljellä vain Vercelin subdomain-julkaisu ja Cloudflare-CNAME
(manuaaliset, ks. Vaihe 9).

---

## Vaihe 0 — Ennen kuin tilaat mitään

- [x] **DNS**: `karotammela.fi`:n tietueita hallitaan **Cloudflaressa** —
      pääsy jo olemassa, tarvitaan vaiheessa 8.
- [x] **Off-server-varmuuskopiot**: **Cloudflare R2**, tili jo olemassa —
      ks. Vaihe 4.
- [x] **Salaisuuksien säilytys**: **Google Drive**, salattuna paikallisesti
      ennen latausta (esim. `age`- tai `gpg`-salaus). Älä koskaan lataa
      plain-text-tiedostoa `/etc/goth/goth.env`-arvoista Driveen — salaa
      ensin, ja pidä salauksen salasana **erillään** Drivestä.

---

## Vaihe 1 — Hetzner CX23 + SSH + palomuuri

1. Luo Hetzner Cloud -tili (jos ei ole) ja uusi projekti.
2. Luo CX23-palvelin:
   - **PÄÄTÄ sijainti**: suositus Helsinki (fsn1/hel1 — matala latenssi
     Suomeen, `.fi`-domain).
   - **PÄÄTÄ käyttöjärjestelmä**: suositus Ubuntu 24.04 LTS.
   - Liitä **SSH-avaimesi** palvelimen luonnissa. Salasanakirjautumista ei
     tarvita eikä sitä pidä sallia.
3. **PÄÄTÄ palomuuri** (Hetzner Cloud Firewall, liitetään palvelimeen):
   - Salli TCP 22 (SSH) — rajaa mielellään omaan IP-osoitteeseesi.
   - Salli TCP 80 ja 443 (Caddy: HTTP→HTTPS-redirect + ACME + liikenne).
   - **Ei muita portteja auki ulos.** Go-sovellus (8080) ja Next.js (3000)
     jäävät vain `localhost`:iin, Caddy on ainoa julkinen ovi.
4. Kirjaudu SSH:lla varmistaaksesi pääsyn: `ssh root@<palvelimen-IP>`.

**Tila:** ✅ Palvelin luotu (CX23, hel1/Helsinki, Ubuntu 26.04 LTS,
IP `135.181.27.78`). ✅ SSH-avainkirjautuminen toimii, salasanakirjautuminen
poistettu käytöstä. ✅ Cloud Firewall luotu (TCP 22/80/443 + ICMP) ja
liitetty palvelimeen. **Vaihe 1 valmis.**

**RAPORTOI minulle:** palvelimen julkinen IP-osoite, ja että SSH-kirjautuminen
avaimella onnistui.

---

## Vaihe 2 — Palvelimen peruskunto

### SSH lyhyesti: missä mikä komento ajetaan

SSH on tapa avata **etäterminaali** Hetzner-palvelimelle — kun olet
yhdistänyt, näppäimistölläsi kirjoittamasi komennot suoritetaan CX23:lla,
ei omalla koneellasi. Kaikki tämän suunnitelman `[host]`-merkityt komennot
ajetaan tässä etäterminaalissa; `[dev]`-merkityt omalla koneellasi
**eri** terminaali-ikkunassa/välilehdessä.

1. Avaa terminaali **omalla koneellasi** (Linux/Mac: Terminal-sovellus;
   Windows 10/11: Windows Terminal tai PowerShell, `ssh` toimii niissä
   valmiiksi).
2. Yhdistä palvelimelle (IP Vaiheesta 1):
   ```bash
   ssh root@<palvelimen-IP>
   ```
3. Kun yhteys onnistuu, komentokehote vaihtuu muotoon
   `root@<palvelimen-nimi>:~#` — **tästä eteenpäin olet CX23:lla**, ei
   omalla koneellasi. Kaikki alla olevat `[host]`-komennot kirjoitetaan
   tähän samaan ikkunaan.
4. Kun haluat lopettaa etäistunnon: `exit` (palaat omalle koneellesi —
   kehote vaihtuu takaisin normaaliksi). Ei tarvitse sulkea nyt, voit pitää
   ikkunan auki loppuun asti.

Jos tarvitset toisen komennon omalla koneellasi (esim. `scp` seuraavassa
kohdassa) samaan aikaan, avaa **uusi** terminaali-ikkuna/-välilehti — SSH-
istunto jää auki ensimmäiseen.

### Deploy-kansion tuonti palvelimelle

`deploy/`-kansio (systemd-yksiköt, Caddyfile, env-esimerkki) täytyy saada
CX23:lle. Helpoin tapa: kopioi se omalta koneeltasi `scp`:llä.

```bash
# [dev] — OMALLA koneella, UUDESSA terminaali-ikkunassa (ei SSH-istunnossa)
# aja tämä komento tämän repon goth/-kansiosta
scp -r deploy root@<palvelimen-IP>:/root/deploy
```

Tämä luo `/root/deploy/`-kansion palvelimelle. Palaa SSH-ikkunaan (kohta 2)
jatkoa varten.

### Ohjelmistot ja peruskunto (CX23:lla, SSH-ikkunassa)

```bash
# [host]
apt update && apt upgrade -y
# Caddy: virallinen apt-repo, ks. https://caddyserver.com/docs/install
apt install -y caddy
caddy version   # varmista ≥ v2.10
```

### Käyttäjä, hakemistot ja konfiguraatiotiedostot (CX23:lla)

```bash
# [host]
useradd --system --home /var/lib/goth --shell /usr/sbin/nologin goth
mkdir -p /opt/goth/releases /var/lib/goth/backups /etc/goth
chown -R goth:goth /var/lib/goth

cp /root/deploy/systemd/goth.service /root/deploy/systemd/goth-refresh.{service,timer} \
   /root/deploy/systemd/goth-backup.{service,timer} /etc/systemd/system/
cp /root/deploy/caddy/Caddyfile /etc/caddy/Caddyfile
caddy validate --config /etc/caddy/Caddyfile
systemctl daemon-reload
```

Timerit **eivät** aktivoidu tässä vaiheessa (aktivointi on Vaihe 7).

**RAPORTOI:** `caddy version` -tulos, ja että `caddy validate` läpäisi
virheittä.

**Tila:** ✅ Caddy v2.11.4 asennettu, ✅ `goth`-käyttäjä ja hakemistot
luotu, ✅ systemd-yksiköt ja Caddyfile kopioitu, ✅ `caddy validate` →
"Valid configuration". **Vaihe 2 valmis.**

---

## Vaihe 3 — Salaisuudet (`/etc/goth/goth.env`)

Kopioi `deploy/goth.env.example` → `/etc/goth/goth.env` (0640, root:goth) ja
täytä seuraavat rivi kerrallaan. Käytä `openssl rand -hex 32` generointiin
missä pyydetty.

| Muuttuja | Mistä arvo tulee |
| --- | --- |
| `UNLOCK_COOKIE_SECRET` | `openssl rand -hex 32` — **täytyy olla sama** kuin Next.js-puolen `UNLOCK_COOKIE_SECRET`, muuten unlock-tila hajoaa tech-vaihdossa |
| `CRON_SECRET` | `openssl rand -hex 32` — **sama** kuin Next.js `CRON_SECRET` |
| `GOOGLE_GENERATIVE_AI_API_KEY` | olemassa oleva Gemini-avaimesi |
| `AI_MODEL` | oletus `gemini-3.1-flash-lite` käy, jos et halua muuttaa |
| `RESEND_API_KEY`, `CONTACT_FROM_EMAIL`, `CONTACT_TO_EMAIL` | Resend-tililtäsi; kaikki kolme pakollisia, muuten yhteydenottolomake palauttaa 503 |
| `NEXT_PING_URL` | oletus `https://next.karotammela.fi/api/ping` — Next.js-vertailu hostataan Vercelissä (ks. Vaihe 9); ei enää `localhost:3000` |
| `GOTH_BACKUP_DIR` / `GOTH_BACKUP_KEEP` | oletukset (`/var/lib/goth/backups`, `14`) käyvät sellaisenaan |

Kirjaa **kaikki** arvot ulkopuoliseen salasananhallintaan (Vaihe 0). Tiedosto
itse ei koskaan mene gittiin.

**RAPORTOI:** vahvistus per rivi "täytetty: kyllä/ei" (ei arvoja itse
viestiin), ja että `chmod 0640` + `chown root:goth` on tehty.

**Tila:** ✅ `/etc/goth/goth.env` luotu ja täytetty. ✅ `UNLOCK_COOKIE_SECRET`
ja `CRON_SECRET` korvattu Vercelin tuotantoarvoilla (täsmäävät Next.js:n
kanssa). **Vaihe 3 valmis.**

---

## Vaihe 4 — Off-server-varmuuskopiokanava (Cloudflare R2)

Koodi tekee vain **paikallisen** snapshotin (`/var/lib/goth/backups/`) —
R2-kopiointi on erillinen, manuaalinen asennus hostilla.

1. Käytä olemassa olevaa R2-bucketiasi (tai luo uusi vain tälle,
   esim. `karotammela-goth-backups`, jotta oikeudet on helppo rajata).
2. Luo R2 API-token **vain tälle bucketille** (Object Read & Write) →
   Access Key ID + Secret Access Key. Endpoint on
   `https://<account-id>.r2.cloudflarestorage.com`.
3. Asenna hostille `restic` tai `rclone`:
   - `restic`: `restic init -r s3:https://<account-id>.r2.cloudflarestorage.com/<bucket>`
     (kysyy repo-salasanan — tallenna se Drive-salaukseen samalla tavalla
     kuin muut salaisuudet; sitä tarvitaan restoreen).
   - `rclone`: `rclone config` → tyyppi `s3`, provider `Cloudflare`,
     endpoint yllä.
4. Lisää yksinkertainen cron/systemd-ajo, joka synkkaa/varmuuskopioi
   `/var/lib/goth/backups/` R2:een päivittäin (paikallisen
   `goth-backup.timer`:in jälkeen, klo 03:30 UTC → esim. 04:00 UTC).

**RAPORTOI:** bucketin nimi (ei API-avaimia), ja että ensimmäinen synkkaus
onnistui (esim. `rclone ls r2:<bucket>` tai `restic snapshots` -tuloste).

**Tila:** ✅ R2-bucket `karotammela-goth-backups` + oikein rajattu Account
API -token (Object Read & Write, R2:n omalta "Manage R2 API Tokens"
-sivulta, **ei** yleiseltä "Manage account" -sivulta). ✅ `rclone` asennettu
ja konfiguroitu (`no_check_bucket = true` tarvittiin bucket-rajatulle
tokenille). ✅ Kirjoitus/luku/poisto -testi onnistui. ✅ Cron-ajastus
päivittäiselle synkkaukselle klo 04:00 UTC lisätty. Päästä-päähän-testi
oikeilla varmuuskopiotiedostoilla tehdään Vaiheessa 6–7. **Vaihe 4 valmis.**

---

## Vaihe 5 — Ensimmäinen julkaisu

`[dev]` = omalla koneellasi tässä repossa.

```bash
# [dev]
make release
scp dist/goth-*-linux-amd64.tar.gz{,.sha256} <host>:/tmp/

# [host]
sudo ./deploy.sh /tmp/goth-YYYYMMDD-linux-amd64.tar.gz
```

`deploy.sh` tarkistaa checksummin, ajaa migraatiot uudella binäärillä ennen
liikenteen kytkemistä, vaihtaa symlinkin atomisesti, käynnistää palvelun
uudelleen ja health-checkkaa. Epäonnistuessa se **rollbackkaa itsestään**.

**RAPORTOI:** `deploy.sh`:n tuloste kokonaan, ja
`curl -s http://127.0.0.1:8080/api/ping` -vastaus.

---

## Vaihe 6 — Restore-drilli (tee ennen DNS-cutoveria)

Aja `runbook.md` §5 "Restore drill" kertaalleen kokonaan läpi ensimmäisen
backupin valmistuttua (aja tarvittaessa `systemctl start
goth-backup.service` manuaalisesti ensin, jotta on tiedosto restoroitavaksi).

**RAPORTOI:** onnistuiko palautus, ja `PRAGMA integrity_check;` -tulos
(pitää olla `ok`).

---

## Vaihe 7 — Timerit käyttöön

```bash
# [host]
systemctl enable --now goth-refresh.timer goth-backup.timer
systemctl list-timers 'goth-*'
```

Odota vähintään yksi ajo kummastakin (backup 03:30 UTC, refresh 08:00 UTC),
tai käynnistä manuaalisesti testiksi (`systemctl start
goth-backup.service` / `goth-refresh.service`).

**RAPORTOI:** `systemctl list-timers` -tuloste, ja
`journalctl -u goth-refresh.service -n 20` / `goth-backup.service -n 20`
lopputulos (exit 0 / status).

---

## Vaihe 8 — DNS-cutover + TLS

⚠️ **Tärkeä Cloudflare-huomio:** Caddy hankkii TLS-sertifikaatin
automaattisesti Let's Encryptilta HTTP-01-haasteella, joka vaatii **suoran**
yhteyden CX23:een porttiin 80. Cloudflaren oranssi pilvi (proxied) ohjaisi
liikenteen Cloudflaren edgen läpi ja **rikkoisi** tämän — Caddy ei koskaan
näkisi ACME-haastetta.

**PÄÄTÄ:** pidä `karotammela.fi`:n A/AAAA-tietue Cloudflaressa
**"DNS only" (harmaa pilvi)** -tilassa, ei "Proxied" (oranssi). Tämä vastaa
suoraan sitä, mitä runbook ja Caddyfile olettavat. (Jos haluat myöhemmin
Cloudflaren proxyn/CDN:n käyttöön, se vaatii erillisen muutoksen — Cloudflare
Origin -sertifikaatin tai DNS-01-haasteen Cloudflare API -tokenilla — jätä se
myöhemmäksi, ei osa tätä cutoveria.)

**Caddyfile-puhdistus ennen cutoveria (arkkitehtuuripäätös 2026-07-25):** Go on
oletusstack apexissa; Next.js-vertailubuildi hostataan natiivisti Vercelissä
subdomainissa `next.karotammela.fi` (ei co-lokaatio CX23:llä, ei proxy Caddyn
läpi). Siksi Caddyfilestä on poistettu `tech`-evästereititys ja
`/__compare/next/ping` → `localhost:3000` -säännöt — Caddy proxyttaa vain Go:hon
(`localhost:8080`). Päivitä hostin Caddyfile (`caddy validate` + `systemctl
reload caddy`) osana tätä vaihetta. Vertailu mitataan selaimen direct
client-side -pingeillä (`karotammela.fi/api/ping` vs `next.karotammela.fi/api/ping`),
ei Caddy-evästeellä.

1. Laske `karotammela.fi`:n TTL 300 sekuntiin **≥24h etukäteen** (Cloudflare
   DNS-hallinnasta).
2. Varmista paikallisesti hostilla: `curl -s http://127.0.0.1:8080/api/ping`
   toimii.
3. Osoita A- (ja AAAA-, jos käytössä) -tietue CX23:n IP-osoitteeseen,
   **proxy-tila "DNS only"** (harmaa pilvi, ei oranssi).
4. Odota Caddyn automaattinen ACME-sertifikaatti:
   `journalctl -u caddy -f` — katso onnistunut issuance-viesti.
5. Varmista: `curl -sI https://karotammela.fi` (200, HSTS-header),
   `/__compare/go/ping` vastaa, FI/EN-sivut renderöityvät.
6. Vasta viikon vakauden jälkeen: nosta HSTS `max-age` Caddyfilessä ja
   palauta TTL normaaliksi.

**RAPORTOI:** milloin DNS vaihdettiin, että proxy-tila on "DNS only",
`curl -sI https://karotammela.fi` -tuloste, ja mahdolliset virheet
`journalctl -u caddy`:sta.

**Tila:** ✅ Apex `karotammela.fi` → CX23 (`135.181.27.78`, Cloudflare
"DNS only" / harmaa pilvi) 2026-07-25. ✅ Caddy hankki **tuotanto**-Let's
Encrypt -sertifikaatin (`issuer=Let's Encrypt`, `CN=YE1`); `curl -sI
https://karotammela.fi` → `Strict-Transport-Security: max-age=15552000` +
`x-content-type-options: nosniff` (HEAD palauttaa 405, koska `/`-reitti ei salli
HEAD:ia — `GET` antaa 200; ei vika). ✅ Go oletusstackina, `/__compare/go/ping`
→ `{"stack":"go",...}`. **Vaihe 8 valmis.**

**Huomio (staging-CA-ansa):** ensimmäinen issuance tuli virheellisesti
STAGING-CA:sta (`acme-staging-v02...`, jonka selaimet hylkäävät). Korjattu
pingaamalla `acme_ca https://acme-v02.api.letsencrypt.org/directory` +
`email` Caddyfile-globaaliin, poistamalla staging-sertifikaatti Caddyn
tallennuksesta ja käynnistämällä Caddy uudelleen. Tarkista aina `openssl
s_client ... | openssl x509 -noout -issuer` ennen kuin julistat TLS:n valmiiksi.

---

## Vaihe 9 — Next.js-vertailu Vercelissä (erillinen, ei estä cutoveria)

Päivitetty arkkitehtuuri 2026-07-25: Go on oletusstack apexissa (Vaihe 8).
Next.js-vertailubuildi **ei co-lokoidu CX23:llä**, vaan hostataan natiivisti
Vercelissä omassa subdomainissa `next.karotammela.fi`. Näin vältetään epäreilu
verkkohyppy (Caddy-proxy) ja mitataan oikea "Self-hosted Go VPS (Helsinki) vs
Vercel Edge" -ero. Caddy-puoli on jo puhdistettu (Vaihe 8).

Toteutettavat kohdat (Next.js-repo, `src/`):
- **Subdomain Vercelissä:** lisää `next.karotammela.fi` Vercelin custom
  domainiksi (CNAME → `cname.vercel-dns.com`); Vercel hoitaa sertifikaatin.
- **`GET /api/ping` (uusi reitti):** palauta `{ "status": "ok", "stack": "next" }`,
  `Cache-Control: no-store`, ja **CORS** sallien `https://karotammela.fi` sekä
  `https://next.karotammela.fi`. Tämä puuttui aiemmin (404).
- **Go-pingin CORS:** `goth`-puolen `GET /api/ping` on lisättävä
  `Access-Control-Allow-Origin` sallimaan `https://next.karotammela.fi`, jotta
  Next.js-sivun widget voi pingata Go:ta client-side.
- **Tech Switcher:** vaihda evästepohjainen reititys cross-origin-
  uudelleenohjukseksi. "Next.js" → `https://next.karotammela.fi{path}{query}`
  (säilytä polku + kieli), "Go" → `https://karotammela.fi{path}{query}`. Caddyn
  `tech`-evästereititys on poistettu Vaiheessa 8.
- **Perf-widget:** selaimen direct `fetch()`-pingit molempiin `/api/ping`-
  osoitteisiin (Go `https://karotammela.fi/api/ping`, Next
  `https://next.karotammela.fi/api/ping`); mittaa TTFB suoraan selaimesta.
  `NEXT_PING_URL` (server-side probe) jää väliaikaiseksi tueksi, kunnes widget
  on migroitu.

**RAPORTOI (kun päätät tehdä tämän):** haluatko että toteutan nämä Next.js- ja
Go/CORS-muutokset, ja millä aikataululla (esim. heti cutoverin jälkeen).

**Tila (koodi valmis 2026-07-25):** ✅ Next.js `GET /api/ping`
(`src/app/api/ping/route.ts`) palauttaa `{"status":"ok","stack":"next"}`,
`Cache-Control: no-store` ja CORS-otsakkeet molemmille vertailuorigineille.
✅ Go `GET /api/ping` (`internal/handler/handler.go`) lisää
`Access-Control-Allow-Origin`:n (peilaa sallitun Originin: apex + subdomain),
`Vary: Origin`, `Cache-Control: no-store` ja OPTIONS-preflightin. ✅ Tech
Switcher molemmilla stackeilla vaihtaa cross-origin-uudelleenohjauksella
(polku+query+hash säilyy, ei evästettä) — Go: `internal/view/.../footer.html`,
Next.js: `src/components/tech-comparison.tsx` (mountattu locale-layoutiin). ✅
Perf-widget pingaa molemmat `/api/ping`-osoitteet suoraan selaimesta ja näyttää
TTFB:n mediaanin (`n/a` jos ei saatavilla, ei valheellista 0:aa). Originit ovat
konfiguroitavissa: Go `NEXT_URL`/`APP_URL`, Next.js
`NEXT_PUBLIC_GO_ORIGIN`/`NEXT_PUBLIC_NEXT_ORIGIN` (oletukset apex +
`next.karotammela.fi`).

**Jäljellä (manuaalinen, ei koodia):**
1. Cloudflare: lisää CNAME `next.karotammela.fi` → `cname.vercel-dns.com`.
2. Vercel: lisää `next.karotammela.fi` custom domainiksi (Vercel myöntää TLS:n).
3. Deployaa Next.js-buildi Verceliin ja varmista, että
   `https://next.karotammela.fi/api/ping` vastaa julkisesti.

---

## Vaihe 10 — Viikon jälkeen

- Nosta HSTS `max-age` Caddyfilessä alustavasta (6 kk) pidemmäksi kun kaikki
  on ollut vakaa viikon.
- Palauta DNS-TTL normaalille tasolle.

**RAPORTOI:** kaikki vakaana viikon ajan → julkaisu katsotaan valmiiksi.

---

## Yhteenveto — mitä tarvitsen sinulta matkan varrella

1. Palvelimen IP + SSH-vahvistus (Vaihe 1)
2. Caddy-versio + validate-tulos (Vaihe 2)
3. Vahvistus salaisuuksien täytöstä (Vaihe 3)
4. R2-bucketin nimi + synkin onnistuminen (Vaihe 4)
5. deploy.sh-tuloste + ping-vastaus (Vaihe 5)
6. Restore-drillin tulos (Vaihe 6)
7. Timerien tila + lokit (Vaihe 7)
8. DNS-cutoverin ajankohta, proxy-tila "DNS only" -vahvistus, tulokset (Vaihe 8)
9. Next.js-tuotantoajon suunnitelma, jos/kun etenet siihen (Vaihe 9)
