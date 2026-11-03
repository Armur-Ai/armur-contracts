package file

import (
	"context"
	"testing"

	contracts "github.com/Armur-Ai/armur-contracts"
	"github.com/Armur-Ai/armur-contracts/conformance"
)

func TestFileConformance(t *testing.T) {
	conformance.RunAdapterSuite(t, func() contracts.Adapter {
		a, err := New(t.TempDir(), contracts.NSHayagriva, "DiscoveryBundle")
		if err != nil {
			t.Fatal(err)
		}
		return a
	})
}

func TestPersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	env, _ := contracts.NewEnvelope(contracts.NSHayagriva, contracts.Producer{Product: "hayagriva"},
		"DiscoveryBundle", contracts.DiscoveryBundle{BundleID: "b-1"})

	a1, _ := New(dir, contracts.NSHayagriva, "DiscoveryBundle")
	rec, err := a1.Submit(ctx, env)
	if err != nil {
		t.Fatal(err)
	}

	// A brand-new adapter over the same directory can read it back.
	a2, _ := New(dir, contracts.NSHayagriva, "DiscoveryBundle")
	got, err := a2.Fetch(ctx, rec.JobRef)
	if err != nil {
		t.Fatalf("fetch from fresh instance: %v", err)
	}
	if got.Kind != "DiscoveryBundle" {
		t.Fatalf("kind = %q", got.Kind)
	}
}
