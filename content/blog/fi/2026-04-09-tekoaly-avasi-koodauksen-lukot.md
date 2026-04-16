---
title: "Miten tekoäly avasi koodaukseni lukot parissa vuodessa"
description: "Matkani puisevasta Python-opiskelusta usean samanaikaisen projektin kehittäjäksi tekoälyagenttien ja uudenlaisen työtavan avulla."
publishedAt: "2026-04-09"
slug: "tekoaly-avasi-koodaukseni-lukot"
draft: false
tags: ["koodaus", "ai", "web-kehitys", "yrittäjyys"]
---

Matkani koodaamisen parissa alkoi perinteisesti: kävin muutamia Python-kursseja ja opettelin ohjelmoinnin perusteita. Rehellisyyden nimissä on kuitenkin sanottava, että koodaaminen ja sen opiskelu tuntui tuolloin puisevalta. Teoriapainotteinen harjoittelu ei oikein ottanut tuulta alle, ja tekeminen jäi helposti junnaamaan paikalleen, kun motivaatio katosi syntaksin ja harjoitustehtävien uumeniin.

Tilanne muuttui, kun Gemini 2.0 tuli markkinoille. Se tarjosi tavan siirtyä suoraan perusteiden kertaamisesta käytännön rakentamiseen. Enää koodaaminen ei ollut vain sääntöjen opettelua, vaan pääsin ratkomaan oikeita teknisiä haasteita suoraan projektitasolla. Tämä muutos teki oppimisesta tavoitteellista ja huomattavasti palkitsevampaa.

### Copy-paste koodauksella alkuun

Ensimmäinen konkreettinen projektini oli PHP-pohjainen Green Fee -myyntialusta. Toteutustapa oli suoraviivainen, mutta hidas ja virhealtis: keskustelin Geminin kanssa sovelluksen tarvitsemasta logiikasta, otin sen ehdottamat PHP-pätkät ja siirsin ne suoraan cPanelin tiedostonhallintaan. Integroitavana oli MariaDB-tietokanta ja Stripe-maksunvälitys. Kehitin koodaukseen systeemin, jolla geminin konteksti-ikkunan täyttyessä, minulla oli tekstipohjainen "siirtotiedosto", jonka avulla siirsin edellisen keskustelun tehdyt työt seuraavaan keskusteluun ja liitin mukaan oleellisimmat kooditiedostot. Huomasin nopeasti, että tekoäly alkoi sekoilla omiaan, kun konteksti-ikkunaan tuli riittävän paljon sisältöä.

Vaikka alkuun liikuttiin vahvasti kopioimalla ja kokeilemalla, prosessi toimi yllättävän hyvänä opettajana. Kun näin sovelluksen rakentuvan ja virheiden korjaantuvan keskustelun kautta, ymmärrykseni koodin rakenteesta kasvoi huomattavasti nopeammin kuin pelkillä verkkokursseilla. Opin hahmottamaan, miten tietokantakyselyt, rajapinnat ja käyttöliittymä toimivat yhteen, vaikka en ollut vielä ehtinyt kirjoittaa jokaista riviä itse alusta asti. Ymmärrykseni kasvoi tuossa vaiheessa suorastaan eksponentiaalisesti.

### Kaksi vuotta takana – PHP:sta moderniin stackiin

Tuosta ensimmäisestä kokeilusta on nyt kulunut kohta pari vuotta. Ehkä noin 1,5. Nykyään koodaaminen on kiinteä osa arkeani, ja käytän siihen noin 20–30 tuntia viikossa vapaa-ajallani. Matkan varrella olen löytänyt itselleni toimivan tech stackin, joka mahdollistaa useiden projektien hallinnan yhtä aikaa ilman, että paketti hajoaa käsiin.

Työpöydälläni on tällä hetkellä useita aktiivisia projekteja:

* **Next.js-pohjainen myyntialusta:** Olen rakentamassa alkuperäisestä Green Fee -ideasta uutta versiota Next.js:llä, mikä on tuonut mukanaan paljon uutta oppia modernista web-kehityksestä.
* **Mökin varaussivusto:** Tämä projekti on jo valmis ja käytössä, ja se opetti valtavasti varausjärjestelmien logiikasta ja reaktiivisista käyttöliittymistä.
* **Tukkuliikkeen verkkokauppa:** Rakennan räätälöityjä tuotteita myyvälle tukkuliikkeelle omaa B2B verkkokauppaa, jossa keskitytään erityisesti tilausprosessin sujuvuuteen ja monimutkaiseen hinnoittelulogiikkaan. 

Tekeminen on muuttunut satunnaisista kokeiluista tavoitteelliseksi kehittämiseksi, jossa jokainen uusi projekti opettaa jotain uutta arkkitehtuurista ja skaalautuvuudesta.

### Agentic AI – keskustelua tiimin kanssa

Työtapani ovat kehittyneet huomattavasti alkuajoista. Nykyään puhun mieluummin agentic AI -koodauksesta. Koodatessa tuntuu siltä, että keskustelen osaavien työkavereiden kanssa sen sijaan, että antaisin koneelle vain mekaanisia komentoja. 

Olen oppinut hallitsemaan kielimallien konteksti-ikkunaa niin, että projektin toteutus pysyy tehokkaana. Promptaustaitoni ovat kasvaneet tasolle, jossa minun ei juurikaan tarvitse enää käyttää aikaa sen varmistamiseen, ymmärtävätkö agentit tarkoitukseni oikein. Tässä on auttanut sekin, että kielimallit ovat ottaneet tänä aikana huimia harppauksia eteenpäin. Käytän eri kielimalleja eri tarkoituksiin niiden vahvuuksien mukaan; tällä hetkellä pääasiallisena työkalunani on **GPT Codex 5.3**. Se toimii luotettavana kumppanina, kun rakennan monimutkaisempia kokonaisuuksia ja hion sovellusten logiikkaa. Kevyempiin hommiin taas edullisemmat kielimallit ovat minulla ahkerassa käytössä.

### Oppiminen ei ole aidosti koskaan ollut näin nopeaa ja hauskaa

Vaikka takana on jo paljon tehtyjä tunteja, matka on vasta alussa. Oppimiskäyrä on pysynyt todella loivana, ja uuden omaksuminen on nykyisillä työkaluilla nopeampaa kuin koskaan aiemmin. Se, mikä tuntui alussa puuduttavalta, on nyt muuttunut koukuttavaksi ongelmanratkaisuksi. Olen pentuna tykännyt pelata erilaisia maailmojen rakennuspelejä, kuten Civilization, Command&Conquer, Theme Park, A-train, jne. Tämä nykyinen tapa koodata on kuin ne pelit tosielämässä. Sen sijaan, että rakennan kuvitteellista pelimaailmaa, rakennankin tosielämän sovelluksia. Se on äärimmäisen palkitsevaa. Kuulen tämän saman melkein jokaisessa podcastissa tai blogissa, joita luen agentic AI -koodaukseen tai kielimalleihin liittyen. Moni koodari on löytänyt taas palon ohjelmistojen rakentamiseen, kun tekoäly tekee kaiken tylsän työn ja voi itse keskittyä suureen kuvaan ja haastaviin ongelmiin. 

### Osaamiseni ajan tasalla

Pidän osaamiseni ajan tasalla lukemalla teknologia-blogeja, kuuntelemalla podcasteja ja käyttämällä Geminiä apuna uusien teknologioiden opiskelussa. Kasvatan ymmärrystäni jatkuvasti, ja jokainen uusi kirjasto tai työkalu on helpompi ottaa haltuun kuin edellinen. Nykyiset työkalut tarjoavat valtavasti mahdollisuuksia, ja on kiinnostavaa nähdä, miten paljon ammattitaitoni kehittyy tästä eteenpäin, kun tekeminen on päivittäistä ja tavoitteellista. Matka on todellakin vasta alussa.
