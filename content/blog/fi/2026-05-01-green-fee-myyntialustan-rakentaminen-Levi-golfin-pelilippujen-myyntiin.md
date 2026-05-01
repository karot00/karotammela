---
title: "Näin rakensin agentic AI:ta hyödyntäen alustan Levi Golfin pelilippujen myymiseksi"
description: "Miten manuaalinen viestien ja osakkeiden hallinta muuttui täysin automatisoiduksi alustaksi modernin tech stackin ja tekoälytiimin avulla."
publishedAt: "2026-05-01"
slug: "whatsapp-viestittelysta-automaatioon-green-fee"
draft: false
tags: ["koodaus", "ai", "golf", "automaatio", "liiketoiminta"]
---

Omistan yhtiön kautta useita Levi Golfin osakkeita ja hoidan samalla muidenkin omistajien lippujen myyntiä. Kun yhdellä osakkeella on 35 käytettävää pelilippua kaudessa, mutta niitä saa käyttää vain yhden per päivä, manuaalinen hallinta on ollut todella vaivalloista: Whatsapp keskusteluita, excelin tarkistusta ja usean viestin kirjoittamista yhden myyntitapahtuman saamiseksi aikaan. Tuplabuukkausten riski oli todellinen ja asiakaskokemus kaukana nykyaikaisesta.

Päätin rakentaa täysin automatisoidun alustan, jossa asiakas voi hoitaa kaiken itse – valitsee päivän, maksaa turvallisesti ja saa vahvistuksen välittömästi. 

Näin syntyi [greenfee.levifinland.fi](https://greenfee.levifinland.fi).

![Green Fee sales platform](/media/greenfee-levifinland-fi.png)

### Haaste: Pelioikeuksien tiukka logiikka

Järjestelmän tekninen sydän ei ole vain maksujen vastaanottaminen, vaan monimutkaisen säännöstön hallinta. Levi Golfin mallissa saatavuus perustuu osakkeisiin: jokainen osake sisältää 35 lippua per kausi, mutta vain yksi lippu per osake voidaan myydä yhdelle päivälle. Yhteensä minulla on kaudessa siis parisataa myytävää pelilippua Levi Golfiin.

Toteutin tätä varten relaatiotietokanta-arkkitehtuurin, joka skannaa kaikki hallinnassa olevat osakkeet reaaliajassa. Kun maksu vahvistuu Stripen kautta, järjestelmä allokoi lipun automaattisesti ensimmäiselle vapaalle osakkeelle. Se myös tarkistaa, mitä osaketta on eniten jäljellä ja priorisoi niitä osakkeita. Haluan myydä peliliput tasaisesti läpi kauden, jotta en myy yhtä osaketta tyhjäksi, jolloin loppukauden ajaksi olisi vähemmän myytävää per päivä. Tämä poistaa inhimillisen virheen mahdollisuuden ja takaa, ettei osakkeita ylibuukata ja peliliput tulee myös myytyä tasaisesti jokaiselta osakkeelta. 

### Tietokanta ja turvallisuus keskiössä

Valitsin tietokannaksi **Turson (LibSQL)**. Se on hajautettu SQLite-pohjainen ratkaisu, joka mahdollistaa datan pitämisen lähellä käyttäjää (Edge). Tämä takaa salamannopeat latausajat riippumatta siitä, varaako asiakas pelilippunsa Leviltä, Helsingistä vai Amsterdamista.

Turvallisuus on leivottu sisään arkkitehtuuriin:
* **Eristetty data:** Järjestelmä on monen käyttäjän (multi-tenant) alusta. Jokainen osakkeenomistaja näkee dashboardillaan vain omat datansa, analytiikkansa ja osakkeidensa tilan.
* **Stripe Webhookit:** Lipun allokointi tapahtuu vasta, kun Stripen palvelin vahvistaa maksun onnistuneen. Tämä varmistaa, että varausjärjestelmä pysyy synkronoituna todellisen kassavirran kanssa.

### Dashboard – Hallinta ilman vaivaa

Rakensin kattavan hallintapaneelin, jossa osakkeenomistajat voivat seurata 35 lipun kiintiönsä täyttymistä ja ladata raportit kirjanpitoa varten. Integroin mukaan myös **Resend**-sähköpostirajapinnan, joka lähettää asiakkaalle välittömästi vahvistukset ostoksesta ja tiedot siitä, miten pelilipun saa aktivoitua käyttöönsä.

Koska Levi houkuttelee kansainvälisiä matkailijoita, sovellus on täysin lokalisoitu viidelle kielelle: suomi, englanti, ranska, saksa ja hollanti.

### Agentic AI – Mallien orkestrointi tuotannossa

Tämän projektin mielenkiintoisin vaihe oli se, miten se rakennettiin. Hyödynsin **Agentic AI** -koodausta ja siinä hyödynsin [kilo.ai](https://kilo.ai) -alustaa, joka on pääasiallinen agentic AI -työkaluni VS Codessa. Tekoälyagentti toimii kuin ohjelmistoinsinööri, joka toteuttaa tiedostomuutokset ja suunnitelmat ohjeideni mukaisesti.

Käytin prosessissa useita eri kielimalleja niiden vahvuuksien mukaan:
1.  **Rakentaja (GPT Codex 5.3):** Hoiti 80 % "lihastyöstä" kääntämällä liiketoimintavaatimukset Next.js-komponenteiksi ja API-reiteiksi. Codex on hinta-laatu suhteeltaan tällä hetkellä selkeä suosikkini.
2.  **Arkkitehdit (Opus 4.6 & 4.7):** Senior-tason tarkastajat. Nämä mallit hoitivat kriittiset osat, kuten maksulogiikan turvallisuuden ja tietoturva-auditoinnin. Käyttäisin näitä jatkuvasti, mutta niiden tokenit ovat todella kalliita verrattuna muihin, joten optimoin niiden käyttöä vain kriittisiin tehtäviin.
3.  **Kieliasiantuntija (Gemini 3.1 Flash Lite Preview):** Varmisti, että lokalisointi viidelle kielelle ei kuulosta konekäännökseltä, vaan luonnolliselta. Tämä token-hinnoittelu on näistä selkeästi edullisin ja käännökset maksavat vain joitain kymmeniä senttejä, jos niitä on paljon.

### Lopputuloksena automatisoitu ostoprosessi ilman manuaalisia vaiheita

Muutos on ollut valtava. Se, mikä aiemmin vei pelikauden aikana tuntikausia viestittelyä ja manuaalista tarkistamista, on nyt vähentynyt lähes nollaan. 

Manuaalinen aikakausi on osaltani virallisesti ohi. 

---
*Oletko Levi Golfin osakkeenomistaja ja haluat automatisoida myyntisi? Tutustu sovellukseen osoitteessa [greenfee.levifinland.fi](https://greenfee.levifinland.fi) tai lue lisää koodauksen tulevaisuudesta osoitteessa [kilo.ai](https://kilo.ai).*