// Command viphash generates VIP portal credentials offline, outside the
// repository (plan §6.2, Phase 7). It prints an Argon2id hash of the access
// code plus a fresh independent cookie signing secret, ready to paste into
// /etc/goth/goth.env as VIP_PASSWORD_HASH and VIP_COOKIE_SECRET.
//
// Prefer piping the code so it never appears in the process list:
//
//	printf '%s' "$ACCESS_CODE" | go run ./cmd/viphash
//
// or run `go run ./cmd/viphash` and type the code at the prompt (stdin).
// The access code itself is never written anywhere by this tool.
package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Generation parameters (RFC 9106 second recommendation). They sit inside the
// verifier's bounds in internal/security/vip_password.go.
const (
	argon2MemoryKiB  = 65536 // 64 MiB
	argon2Iterations = 3
	argon2Threads    = 4
	argon2KeyLen     = 32
	saltLen          = 16
	cookieSecretLen  = 32
)

func main() {
	codeFlag := flag.String("code", "", "access code to hash (prefer stdin: the flag is visible in the process list)")
	flag.Parse()

	code := strings.TrimSpace(*codeFlag)
	if code == "" {
		fmt.Fprint(os.Stderr, "Access code (input is not echoed if stdin is a terminal): ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			fatal("reading access code: %v", err)
		}
		code = strings.TrimSpace(line)
	}
	if code == "" {
		fatal("empty access code")
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		fatal("generating salt: %v", err)
	}
	key := argon2.Key([]byte(code), salt, argon2Iterations, argon2MemoryKiB, argon2Threads, argon2KeyLen)

	secret := make([]byte, cookieSecretLen)
	if _, err := rand.Read(secret); err != nil {
		fatal("generating cookie secret: %v", err)
	}

	fmt.Printf("VIP_PASSWORD_HASH=$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s\n",
		argon2MemoryKiB, argon2Iterations, argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
	fmt.Printf("VIP_COOKIE_SECRET=%s\n", hex.EncodeToString(secret))

	fmt.Fprintln(os.Stderr, "Paste both values into /etc/goth/goth.env. Rotate both (regenerate) before any re-enable after suspected code sharing.")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "viphash: "+format+"\n", args...)
	os.Exit(1)
}
