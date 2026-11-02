package inproc

import (
	"context"
	"testing"

	contracts "github.com/Armur-Ai/armur-contracts"
	"github.com/Armur-Ai/armur-contracts/conformance"
)

func TestInProcConformance(t *testing.T) {
	conformance.RunAdapterSuite(t, func() contracts.Adapter {
		return New(contracts.NSHayagriva, "DiscoveryBundle")
	})
}

func TestUnsupportedKindRejected(t *testing.T) {
	a := New(contracts.NSHayagriva, "DiscoveryBundle")
	env, _ := contracts.NewEnvelope(contracts.NSHayagriva, contracts.Producer{Product: "x"}, "CodeAnalysisBundle", struct{}{})
	if _, err := a.Submit(context.Background(), env); err != contracts.ErrUnsupportedKind {
		t.Fatalf("want ErrUnsupportedKind, got %v", err)
	}
}
