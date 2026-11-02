package contracts

import "context"

// Adapter is the single exchange contract implemented identically by all three
// transports — local file/CLI, in-process Go, and optional authenticated HTTP.
// A record submitted or fetched over any of them carries the same semantics,
// verified by the shared conformance suite. Implementations MUST test
// cancellation, idempotency, capability/version negotiation and bounded errors.
type Adapter interface {
	// Submit delivers a record (as an Envelope) and returns a durable receipt.
	// It is idempotent on (producer, content digest): re-submitting the same
	// bytes returns the same receipt and does not duplicate the effect.
	Submit(ctx context.Context, env Envelope) (Receipt, error)

	// Fetch retrieves a previously exchanged record by its opaque ref. A missing
	// peer, a partial result, a stale baseline and a failed operation are
	// distinguishable via the returned error (see the Err* sentinels).
	Fetch(ctx context.Context, ref JobRef) (Envelope, error)

	// Negotiate exchanges capability + schema-version info before bulk transfer.
	Negotiate(ctx context.Context) (CapabilitySet, error)
}

// CapabilitySet is what a peer advertises during negotiation.
type CapabilitySet struct {
	SchemaVersion string    `json:"schema_version"`
	Namespace     Namespace `json:"namespace"`
	Kinds         []string  `json:"kinds"`      // record kinds this peer can accept/produce
	Transports    []string  `json:"transports"` // "file" | "inproc" | "http"
}

// TransportError distinguishes the failure modes a receiver must not conflate.
type TransportError string

func (e TransportError) Error() string { return string(e) }

const (
	// ErrAbsentPeer: the peer/record does not exist or is unreachable.
	ErrAbsentPeer TransportError = "contracts: absent peer"
	// ErrPartialResult: a result was returned but is explicitly incomplete.
	ErrPartialResult TransportError = "contracts: partial result"
	// ErrStaleBaseline: the referenced baseline revision no longer matches.
	ErrStaleBaseline TransportError = "contracts: stale baseline"
	// ErrFailedOperation: the operation ran and failed (not "found nothing").
	ErrFailedOperation TransportError = "contracts: failed operation"
	// ErrUnsupportedKind: the peer cannot handle this record kind/version.
	ErrUnsupportedKind TransportError = "contracts: unsupported kind"
)

// Note: the conformance suite (shared fixtures asserting identical behavior
// across the three transports) lives in a conformance subpackage, added with
// the first concrete adapter implementation.
