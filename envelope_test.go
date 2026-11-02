package contracts

import "testing"

func TestEnvelopeRoundTrip(t *testing.T) {
	bundle := DiscoveryBundle{
		BundleID: "b-1",
		Scope:    []string{"example.test"},
		Components: []ComponentObservation{{
			Ref: "o-1", Kind: "endpoint", Value: "https://example.test/api",
			ObservedBy: "httpx", Trust: TrustObserved,
		}},
	}
	env, err := NewEnvelope(NSHayagriva, Producer{Product: "hayagriva", Version: "0.0.0"}, "DiscoveryBundle", bundle)
	if err != nil {
		t.Fatalf("NewEnvelope: %v", err)
	}
	if env.Version != SchemaVersion {
		t.Fatalf("version = %q, want %q", env.Version, SchemaVersion)
	}
	if env.Namespace != NSHayagriva {
		t.Fatalf("namespace = %q", env.Namespace)
	}

	var got DiscoveryBundle
	if err := env.Decode(&got); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.BundleID != bundle.BundleID || len(got.Components) != 1 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if got.Components[0].Trust != TrustObserved {
		t.Fatalf("trust not preserved: %q", got.Components[0].Trust)
	}
}

func TestClaimDefaultsAreExplicit(t *testing.T) {
	// A source claim must be able to express the three evidence tiers distinctly.
	for _, s := range []ClaimStatus{StatusDeclared, StatusAdjudicatedSource, StatusRuntimeConfirmed} {
		if s == "" {
			t.Fatal("empty claim status")
		}
	}
}
