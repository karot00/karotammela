---
title: "Oma tekoälylaboratorio olohuoneessa: Kaksi näytönohjainta, Proxmox ja paikallinen LLM taskussani"
description: "Miksi rakensin kotiini kahden GPU:n tekoälypalvelimen, miten Proxmox, Ollama ja Qwen 3.8 27B toimivat arjessa ja miltä tuntuu ajaa omaa AI-chattia puhelimesta ilman pilvikuluja."
publishedAt: "2026-08-16"
slug: "kotilabra-ja-qwen-3-8-27b"
draft: false
tags: ["tekoäly", "kotilabra", "proxmox", "ollama", "hardware", "koodaus"]
---

Viime ajat tekoälymaailmassa ovat olleet melkoista hulinaa. Pilvipalveluiden jatkuviin kuukausimaksuihin ja API-laskuihin kyllästyneenä päätin ottaa ohjat omiin käsiini. Halusin rakentaa kotiini täysin omavaraisen tekoäly- ja kehityspalvelimen, joka pyörittää raskaita kielimalleja lokaalisti, palvelee etätyöpöytänä ja kulkee tarvittaessa mukanani suoraan puhelimen ruudulla.

Tässä artikkelissa käyn läpi, miten viikonlopun mittainen vapaa-ajan projektini eteni raudan ruuvaamisesta aina oman tekoälychatin käyttöönottoon saakka.

### 1. Rauta kasaan: Miksi päädyin kahteen RTX 5060 Ti -näytönohjaimeen?

Käytin kaksi iltaa uuden PC-kokoonpanoni kasaamiseen ja kaapelointiin. Kokoonpanon ehdoton kulmakivi on näytönohjainosasto: asensin koneeseen kaksi Nvidia RTX 5060 Ti 16 GB -korttia.

Tekoälymaailmassa eli suurten kielimallien (LLM) ajossa GPU:n raaka laskentateho ei yksin riitä – kaikkein kriittisin pullonkaula on näytönohjaimen videomuisti eli VRAM. Mallit pitää pystyä lataamaan kokonaisuudessaan näytönohjaimen muistiin, jotta vastausten generointi eli inferenssi pysyy nopeana.

Yksi 16 gigatavun näytönohjain käy nopeasti ahtaaksi, kun siirrytään järeämpiin 27–32 miljardin parametrin malleihin. Kaksi 16 GB -korttia rinnan tarjoaa yhteensä 32 GB VRAM-muistia. Tämä mahdollistaa raskaampien mallien ja laajojen kontekstien ajamisen ilman, että dataa joudutaan valuttamaan huomattavasti hitaampaan keskusmuistiin (RAM). Lisäksi kaksi erillistä korttia antaa joustavuutta virtuaaliympäristöjen resurssijakoon.

### 2. Proxmox VE: Perusta kahdelle yhtäaikaiselle käyttöjärjestelmälle

Raudan valmistuttua asensin koneen pohjalle Proxmox VE:n (Virtual Environment). Proxmox on Type-1 Hypervisor eli palvelintason hypervisori, joka asennetaan suoraan "paljaalle metallille" ilman erillistä isäntäkäyttöjärjestelmää. Sen avulla voin jakaa fyysisen koneeni resursseja eristetyille virtuaalikoneille (VM).

Loin Proxmoxiin kaksi erillistä virtuaalikonetta:

- **Windows 11 (VM 200):** Hyödyllinen peruskäyttöön, testaukseen ja tiettyihin erikoisohjelmiin.
- **Pop!_OS Linux (VM 100):** Nvidian ajureilla valmiiksi optimoitu Ubuntu-pohjainen Linux-jakelu, joka toimii pääasiallisena kehitysympäristönäni ja tekoälymoottorinani.

Näytönohjainten PCI Passthrough -tekniikan ansiosta pystyn ohjaamaan fyysiset näytönohjaimet suoraan virtuaalikoneille. Voin ajaa Windowsia ja Pop!_OS:ää yhtäaikaisesti niin, että molemmilla järjestelmillä on oma rautakiihdytetty näytönohjaimensa käytettävissään.

### 3. Etäkäyttö sulavaksi: Sunshine + Moonlight

Aina ei huvita istua työhuoneessa fyysisen palvelimen ääressä, joten halusin ohjata järjestelmää vaivattomasti kannettavaltani mistä tahansa huoneesta. Perinteinen Windows RDP tai Linuxin VNC ovat koodaamiseen ja aktiiviseen työpöytäkäyttöön usein liian tahmeita ja nykiviä.

Ratkaisuksi valikoitui avoimen lähdekoodin tehokas parivaljakko: **Sunshine** ja **Moonlight**.

- **Sunshine** toimii palvelimelle asennettuna isäntäohjelmana. Se kaappaa työpöytäkuvan ja pakkaa sen Nvidian rautapohjaisella NVENC-enkooderilla suoraan videovirtaksi lähes ilman prosessorikuormitusta.
- **Moonlight** toimii asiakasohjelmana kannettavallani.

Tällä yhdistelmällä Pop!_OS:n työpöytä striimautuu kotiverkossani kannettavani ruudulle täydellä 60–120 fps nopeudella ja käytännössä huomaamattomalla viiveellä. Käyttötuntuma on täysin sama kuin istuisi suoraan tehomyllyn ääressä.

### 4. Tekoälymoottori käyntiin: Ollama, Qwen 3.8 27B ja konteksti-ikkunat

Sitten päästiin itse asiaan eli paikallisen kielimallin ajamiseen. Käytän tekoäly-ympäristön hallintaan Ollamaa, joka toimii kuin "Docker kielimalleille" – se lataa, määrittelee ja ajaa avoimen lähdekoodin malleja kevyesti suoraan terminaalista.

Pääasialliseksi malliksi valikoitui Alibaban suosittu **Qwen 3.8 27B** (27 miljardin parametrin malli). Se tarjoaa kokoluokassaan poikkeuksellisen tarkkaa loogista päättelykykyä ja erinomaista koodausapua.

Koska 32 GB VRAM-muistikapasiteetti antaa mukavasti liikkumavaraa, loin Ollaman `Modelfile`-konfiguraatioilla mallista kolme eri variaatiota eri konteksti-ikkunoilla (*Context Window*):

- **16k konteksti:** Erittäin kevyt ja salamannopea lyhyisiin kyselyihin sekä pieniin koodinpätkiin.
- **32k konteksti:** Täydellinen tasapaino ja arjen sweet spot: malli muistaa kymmeniä sivuja koodia tai dokumentaatiota kerralla.
- **64k konteksti:** Raskassarjalainen laajojen tiedostokokonaisuuksien analysointiin (vie lähes koko käytettävissä olevan VRAM-muistin).

### 5. Koodausta Kilo Codella ja ensikosketus TTFT-viiveeseen

Otin kehitysympäristökseni VS Coden ja siihen asennetun Kilo Code -agenttilaajennuksen. Se toimii samaan tapaan kuin tekoälyavusteiset kehitystyökalut yleensä, mutta reitittää API-kutsut pilvipalveluiden sijaan suoraan omaan lähiverkkooni paikalliselle Ollama-instanssille.

Koodatessa huomasin paikallisen 27B-mallin tyypillisen ominaispiirteen: TTFT-viiveen (*Time to First Token*). Kun mallille syöttää laajan kooditiedoston, näytönohjaimilta kuluu muutama sekunti syötteen pureskeluun ja kontekstin valmisteluun muistissa. Mutta heti kun tämä esikäsittely on valmis, vastausta ja puhdasta koodia alkaa syntyä ruudulle hämmentävän ripeästi.

### 6. Testi tositilanteessa: Oman chat-sovelluksen koodaus kahdella kehotteella

Testatakseni laitteistoni todellista suorituskykyä annoin paikalliselle Qwen 3.8 -mallilleni käytännön tehtävän: *"Koodaa minulle toimiva Web Chat -sovellus, joka käyttää taustamoottorinaan palvelimellani pyörivää Ollaman API-rajapintaa."*

Lopputulos veti hiljaiseksi. Tekoäly kirjoitti sovelluksen taustajärjestelmän (backend), käyttöliittymän (frontend) sekä tarvittavan virheenkäsittelyn valmiiksi käytännössä kahdella peräkkäisellä promptilla. Ei ylimääräistä bugien viilausta tai manuaalista säätämistä – vain toimivaa koodia suoraan ajoon.

### 7. Viimeinen silaus: Tailscale ja tekoäly taskussa

Mitä hyötyä on tehokkaasta kotipalvelimesta, jos sitä voisi käyttää vain olohuoneen sohvalta? Ratkaisin turvallisen etäpääsyn asentamalla laitteisiin Tailscalen.

Tailscale on äärimmäisen suoraviivainen ja turvallinen, WireGuard-protokollaan perustuva Mesh VPN -ratkaisu. Se yhdistää eri laitteeni samaan suojattuun virtuaaliseen lähiverkkoon ilman reikien puhkomista kotireitittimen palomuuriin tai julkisten IP-osoitteiden avaamista.

Nyt kokonaisuus toimii saumattomasti:
Kun avaan puhelimestani Tailscale-yhteyden ja suuntaan selaimella kotipalvelimeni sisäiseen IP-osoitteeseen, ruudulle aukeaa oma chat-sovellukseni. Voin kysyä siltä asioita tai pyytää apua koodaukseen missä päin maailmaa tahansa. Vastauksen generoi ja laskee olohuoneessani huriseva kahden RTX-kortin palvelin. Data ei kierrä minkään ulkopuolisen yrityksen pilvessä, eikä kuukausimaksuja mene senttiäkään.

### Yhteenveto ja arvosanat

Tämä projekti oli yksi opettavaisimmista ja palkitsevimmista pitkään aikaan. Laite- ja ohjelmistopalat loksahtivat paikoilleen tavalla, joka osoitti konkreettisesti paikallisen tekoälyn kypsyyden vuonna 2026.

Jos viikonlopun saavutuksista pitäisi antaa arvosanat, ne ovat poikkeuksetta täydet:

- ⭐️⭐️⭐️⭐️⭐️ **Proxmox VE:** Maailman joustavin alusta virtuaalikoneiden hallintaan ja GPU-passthrough'n toteutukseen.
- ⭐️⭐️⭐️⭐️⭐️ **Qwen 3.8 27B:** Avoimen lähdekoodin mallien ehdotonta kärkeä. Koodaa tarkasti ja hahmottaa suuria kokonaisuuksia.
- ⭐️⭐️⭐️⭐️⭐️ **2x RTX 5060 Ti 16GB:** 32 GB VRAM-muistia tähän hintaluokkaan on ehdoton sweet spot kenelle tahansa omasta tekoälylaboratoriosta kiinnostuneelle.

---

### Kotipalvelimen komponentit

Tässä vielä referenssiksi koneen tarkka osalista:

- **Näytönohjaimet (2 kpl):** PNY GeForce RTX 5060 Ti 16GB GDDR7 (yhteensä 32 GB VRAM)
- **Prosessori:** AMD Ryzen 7 7700
- **Prosessorijäähdytin:** Thermalright Phantom Spirit 120 SE
- **Emolevy:** ASUS ProArt B850-CREATOR WIFI NEO
- **Keskusmuisti (RAM):** Patriot 32GB (2x16GB) DDR5 6000MHz CL30
- **Tallennustila (SSD):** Sandisk / WD Blue SN5100 1TB NVMe M.2 SSD (PCIe 4.0)
- **Virtalähde:** MSI MPG A1000G PCIE5 1000W Gold (ATX 3.1)
- **Kotelo:** Lian Li LANCOOL 216 RGB (erinomainen ilmankierto)
