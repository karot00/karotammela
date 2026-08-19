package security

import (
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/scrypt"
)

// Parameter bounds protect the verifier from pathological (e.g. accidentally
// mis-generated) hashes: the encoded parameters drive the work factor, so a
// huge memory or iteration count must be rejected, not executed.
const (
	vipArgon2MaxMemoryKiB  = 256 * 1024
	vipArgon2MaxIterations = 64
	vipArgon2MaxThreads    = 16
	vipScryptMaxLogN       = 20
	vipScryptMaxR          = 16
	vipScryptMaxP          = 16
	vipHashMaxKeyLen       = 64
	vipHashMinKeyLen       = 16
	vipHashMaxSegmentLen   = 256
)

// VerifyVIPPassword checks a candidate access code against an encoded hash in
// constant time (via the hash comparison; parameter parsing is not secret).
// Supported PHC-style encodings with raw unpadded standard-base64 segments:
//
//	$argon2id$v=19$m=<KiB>,t=<iterations>,p=<threads>$<salt>$<hash>
//	$scrypt$ln=<log2N>,r=<r>,p=<p>$<salt>$<hash>
//
// Any malformed encoding, out-of-bound parameter, or length mismatch returns
// false. The plaintext access code is never stored or compared directly
// (plan §6.2).
func VerifyVIPPassword(encoded, password string) bool {
	if encoded == "" || password == "" {
		return false
	}
	fields := strings.Split(encoded, "$")
	switch {
	case len(fields) == 6 && fields[1] == "argon2id":
		return verifyArgon2id(fields, password)
	case len(fields) == 5 && fields[1] == "scrypt":
		return verifyScrypt(fields, password)
	default:
		return false
	}
}

// ValidateVIPPasswordHash checks the encoded profile without performing the
// expensive derivation. Production callers use it before accepting traffic.
func ValidateVIPPasswordHash(encoded string, production bool) bool {
	fields := strings.Split(encoded, "$")
	if len(fields) == 6 && fields[1] == "argon2id" {
		if fields[0] != "" || fields[2] != "v=19" {
			return false
		}
		params := parseParams(fields[3])
		if params == nil || params["m"] > vipArgon2MaxMemoryKiB || params["t"] > vipArgon2MaxIterations || params["p"] > vipArgon2MaxThreads {
			return false
		}
		if production && (params["m"] < 64*1024 || params["t"] < 3 || params["p"] < 1) {
			return false
		}
		salt, key, ok := decodeVIPHashSegments(fields[4], fields[5])
		return ok && len(salt) >= 16 && len(key) >= 32
	}
	if len(fields) == 5 && fields[1] == "scrypt" {
		params := parseParams(fields[2])
		if params == nil || params["ln"] > vipScryptMaxLogN || params["r"] > vipScryptMaxR || params["p"] > vipScryptMaxP {
			return false
		}
		if production && (params["ln"] < 15 || params["r"] < 8 || params["p"] < 1) {
			return false
		}
		_, key, ok := decodeVIPHashSegments(fields[3], fields[4])
		return ok && len(key) >= 32
	}
	return false
}

func parseParams(raw string) map[string]int {
	params := map[string]int{}
	for _, kv := range strings.Split(raw, ",") {
		k, v, ok := strings.Cut(kv, "=")
		n, err := strconv.Atoi(v)
		if !ok || k == "" || err != nil || n < 1 {
			return nil
		}
		params[k] = n
	}
	return params
}

func verifyArgon2id(fields []string, password string) bool {
	// fields: ["", "argon2id", "v=19", "m=..,t=..,p=..", salt, hash]
	if fields[0] != "" || fields[2] != "v=19" {
		return false
	}
	params := map[string]int{}
	for _, kv := range strings.Split(fields[3], ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return false
		}
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return false
		}
		params[k] = n
	}
	memory, hasM := params["m"]
	iterations, hasT := params["t"]
	threads, hasP := params["p"]
	if !hasM || !hasT || !hasP {
		return false
	}
	if memory > vipArgon2MaxMemoryKiB || iterations > vipArgon2MaxIterations || threads > vipArgon2MaxThreads {
		return false
	}
	salt, key, ok := decodeVIPHashSegments(fields[4], fields[5])
	if !ok {
		return false
	}
	computed := argon2.Key([]byte(password), salt, uint32(iterations), uint32(memory), uint8(threads), uint32(len(key)))
	return subtle.ConstantTimeCompare(computed, key) == 1
}

func verifyScrypt(fields []string, password string) bool {
	// fields: ["", "scrypt", "ln=..,r=..,p=..", salt, hash]
	if fields[0] != "" {
		return false
	}
	params := map[string]int{}
	for _, kv := range strings.Split(fields[2], ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return false
		}
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return false
		}
		params[k] = n
	}
	logN, hasLn := params["ln"]
	r, hasR := params["r"]
	p, hasP := params["p"]
	if !hasLn || !hasR || !hasP {
		return false
	}
	if logN > vipScryptMaxLogN || r > vipScryptMaxR || p > vipScryptMaxP {
		return false
	}
	salt, key, ok := decodeVIPHashSegments(fields[3], fields[4])
	if !ok {
		return false
	}
	computed, err := scrypt.Key([]byte(password), salt, 1<<logN, r, p, len(key))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(computed, key) == 1
}

func decodeVIPHashSegments(saltB64, hashB64 string) (salt, key []byte, ok bool) {
	if len(saltB64) == 0 || len(saltB64) > vipHashMaxSegmentLen ||
		len(hashB64) == 0 || len(hashB64) > vipHashMaxSegmentLen {
		return nil, nil, false
	}
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil || len(salt) == 0 {
		return nil, nil, false
	}
	key, err = base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil || len(key) < vipHashMinKeyLen || len(key) > vipHashMaxKeyLen {
		return nil, nil, false
	}
	return salt, key, true
}
