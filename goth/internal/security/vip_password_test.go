package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/scrypt"
)

// testArgon2Hash builds a PHC-style argon2id hash with test-small parameters
// (the verifier accepts any parameters within bounds; production hashes use
// the stronger cmd/viphash defaults).
func testArgon2Hash(t *testing.T, password string, memory, iterations, threads int) string {
	t.Helper()
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}
	key := argon2.Key([]byte(password), salt, uint32(iterations), uint32(memory), uint8(threads), 32)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, iterations, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

func testScryptHash(t *testing.T, password string, logN, r, p int) string {
	t.Helper()
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}
	key, err := scrypt.Key([]byte(password), salt, 1<<logN, r, p, 32)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("$scrypt$ln=%d,r=%d,p=%d$%s$%s",
		logN, r, p,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

func TestVerifyVIPPasswordArgon2id(t *testing.T) {
	hash := testArgon2Hash(t, "correct-horse-battery", 1024, 1, 2)
	if !VerifyVIPPassword(hash, "correct-horse-battery") {
		t.Error("correct password rejected")
	}
	if VerifyVIPPassword(hash, "wrong-password") {
		t.Error("wrong password accepted")
	}
	if VerifyVIPPassword(hash, "") {
		t.Error("empty password accepted")
	}
	if VerifyVIPPassword("", "correct-horse-battery") {
		t.Error("empty hash accepted")
	}
}

func TestVerifyVIPPasswordScrypt(t *testing.T) {
	hash := testScryptHash(t, "correct-horse-battery", 12, 8, 1)
	if !VerifyVIPPassword(hash, "correct-horse-battery") {
		t.Error("correct password rejected")
	}
	if VerifyVIPPassword(hash, "wrong-password") {
		t.Error("wrong password accepted")
	}
}

func TestVerifyVIPPasswordMalformedHashes(t *testing.T) {
	cases := []string{
		"plaintext-password",
		"$bcrypt$v=19$m=1024,t=1,p=2$c2FsdA$aGFzaA",
		"$argon2id$v=16$m=1024,t=1,p=2$c2FsdA$aGFzaA",                           // wrong version
		"$argon2id$v=19$m=1024,t=1$c2FsdA$aGFzaA",                               // missing p
		"$argon2id$v=19$m=1024,t=1,p=2$c2FsdA",                                  // missing hash
		"$argon2id$v=19$m=1024,t=1,p=2$!!!not-base64!!!$aGFzaA",                 // bad salt b64
		"$argon2id$v=19$m=1024,t=1,p=2$c2FsdA$!!!not-base64!!!",                 // bad hash b64
		"$argon2id$v=19$m=1024,t=1,p=2$c2FsdA$c2hvcnQ",                          // hash too short
		"$argon2id$v=19$m=999999999,t=1,p=2$c2FsdA$aGFzaGFzaGFzaGFzaGFzaGFzaGE", // memory out of bounds
		"$argon2id$v=19$m=1024,t=9999,p=2$c2FsdA$aGFzaGFzaGFzaGFzaGFzaGFzaGE",   // iterations out of bounds
		"$argon2id$v=19$m=1024,t=1,p=99$c2FsdA$aGFzaGFzaGFzaGFzaGFzaGFzaGE",     // threads out of bounds
		"$argon2id$v=19$m=-1,t=1,p=2$c2FsdA$aGFzaA",                             // negative param
		"$argon2id$v=19$m=abc,t=1,p=2$c2FsdA$aGFzaA",                            // non-numeric param
		"$scrypt$ln=99,r=8,p=1$c2FsdA$aGFzaGFzaGFzaGFzaGFzaGFzaGE",              // scrypt logN out of bounds
		"$scrypt$ln=12,r=8,p=1$c2FsdA",                                          // scrypt missing hash
	}
	for _, hash := range cases {
		if VerifyVIPPassword(hash, "any-password") {
			t.Errorf("malformed hash accepted: %q", hash)
		}
	}
}
