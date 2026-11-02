// Package conformance provides a reusable test suite that every contracts.Adapter
// implementation (in-process, file/CLI, HTTP) must pass, so the three transports
// carry identical semantics. Call RunAdapterSuite from a _test.go in each
// adapter package.
package conformance

import (
	"context"
	"errors"
	"testing"

	contracts "github.com/Armur-Ai/armur-contracts"
)

// RunAdapterSuite exercises the Adapter contract: round-trip, idempotency,
// absent-ref error, and capability negotiation. newAdapter must return a fresh,
// empty adapter that accepts the "DiscoveryBundle" kind.
func RunAdapterSuite(t *testing.T, newAdapter func() contracts.Adapter) {
	t.Helper()
	ctx := context.Background()

	t.Run("round-trip", func(t *testing.T) {
		a := newAdapter()
		env := sampleEnvelope()
		rec, err := a.Submit(ctx, env)
		if err != nil {
			t.Fatalf("submit: %v", err)
		}
		if rec.JobRef == "" || rec.Digest == "" {
			t.Fatalf("receipt incomplete: %+v", rec)
		}
		got, err := a.Fetch(ctx, rec.JobRef)
		if err != nil {
			t.Fatalf("fetch: %v", err)
		}
		if got.Kind != env.Kind {
			t.Fatalf("kind = %q, want %q", got.Kind, env.Kind)
		}
	})

	t.Run("idempotent-on-identical-bytes", func(t *testing.T) {
		a := newAdapter()
		env := sampleEnvelope()
		r1, err := a.Submit(ctx, env)
		if err != nil {
			t.Fatal(err)
		}
		r2, err := a.Submit(ctx, env)
		if err != nil {
			t.Fatal(err)
		}
		if r1.JobRef != r2.JobRef || r1.Digest != r2.Digest {
			t.Fatalf("re-submitting identical bytes must be idempotent: %v vs %v", r1, r2)
		}
	})

	t.Run("absent-ref", func(t *testing.T) {
		a := newAdapter()
		_, err := a.Fetch(ctx, contracts.JobRef("does-not-exist"))
		if !errors.Is(err, contracts.ErrAbsentPeer) {
			t.Fatalf("want ErrAbsentPeer, got %v", err)
		}
	})

	t.Run("negotiate", func(t *testing.T) {
		a := newAdapter()
		cs, err := a.Negotiate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if cs.SchemaVersion != contracts.SchemaVersion {
			t.Fatalf("schema version = %q", cs.SchemaVersion)
		}
	})
}

func sampleEnvelope() contracts.Envelope {
	env, _ := contracts.NewEnvelope(
		contracts.NSHayagriva,
		contracts.Producer{Product: "hayagriva", Version: "test"},
		"DiscoveryBundle",
		contracts.DiscoveryBundle{BundleID: "b-1", Scope: []string{"example.test"}},
	)
	return env
}
