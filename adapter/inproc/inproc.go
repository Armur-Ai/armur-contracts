// Package inproc is the in-process implementation of contracts.Adapter: the
// reference transport used for embedding one product inside another's process
// and for tests/fixtures. The file/CLI and authenticated-HTTP adapters must
// behave identically (verified by the conformance suite).
package inproc

import (
	"context"
	"sync"
	"time"

	contracts "github.com/Armur-Ai/armur-contracts"
)

// Adapter is an in-memory Adapter. Submissions are idempotent on the content
// digest: re-submitting identical bytes returns the same receipt and stores one
// record. Zero value is not usable — use New.
type Adapter struct {
	ns    contracts.Namespace
	kinds []string

	mu      sync.Mutex
	byRef   map[contracts.JobRef]contracts.Envelope
	byDigin map[string]contracts.Receipt // digest -> receipt (idempotency)
}

// New builds an in-process adapter advertising the given namespace + kinds.
func New(ns contracts.Namespace, kinds ...string) *Adapter {
	return &Adapter{
		ns:      ns,
		kinds:   kinds,
		byRef:   map[contracts.JobRef]contracts.Envelope{},
		byDigin: map[string]contracts.Receipt{},
	}
}

// Submit stores the envelope and returns a durable receipt, idempotently.
func (a *Adapter) Submit(ctx context.Context, env contracts.Envelope) (contracts.Receipt, error) {
	if err := ctx.Err(); err != nil {
		return contracts.Receipt{}, err
	}
	if len(a.kinds) > 0 && !contains(a.kinds, env.Kind) {
		return contracts.Receipt{}, contracts.ErrUnsupportedKind
	}
	dig := contracts.Digest(env)
	a.mu.Lock()
	defer a.mu.Unlock()
	if r, ok := a.byDigin[dig]; ok {
		return r, nil // idempotent: same bytes -> same receipt, no duplicate
	}
	ref := contracts.JobRef("inproc:" + dig[:16])
	rec := contracts.Receipt{JobRef: ref, IssuedAt: time.Now().UTC(), Digest: dig}
	stored := env
	stored.Receipt = &rec
	a.byRef[ref] = stored
	a.byDigin[dig] = rec
	return rec, nil
}

// Fetch returns a previously submitted record, or ErrAbsentPeer if unknown.
func (a *Adapter) Fetch(ctx context.Context, ref contracts.JobRef) (contracts.Envelope, error) {
	if err := ctx.Err(); err != nil {
		return contracts.Envelope{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	env, ok := a.byRef[ref]
	if !ok {
		return contracts.Envelope{}, contracts.ErrAbsentPeer
	}
	return env, nil
}

// Negotiate advertises this adapter's capabilities.
func (a *Adapter) Negotiate(ctx context.Context) (contracts.CapabilitySet, error) {
	if err := ctx.Err(); err != nil {
		return contracts.CapabilitySet{}, err
	}
	return contracts.CapabilitySet{
		SchemaVersion: contracts.SchemaVersion,
		Namespace:     a.ns,
		Kinds:         a.kinds,
		Transports:    []string{"inproc"},
	}, nil
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// compile-time assertion that Adapter satisfies the contract.
var _ contracts.Adapter = (*Adapter)(nil)
