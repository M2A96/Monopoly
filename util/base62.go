// util/base62.go
package util

import (
	"math/big"

	"github.com/google/uuid"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// UUIDToBase62 converts a UUID to a base62 string representation
// This makes UUIDs shorter and more URL-friendly while maintaining uniqueness
func UUIDToBase62(id uuid.UUID) string {
	// Convert UUID to a big integer
	var i big.Int
	i.SetBytes(id[:])

	// Convert to base62
	var result string
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := &big.Int{}

	// Generate the base62 string by repeatedly dividing by 62 and using the remainder
	for i.Cmp(zero) > 0 {
		i.DivMod(&i, base, mod)
		result = string(base62Chars[mod.Int64()]) + result
	}

	// Handle the special case of UUID being all zeros
	if result == "" {
		return "0"
	}

	return result
}

// Base62ToUUID converts a base62 string back to a UUID
// This is the reverse operation of UUIDToBase62
func Base62ToUUID(s string) (uuid.UUID, error) {
	var i big.Int
	base := big.NewInt(62)

	// Convert from base62 to big integer
	for _, c := range s {
		i.Mul(&i, base)
		pos := int64(0)
		for j, char := range base62Chars {
			if c == char {
				pos = int64(j)
				break
			}
		}
		i.Add(&i, big.NewInt(pos))
	}

	// Convert big integer to UUID
	var id uuid.UUID
	bs := i.Bytes()
	// Ensure the byte slice is the correct length
	if len(bs) > 16 {
		bs = bs[len(bs)-16:]
	} else if len(bs) < 16 {
		padded := make([]byte, 16)
		copy(padded[16-len(bs):], bs)
		bs = padded
	}
	copy(id[:], bs)

	return id, nil
}
