// mintcookie prints a karot_unlock cookie value for drills and smoke tests.
// The secret comes from UNLOCK_COOKIE_SECRET (or -secret).
//
// Run from the goth module:  go run ./test/mintcookie [-secret s]
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"goth/internal/security"
)

func main() {
	secret := flag.String("secret", os.Getenv("UNLOCK_COOKIE_SECRET"), "HMAC secret")
	locale := flag.String("locale", "fi", "payload locale")
	flag.Parse()
	if *secret == "" {
		fmt.Fprintln(os.Stderr, "mintcookie: no secret (set UNLOCK_COOKIE_SECRET or -secret)")
		os.Exit(1)
	}
	fmt.Print(security.CreateUnlockCookieValue(security.UnlockPayload{
		SessionID:  "drill-session",
		Locale:     *locale,
		UnlockedAt: time.Now().UnixMilli(),
	}, *secret))
}
