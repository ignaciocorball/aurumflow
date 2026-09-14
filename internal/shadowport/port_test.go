package shadowport

import "testing"

func TestEqualRAndCollision(t *testing.T) {
	var p Portfolio
	p.Add(Trade{Market: "US100", Dir: 1, R: 1})
	p.Add(Trade{Market: "US500", Dir: 1, R: 1})
	if p.ReturnR() != 2 || p.Collisions() < 1 {
		t.Fatalf("%+v coll=%d", p, p.Collisions())
	}
	if p.MaxDDR() != 0 {
		t.Fatal(p.MaxDDR())
	}
}
