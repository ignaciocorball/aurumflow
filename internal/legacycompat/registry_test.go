package legacycompat

import "testing"

func TestCompatibleIsNotValidated(t *testing.T) {
	g := Of("GOLD")
	if g.Status != Compatible || g.Validated {
		t.Fatal(g)
	}
	if Of("BTC").Status != Unsupported {
		t.Fatal(Of("BTC"))
	}
	if Of("US100").Status != NeedsConfig {
		t.Fatal(Of("US100"))
	}
}
