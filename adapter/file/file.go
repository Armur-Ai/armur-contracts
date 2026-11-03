// Package file is the file-backed implementation of contracts.Adapter: records
// are exchanged as JSON documents in a directory. It is the durable, offline
// transport (drop a bundle in a directory; another product picks it up) and
// behaves identically to the in-process and HTTP adapters — verified by the
// shared conformance suite. Records persist across process restarts.
package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	contracts "github.com/Armur-Ai/armur-contracts"
)

// Adapter stores each record at <dir>/<digest16>.json.
type Adapter struct {
	dir   string
	ns    contracts.Namespace
	kinds []string
}

// New creates (if needed) the backing directory and returns an adapter.
func New(dir string, ns contracts.Namespace, kinds ...string) (*Adapter, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Adapter{dir: dir, ns: ns, kinds: kinds}, nil
}

func (a *Adapter) pathFor(id string) string { return filepath.Join(a.dir, id+".json") }

// Submit writes the record, idempotently: identical bytes map to the same file
// and the same receipt.
func (a *Adapter) Submit(ctx context.Context, env contracts.Envelope) (contracts.Receipt, error) {
	if err := ctx.Err(); err != nil {
		return contracts.Receipt{}, err
	}
	if len(a.kinds) > 0 && !contains(a.kinds, env.Kind) {
		return contracts.Receipt{}, contracts.ErrUnsupportedKind
	}
	dig := contracts.Digest(env)
	id := dig[:16]
	path := a.pathFor(id)

	// Idempotent: if already stored, return the stored receipt.
	if b, err := os.ReadFile(path); err == nil {
		var stored contracts.Envelope
		if json.Unmarshal(b, &stored) == nil && stored.Receipt != nil {
			return *stored.Receipt, nil
		}
	}

	rec := contracts.Receipt{JobRef: contracts.JobRef("file:" + id), IssuedAt: time.Now().UTC(), Digest: dig}
	env.Receipt = &rec
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return contracts.Receipt{}, err
	}
	// Atomic write: temp + rename.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return contracts.Receipt{}, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return contracts.Receipt{}, err
	}
	return rec, nil
}

// Fetch reads a stored record by its ref, or ErrAbsentPeer if it isn't present.
func (a *Adapter) Fetch(ctx context.Context, ref contracts.JobRef) (contracts.Envelope, error) {
	if err := ctx.Err(); err != nil {
		return contracts.Envelope{}, err
	}
	id := string(ref)
	if len(id) > 5 && id[:5] == "file:" {
		id = id[5:]
	}
	b, err := os.ReadFile(a.pathFor(id))
	if os.IsNotExist(err) {
		return contracts.Envelope{}, contracts.ErrAbsentPeer
	}
	if err != nil {
		return contracts.Envelope{}, err
	}
	var env contracts.Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return contracts.Envelope{}, err
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
		Transports:    []string{"file"},
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

var _ contracts.Adapter = (*Adapter)(nil)
