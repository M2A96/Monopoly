package dice

import (
	crand "crypto/rand"
	"math/rand/v2"
)

type Provider interface {
	Roll() (int, int)
}

type seededProvider struct {
	rng *rand.Rand
}

// NewProvider returns a deterministic provider for tests and simulations.
// Production rooms must use NewCryptoProvider: seeded rolls are predictable.
func NewProvider(seed int64) Provider {
	return &seededProvider{rng: rand.New(rand.NewPCG(uint64(seed), uint64(seed+1)))}
}

func (p *seededProvider) Roll() (int, int) {
	return p.rng.IntN(6) + 1, p.rng.IntN(6) + 1
}

// NewCryptoProvider returns an unpredictable provider for live games.
func NewCryptoProvider() Provider {
	var seed [32]byte
	if _, err := crand.Read(seed[:]); err != nil {
		panic("dice: crypto/rand unavailable: " + err.Error())
	}
	return &seededProvider{rng: rand.New(rand.NewChaCha8(seed))}
}
