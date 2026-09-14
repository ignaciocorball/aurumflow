package strategy

import (
	"reflect"
	"strings"
	"testing"
)

func TestDecisionContextHasNoSalienceOrAttention(t *testing.T) {
	rt := reflect.TypeOf(DecisionContext{})
	for i := 0; i < rt.NumField(); i++ {
		n := strings.ToLower(rt.Field(i).Name)
		if strings.Contains(n, "salience") || strings.Contains(n, "attention") || strings.Contains(n, "worldstate") {
			t.Fatalf("execution context must not consume %s", rt.Field(i).Name)
		}
	}
}
