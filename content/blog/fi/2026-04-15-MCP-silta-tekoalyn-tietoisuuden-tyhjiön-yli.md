---
title: "MCP – Silta tekoälyn tietoisuuden tyhjiön yli"
description: "Onko MCP oikea ratkaisu vai kuollut syntyessään? Minulle se oli käänteentekevä ohjelmointiprojektien nopeuttamiseksi"
publishedAt: "2026-04-15"
slug: "mcp-bridgin-the-knowledge-gap"
draft: false
tags: ["AI", "MCP", "CLI", "Kilo-Code", "Context7", "Programming"]
---

![Context7 MCP -palvelin käytössä Kilo Codessa](/media/context7_mcp.png)

Minulle vibe-koodaus eli vibeltäminen oli loistava tapa aloittaa ohjelmointi tekoälyn kanssa, mutta törmäsin projekteissani nopeasti näkymättömään seinään. Se on tekoälyn rajallinen tietoisuus: kielimalli (LLM) on suljettu laatikko, jonka maailmankuva päättyy sen koulutusdatan katkaisupisteeseen, joka on kuukausia tai vuodenkin ennen varsinaista julkaisupäivää. Jos kysyt siltä jotain, mikä on tapahtunut viimeisen vuoden aikana, vastauksena on pelkkää arvausta tai hallusinaatiota. Ratkaisuna itselleni on ollut **MCP (Model Context Protocol)**. Se on silta, joka kytkee tekoälyn "aivot" reaalimaailman dataan.

### Mikä on MCP?

Yksinkertaistettuna MCP on standardoitu kommunikaatioprotokolla, joka toimii **universaalina adapterina** tekoälyagenttien ja ulkoisten tietolähteiden välillä. Jos LLM on tietokoneen prosessori, MCP on sen USB-portti. Se mahdollistaa sen, että agentti ei vain "puhu", vaan se voi operoida tietokantoja, kutsua API-rajapintoja tai käyttää erikoistuneita työkaluja, jotka eivät kuulu sen ydinosaamiseen.

### MCP osana Kilo Codea

Käytän työssäni [Kilo Codea](https://kilo.ai), joka käyttää MCP-protokollan natiivisti. Minun ei tarvitse edes kirjoittaa sitä lyhyttä koodinpätkää, vaan voin noutaa yleisimmät MCP:t suoraan ohjauspaneelista. Lisään vain tokenin ja se on valmis käyttöön.

**Konfigurointi on tehty helpoksi:**
Kilo Codessa MCP-palvelimet määritellään `kilo.jsonc`-tiedostossa. Voit asettaa ne joko globaalisti (`~/.config/kilo/kilo.jsonc`) tai projektikohtaisesti, jolloin projektitason asetukset ylikirjoittavat globaalit.

### Vuoden tyhjiön täyttäminen: Context7

Suurin ongelmani agenttien kanssa oli pitkään "koulutusdatan tyhjiö". Tekoäly ei tiennyt uusimmista kirjastopäivityksistä tai framework-muutoksista, jotka olivat tapahtuneet koulutusvaiheen jälkeen. Tämä oli erityisesti siinä ensimmäisessä green fee -myyntialustaprojektissani eniten aikaa vievä asia, sillä monet Geminin ohjeet tai sen olettamat riippuvuudet olivat vanhentuneita. Copy-paste aiheutti erroreita toisen perään ja sitten ongelmaa debugattiin joskus pitkäänkin. Kävin etsimässä viimeisimpiä dokumentaatioita netistä ja liitin niitä keskusteluun.

Ratkaisin tämän kytkemällä **Context7**-palvelun globaalisti Kilo Codeeni. Se on siis yhteisön ylläpitämä ajantasainen koodikirjasto. Se tarjoaa agentille reaaliaikaista kontekstia ja dokumentaatiota suoraan verkosta. En tee nykyään yhtäkään projektia ilman sitä – se on ero toimivan ja "melkein toimivan" koodin välillä. Se täyttää sen kriittisen yhden vuoden aukon, joka muuten tekisi agentista epäluotettavan.

### Debatti: Onko MCP jo kuollut?

Vaikka MCP on useimmiten hyvin toimiva ja vasta reilun vuoden vanha protokolla, koodariyhteisössä on jo pitkään keskusteltu sen tulevaisuudesta. Jotkut kokevat MCP:n turhana "lisäkerroksena", joka tuo mukanaan turhaa kompleksisuutta ja konteksti-ikkunan täytettä.

**CLI vs. MCP**
Monet agenttikoodarit vannovat nykyään raa'an **CLI-pääsyn** (Command Line Interface) nimeen. CLI:n etuna on nopeus ja joustavuus: agentti voi tehdä mitä tahansa komentorivillä ilman, että sille tarvitsee pystyttää erillistä MCP-palvelinta ja määritellä protokollia.

MCP taas edustaa hallittua ja turvallista lähestymistapaa. Se standardoi sen, mitä agentti voi nähdä ja tehdä. MCP:n vahvuus on ekosysteemissä – kun joku rakentaa hyvän MCP-palvelimen, se on heti kaikkien standardia tukevien työkalujen käytössä.

Kriitikot kuitenkin kysyvät: Miksi vaivautua opettelemaan uusi protokolla, jos voimme vain antaa agentille pääsyn tietoihin CLI:n kautta. Esimerkiksi OpenClawn kehittäjä Peter Steinberger vannoo CLI:n nimeen. Lex Fridmanin podcastissa Peter käy tätäkin aihetta läpi.

Tässäpä yksi sadoista artikkeleista, joita aiheesta on kirjoitettu: [MCP is dead, or MCP vs Skills revisited](https://medium.com/@alonisser/mcp-is-dead-or-mcp-vs-skills-revisited-daaa51b9a519).

### Yhteenveto

Olipa MCP tulevaisuuden standardi tai vain välivaihe matkalla kohti autonomisempia CLI-agentteja, se on tällä hetkellä minulle tehokkain tapa kuroa umpeen tekoälyn tietoisuuden aukko. Se on nopeuttanut omia projektejani ja antanut varmuuden, että koodi toimii kerralla, eikä versioyllätyksiä tule siinä vaiheessa, kun teen ensimmäistä buildia.
