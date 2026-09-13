package binanceusdm

import "testing"

func TestSnapshotFirstEventRule(t *testing.T) {
	// lastUpdateId=10; event U=9 u=12 covers 11.
	if !FirstApplicable(10, 9, 12) {
		t.Fatal("expected first applicable")
	}
	if FirstApplicable(10, 12, 13) {
		t.Fatal("gap after snapshot")
	}
	if !Obsolete(10, 10) || Obsolete(10, 11) {
		t.Fatal("obsolete")
	}
}
