package dice

import "math/rand/v2"

type Provider interface {
	Roll() (int, int)
}

type seededProvider struct {
	rng *rand.Rand
}

func NewProvider(seed int64) Provider {
	return &seededProvider{rng: rand.New(rand.NewPCG(uint64(seed), uint64(seed+1)))}
}

func (p *seededProvider) Roll() (int, int) {
	return p.rng.IntN(6) + 1, p.rng.IntN(6) + 1
}
