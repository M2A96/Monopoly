package dice_test

import (
	"testing"

	"github/M2A96/Monopoly.git/pkg/dice"
)

func TestProviderSameSeedSameRolls(t *testing.T) {
	a := dice.NewProvider(42)
	b := dice.NewProvider(42)

	for i := 0; i < 1000; i++ {
		a1, a2 := a.Roll()
		b1, b2 := b.Roll()
		if a1 != b1 || a2 != b2 {
			t.Fatalf("roll %d differs: (%d,%d) vs (%d,%d)", i, a1, a2, b1, b2)
		}
	}
}

func TestProviderRollsInRange(t *testing.T) {
	p := dice.NewProvider(7)
	for i := 0; i < 1000; i++ {
		d1, d2 := p.Roll()
		if d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6 {
			t.Fatalf("roll %d out of range: (%d,%d)", i, d1, d2)
		}
	}
}

func TestProviderDifferentSeedsDiverge(t *testing.T) {
	a := dice.NewProvider(1)
	b := dice.NewProvider(2)
	for i := 0; i < 100; i++ {
		a1, a2 := a.Roll()
		b1, b2 := b.Roll()
		if a1 != b1 || a2 != b2 {
			return
		}
	}
	t.Fatal("different seeds produced 100 identical rolls")
}
