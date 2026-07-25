---
title: "Vibe coding is easy – Production ready is hard"
description: "Miksi tekoälyllä koodaaminen on helppoa aloittaa, mutta vaikeaa hallita?"
publishedAt: "2026-04-11"
slug: "vibe-coding-vs-production-ready"
draft: false
tags: ["coding", "ai", "agentic-workflow", "web-development"]
---

Tekoäly on tehnyt koodaamisen aloittamisesta naurettavan helppoa. Työkalut kuten Lovable ja erilaiset "vibe-koodauksen" alustat ovat mahtavia, koska ne poistavat kynnyksen teknisen osaamisen ja näkyvän lopputuloksen väliltä. Voit vain nopeasti tai hetken mielijohteesta kuvailla idean ja sinulla on valmis prototyyppi tai ainakin jollain tapaa toimiva tuote käsissäsi.

Mutta tässä piilee ansa. Jos tyytyisi pelkkään vibe-koodaukseen, huomaa pian rakentavansa kertakäyttötavaraa. Se on hauskaa leikkimistä, mutta kun sovelluksen pitäisi oikeasti käsitellä rahaa, tallentaa monimutkaisia relaatioita tai skaalautua, pelkkä hyvä fiilis ja nätti ulkokuori eivät enää riitä. Sen lisäksi, että ne sisältävät loputtomasti bugeja ja haavoittuvuuksia, niin tekoäly ei kuitenkaan auta korvaamaan ihmisajattelua ja suunnittelua.

Oma polkuni on sisältänyt paljon tätä testailevaa YOLO-modea. Olen kuitenkin oppinut kantapään kautta, että jos antaa tekoälyn vain täydentää rivejä ilman suunnitelmaa (autocomplete-ansa), päätyy hyvin nopeasti syvälle suohon. Siellä odottaa tuntikausien debuggaus, kun korttitalo eli näennäisesti toimivan näköinen sovellusviritelmä alkaa vähitellen sortua.

### Mockup on eri asia kuin tuotantokoodi

Käytän itsekin vibe-koodausta, mutta se on vain yksi harkittu osa kokonaisuutta. Vibeltäminen on loistava tapa rakentaa nopeita UI-mockupeja. Saatan keskustella agentin kanssa visuaalisesta ilmeestä ja hioa käyttöliittymän sellaiseksi, että se miellyttää silmää. Tässäkin agentti toimii työkaluna ja osaavana CSS:n kirjoittajana, mutta se ei aidosti tunne, miltä se ihmiselle näyttäytyy. 

Kun mockup on valmis, siirrän sen "plans"-kansiooni. Tärkeä vaihe on kuitenkin tässä: merkkaan sen selkeästi vain koodiesimerkiksi. Kiellän koodausagenttiani viemästä sitä sellaisenaan tuotantoon. Miksi? Koska siitä puuttuvat todelliset relaatiot ja logiikka. On huomattavasti helpompaa rakentaa aito toiminnallisuus yhdellä kertaa järkevän suunnitelman päälle, kuin yrittää nikkaroita monimutkaista backend-logiikkaa pelkän "nätin kuoren" sisään. 

### Agentin ohjaaminen: Implementation Plan ja Progress

Ehkä tärkein oppini on se, miten pysyn itse ohjaksissa. Pidän jatkuvasti ajan tasalla "implementation plan" ja "progress" -tiedostoja. Nämä ovat koodausagentin kartta ja kompassi. Kun vaihdan taskia, agentti lukee nämä tiedostot ja tietää sekunnissa, missä mennään. Ja tarkennuksena siis, että usein pyydän tehtävän päätteeksi koodausagentin päivittämään nämä tiedostot, jonka jälkeen itse vielä käyn ne läpi ennen kuin jatkan projektin seuraavaan vaiheeseen.

Jos työstän aihealuetta, jossa teen vaikkapa ensimmäistä kertaa, otan käyttöön "tuplatsekkauksen". Käytän Gemini Prota validointiin: syötän sille agentin tuottaman koodin ja kysyn siltä toteutusvaihtoehdoista ja koodin laadusta. Tähän lyhyeen keskusteluun uppoaa ajoittain tunti tai pari, kun syvennyn opiskelemaan, jotain toista teknologiaa tai kirjastoa, jonka Gemini on nostanut vaihtoehdoksi. Siirrän noiden keskustelujen annin takaisin Kilo Codeen korjauksina tai muutoksina, jos tarvetta on. Tämä ei ainoastaan varmista laatua, vaan toimii itselleni parhaana mahdollisena koodauskouluna – opin jatkuvasti logiikkaa ja uusia teknologioita, joita en muuten täysin ymmärtäisi tai joista minulla ei vielä 15 minuuttia sitten ollut tietoakaan.

### 3 tuntia ideasta tuotantoon

Halusin testata tätä hallittua työnkulkua käytännössä, kun pystytin omien sivujeni ensimmäisen tuotantoversion. Asetin itselleni tiukan kolmen tunnin aikarajan. Pystyin arvioimaan käytettävän ajan melko tarkasti, sillä olin jo toteuttanut kaikki tarvittavat komponentit aiemmissa projekteissani ja kokeiluissani.

Tuo aikaikkuna sisälsi kaiken olennaisen:
- Domainin osto, DNS-asetusten konfigurointi Cloudflaressa ja GitHub-repon pystytys.
- Tietokannan konfigurointi.
- Varsinainen koodaus ja tekoälyportsarin kouluttaminen.
- Sivujen ja sisältöjen ensimmäisten tuotantoversioiden luominen.
- Sähköpostipalvelun integrointi lomakkeita varten.
- Deployaus Verceliin, .env-tiedostojen tuonti ja domainin liittäminen projektiin.

Rajauksen ulkopuolelle (out of scope) jätin tässä vaiheessa blogiosion varsinaiset sisällöt, kävijäanalytiikan pystytyksen sekä evästeilmoituksen.

Aikarajani osui yllättävän hyvin kohdalleen: heittoa tuli vain muutama minuutti asettamaani kolmeen tuntiin nähden. Käytin erillistä agenttia sisällön generointiin ja viimeistelin tekstit manuaalisesti. Työstän sisältöjä tietysti edelleen, mutta olen todella tyytyväinen ensimmäisen version edistymiseen ja lopputulokseen.

CI/CD-putki (jatkuva integraatio ja julkaisu) on nyt valmis ja rutiininomainen osa tekemistäni. Teknologiastackini on jo jotakuinkin vakiintunut, mutta koska tiedonnälkäni tuntuu tällä hetkellä loputtomalta (asia, joka ei ole missään tapahtunut minulle ennen), niin kokeilen uusia asioita jatkuvasti, testaan eri kielimallien kyvykkyyttä eri tehtäviin ja optimoin token-kustannuksia.

Seuraava iso askel itselleni on viedä automaatio vielä pidemmälle. Tavoitteenani on rakentaa itselleni tähän putkeen järjestelmä, jossa tuotantoversion virhelokit luetaan automaattisesti ja tekoälyagentti ehdottaa tai jopa toteuttaa korjaukset välittömästi. Teknisesti se ei ole enää erityisen haastavaa, mutta virhekorjausten automatisointi vaatii paljon guardraileja, testausta ja kokemusta, jotta uskaltaisin sellaisen pystyttää vielä mihinkään varsinaiseen projektiin. Pitää rakentaa jokin hiekkalaatikkoprojekti, jossa pääsen tuota treenaamaan.

Matka vibe-koodauksesta tähän pisteeseen on ollut nopea, mutta agenttien aikakaudella tuntuu kuin juoksisin koko ajan 10 askelta kehityksen perässä. Pari kuukautta on nykyään pitkä aika ja sinä aikana on ehtinyt nousta kymmenittäin hyödyllisiä open-source projekteja, joita voi ottaa käyttöön. Uusia ja parempia kielimalleja tulee lähes päivittäin. Sen lisäksi pitää myös kuroa kiinni sitä teknisen osaamisen velkaa, jota olen ensimmäiset 40 vuotta elämästäni kerryttänyt.  

Minusta tuskin tuleekaan koskaan koodaajaa sanan varsinaisessa merkityksessä, mutta nyt päätavoitteeni on luoda toimintatapoja, joilla voin ottaa agenttikoodauksen hyödyt turvallisesti ja tehokkaasti käyttöön todellisissa projekteissa ja tosielämän ratkaisuissa. 