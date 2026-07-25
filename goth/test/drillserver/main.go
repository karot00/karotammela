// drillserver is a local fixture provider for the Phase 12.5g standalone and
// failure drills. It emulates the Algolia HN API, GitHub Trending pages, the
// Yahoo chart API, and the Gemini generateContent endpoint, with per-source
// failure switches so the drill can fail each provider independently.
//
// Run from the goth module root:
//
//	go run ./test/drillserver -addr 127.0.0.1:8099 [-fail hn,github,yahoo,gemini] \
//	  [-malformed-github] [-delay-yahoo 3s]
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var trendingFixture string

func loadFixture() string {
	if trendingFixture != "" {
		return trendingFixture
	}
	data, err := os.ReadFile("internal/aipulse/testdata/github-trending.html")
	if err != nil {
		log.Fatalf("drillserver: read trending fixture (run from goth/ root): %v", err)
	}
	trendingFixture = string(data)
	return trendingFixture
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8099", "listen address")
	failCSV := flag.String("fail", "", "comma-separated sources to fail: hn,github,yahoo,gemini")
	malformedGH := flag.Bool("malformed-github", false, "serve 200 with unparseable trending markup")
	delayYahoo := flag.Duration("delay-yahoo", 0, "artificial delay for every Yahoo response")
	flag.Parse()

	failed := map[string]bool{}
	for _, s := range strings.Split(*failCSV, ",") {
		if s != "" {
			failed[s] = true
		}
	}

	mux := http.NewServeMux()

	// Always-on readiness marker (unaffected by -fail switches).
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "drillserver-ok")
	})

	// HN Algolia API.
	mux.HandleFunc("/api/v1/search_by_date", func(w http.ResponseWriter, r *http.Request) {
		if failed["hn"] {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"hits":[
			{"title":"Drill story one","url":"https://drill.local/one","points":95,"objectID":"d1"},
			{"title":"Drill story two","url":"https://drill.local/two","points":90,"objectID":"d2"},
			{"title":"Drill story three","url":"https://drill.local/three","points":85,"objectID":"d3"},
			{"title":"Drill story four","url":"https://drill.local/four","points":80,"objectID":"d4"},
			{"title":"Drill story five","url":"https://drill.local/five","points":75,"objectID":"d5"}]}`)
	})

	// GitHub Trending pages.
	ghHandler := func(w http.ResponseWriter, r *http.Request) {
		if failed["github"] {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if *malformedGH {
			fmt.Fprint(w, `<html><body class="redesigned-2026">Nothing parseable here</body></html>`)
			return
		}
		fmt.Fprint(w, loadFixture())
	}
	mux.HandleFunc("/trending", ghHandler)
	mux.HandleFunc("/trending/python", ghHandler)
	mux.HandleFunc("/trending/jupyter-notebook", ghHandler)

	// Yahoo chart API.
	mux.HandleFunc("/v8/finance/chart/", func(w http.ResponseWriter, r *http.Request) {
		if failed["yahoo"] {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if *delayYahoo > 0 {
			time.Sleep(*delayYahoo)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"chart":{"result":[{"timestamp":[1800000000,1800086400],"indicators":{"quote":[{"open":[100.0,101.0],"high":[102.0,103.0],"low":[99.0,100.0],"close":[101.5,102.5],"volume":[1000,2000]}]}}],"error":null}}`)
	})

	// Gemini generateContent.
	mux.HandleFunc("/gemini", func(w http.ResponseWriter, r *http.Request) {
		if failed["gemini"] {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error":{"message":"quota exhausted (drill)"}}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"text":"Drill summary."}]}}]}`)
	})

	log.Printf("drillserver listening on %s (fail=%q malformed-github=%v delay-yahoo=%s)", *addr, *failCSV, *malformedGH, *delayYahoo)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
