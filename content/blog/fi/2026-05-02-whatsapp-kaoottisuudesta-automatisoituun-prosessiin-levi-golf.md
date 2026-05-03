---
title: "WhatsApp-kaaoksesta automatisoiduksi prosessiksi: näin rakensin Levi Golfin Green Fee -myyntialustan"
description: "Miten manuaalinen pelilippujen myynti kehittyi turvalliseksi, monikieliseksi ja usean myyjän alustaksi Agentic AI:n avulla."
publishedAt: "2026-05-02"
slug: "whatsapp-kaoottisuudesta-automatisoituun-prosessiin-levi-golf"
draft: false
tags: ["koodaus", "ai", "golf", "automaatio", "saas"]
---

Jos on hallinnoinut kovalla kysynnällä toimivaa palvelua, ovat manuaaliset työvaiheet tyypillisesti aikaa vieviä ja inhimillisen virheen riski on todellinen. Minulle Levi Golfin green fee -pelilippujen myynti oli pitkään juuri tätä: puheluita, pirstaleisia WhatsApp-viestejä ja jatkuvaa Excel-tetristä, jotta pysyi kartalla siitä, mitä osakkeita oli saatavilla millekin päivälle. Omistan yhtiön kautta useita Levi Golfin osakkeita, joihin jokaiseen sisältyy pelikaudelle 35 pelilippua. Kauden aikana ostotapahtumia on kymmeniä, jopa satoja. 

Tuplabuukkauksen riski oli todellinen, ja hallinnollinen säätäminen oli lähes päivittäistä. Halusin järjestelmän, jossa asiakas voi hoitaa kaiken itse: valita päivän, maksaa turvallisesti ja saada vahvistuksen heti ilman, että minun täytyy tehdä juuri lainkaan manuaalista työtä.

Näin rakensin [Levi Golfin Green Fee -sovelluksen](https://greenfee.levifinland.fi) Agentic AI -työnkululla.

### Ydinhaaste: osakepohjainen allokointilogiikka

Vaikein osa ei ollut maksujen vastaanotto, vaan liiketoimintalogiikka. Tässä mallissa saatavuus sidotaan osakkeisiin. Jokaisella osakkeella on 35 pelilippua per kausi, ja yksi tiukka sääntö: yhdestä osakkeesta voidaan myydä vain yksi lippu per päivä. Teoriassa lippuja voisi myydä kerran 4,5 tunnissa, mutta käytännössä en pysty hallitsemaan asiakkaiden tiiaikojen varauksia, joten lippujen myynti oli parempi rajoittaa yhteen lippuun per päivä per osake.

Jotta ratkaisu olisi luotettava, toteutin relaatiopohjaisen arkkitehtuurin Drizzle ORM:llä ja SQLite:llä. Pelkkä jäljellä olevien lippujen laskenta ei riitä, vaan järjestelmä seuraa jokaisen osakkeen tilaa erikseen. Kun asiakas valitsee päivän, sovellus tekee reaaliaikaisen saatavuustarkistuksen ja palauttaa vain ne osakkeet, jotka ovat juuri sille päivälle kelvollisia.

Kun Stripe vahvistaa maksun, alusta allokoi varauksen automaattisesti ensimmäiselle mahdolliselle osakkeelle. Tämä takaa yhden lipun per osake per päivä -säännön toteutumisen ja poistaa inhimilliset virheet sekä ylibuukkauksen riskin.

![Green Fee -myyntialusta](/media/greenfee-levifinland-fi.png)

### Turvallisuus, vakaus ja admin-hallinta

Koska järjestelmä käsittelee maksuja ja asiakasdataa, turvallisuudesta ei voi tinkiä. Frontend on tarkoituksella kevyt, ja kaikki kriittinen logiikka tapahtuu palvelinpuolella: maksun validointi, osakeallokointi ja transaktionaalinen viestintä.

Stripe-webhookit ovat arkkitehtuurin ytimessä. Pelilippu luodaan vasta, kun Stripe vahvistaa onnistuneen maksutapahtuman. Näin järjestelmä suojautuu valetapahtumilta ja race condition -reunatapauksilta.

Rakensin myös täysimittaisen Admin Dashboardin skaalautuvuus edellä. Alusta tukee uusien osakkeenomistajien lisäämistä myyjiksi, jolloin jokainen myyjä voi hallita omia osakkeitaan ja seurata omaa myyntiään koskematta ydinkoodiin. Näin henkilökohtaisesta työkalusta tuli aidosti usean myyjän myyntialusta.

![Stripe-maksupolku](/media/greenfee-maksaminen-levifinland-levi-golf.png)

### Viestintä ja lokalisointi

Varausprosessi ei ole valmis ilman välitöntä ja luotettavaa viestintää. Poistaakseni manuaaliset kyselyt ja viesetinnän, integroin Resend-rajapinnan. Heti maksun onnistuttua järjestelmä lähettää vahvistussähköpostin, jossa on maksutapahtuman tiedot sekä pelilippujen numerot caddiemasterin toimistolla esitettäväksi.

Levi Golf palvelee kansainvälistä asiakaskuntaa, joten sovellus on lokalisoitu täysin viidelle kielelle: suomi, englanti, ranska, saksa ja hollanti. Kokemus pysyy natiivina ja yhtenäisenä aloitussivulta vahvistussähköpostiin asti.

### Rakennusmetodi: Agentic AI ja mallien orkestrointi

Projektin kiinnostavin osa oli tapa, jolla sen rakensin. Koodasin sivuston Agentic AI -menetelmällä [Kilo Coden](https://kilo.ai) avulla.

Perinteiseen chat-pohjaiseen koodaukseen verrattuna agentti toimii kuin ohjelmistoinsinööri: se lukee koodikantaa, seuraa suunnitelmia, päivittää tiedostoja, ajaa tarkistuksia ja iteroi korjauksia. Tämä paransi sekä toteutusnopeutta että luotettavuutta.

Suurin hyöty tulee mallien orkestroinnista. Käytin eri malleja eri vastuualueisiin:

1. **Rakentaja (GPT Codex 5.3):** ensisijainen toteutusmoottori React-käyttöliittymään, Next.js-reitteihin ja nopeaan ominaisuuksien toimitukseen. Malli tarjoaa erittäin hyvän hinta-laatusuhteen.
2. **Auditoija (Opus 4.6 & 4.7):** validointi tietoturvakriittiseen logiikkaan, transaktioiden eheyteen ja webhook-idempotenssin reunatapauksiin. Opus on erinomainen ja luotettava työjuhta, mutta kustannus on selvästi korkeampi kuin lähes millään muulla kielimallilla.
3. **Copywriter (Gemini 3.1 Flash Lite Preview):** monikielinen sisällöntuotanto ja sävyn hienosäätö viidelle kielelle. Malli on nopea, edullinen ja sopii tehtäviin, jotka eivät vaadi syvää teknistä ymmärrystä.

Tämä roolipohjainen toimintamalli tuottaa parempia tuloksia kuin yhden mallin pakottaminen kaikkeen ja hyödynnän sitä muissakin projekteissani. 

### Lopputulos: manuaalisen työn loppu

Operatiivinen ero on dramaattinen. Tehtävät, jotka aiemmin veivät tunteja manuaalista koordinointia, hoituvat nyt automaattisesti: saatavuustarkistukset, maksun vahvistus, lipun allokointi, asiakasviestintä ja myyjätason seuranta.

En enää käytä aikaa taulukoiden ja viestiketjujen täsmäyttämiseen. Alusta on nyt turvallinen, skaalautuva ja omavarainen. En pitänyt kirjaa tähän projektiin käytetyistä työtunneista, sillä se on ollut osa laajempaa [levifinland.fi](https://levifinland.fi) -projektia. Tämä green fee myyntialusta on yksi applikaatio neljästä, jotka olen tähän samaan Turborepo-projektiin rakentanut. Se esimerkiksi jakaa saman hallintapaneelin kuin mökkien vuokrausalusta. Olin jo vuonna 2025 rakentanut ensimmäisen version green fee myyntialustasta. Silloin vasta harjoittelin agentic AI ohjelmointia, eikä minulla ollut vielä kokemusta Next.js:stä, libSQL:stä tai monorepostakaan. Se oli toimiva järjestelmä, mutta tämä nykyinen versio on nykyaikainen, tietoturvallisempi ja sen skaalaaminen on helpompaa.

Manuaalinen aikakausi lippujen myymiseksi on osaltani virallisesti ohi.
