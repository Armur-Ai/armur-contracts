package contracts

import (
	"encoding/json"
	"time"
)

// Namespace scopes identifiers and records to an owning product so that IDs
// from different producers never collide.
type Namespace string

const (
	NSPentestSwarm Namespace = "swarm"
	NSAkira        Namespace = "akira"
	NSHayagriva    Namespace = "hayagriva"
	NSEngira       Namespace = "engira"
	NSTrishul      Namespace = "trishul"
	NSDhantra      Namespace = "dhantra"
)

// Producer identifies who emitted a record (for attribution and provenance).
type Producer struct {
	Product    string `json:"product"`               // e.g. "trishul"
	Version    string `json:"version"`               // producer build/version
	InstanceID string `json:"instance_id,omitempty"` // opaque run/instance id
}

// Receipt is the durable outbox/inbox acknowledgement for an exchanged record.
// Delivery is at-least-once; receivers deduplicate on (JobRef, Digest).
type Receipt struct {
	JobRef   JobRef    `json:"job_ref"`
	IssuedAt time.Time `json:"issued_at"`
	Digest   string    `json:"digest"` // content digest of the delivered bytes
}

// JobRef is an opaque, purpose-limited reference to a unit of work or an
// exchanged record. It is NOT a filesystem path — a receiver cannot use it to
// read arbitrary sender/host paths.
type JobRef string

// EvidenceRef points at an evidence blob by id + digest. Blobs may be missing,
// expired or redacted; such a ref stays explicitly unavailable rather than
// silently empty.
type EvidenceRef struct {
	ID                string `json:"id"`
	MediaType         string `json:"media_type,omitempty"`
	Digest            string `json:"digest,omitempty"`
	Available         bool   `json:"available"`
	UnavailableReason string `json:"unavailable_reason,omitempty"` // missing|expired|redacted
}

// SourceRevisionRef is the immutable identity of a scanned/analyzed source
// revision. A branch or tag is a locator, not identity — identity is the commit
// id plus the tree/content digest.
type SourceRevisionRef struct {
	RepoIdentity        string   `json:"repo_identity"`    // stable repo id
	Origin              string   `json:"origin,omitempty"` // sanitized; never credential-bearing
	CommitID            string   `json:"commit_id"`
	Subdir              string   `json:"subdir,omitempty"`      // workspace/subdirectory
	TreeDigest          string   `json:"tree_digest,omitempty"` // content digest
	Dirty               bool     `json:"dirty"`                 // dirty-worktree
	DiffDigest          string   `json:"diff_digest,omitempty"` // captured diff digest when dirty
	SubmoduleProvenance []string `json:"submodule_provenance,omitempty"`
}

// TrustStatus describes how strongly a claim/binding is substantiated.
type TrustStatus string

const (
	TrustObserved TrustStatus = "observed"
	TrustAttested TrustStatus = "attested"
	TrustAsserted TrustStatus = "asserted"
	TrustInferred TrustStatus = "inferred"
	TrustUnknown  TrustStatus = "unknown"
)

// ClaimStatus keeps the three evidence tiers distinguishable. Nothing is ever a
// bare "vulnerable=true": a tool result is Declared, an adjudicated source
// weakness is AdjudicatedSource, and only a runtime demonstration is
// RuntimeConfirmed.
type ClaimStatus string

const (
	StatusDeclared          ClaimStatus = "declared"           // native tool result, unadjudicated
	StatusAdjudicatedSource ClaimStatus = "adjudicated_source" // source weakness supported by a validator
	StatusRuntimeConfirmed  ClaimStatus = "runtime_confirmed"  // demonstrated against a running target
	StatusRefuted           ClaimStatus = "refuted"
	StatusUnknown           ClaimStatus = "unknown"
)

// Envelope is the versioned wrapper every exchanged record travels in. The
// payload is an opaque JSON document whose shape is given by Kind.
type Envelope struct {
	Version      string          `json:"version"` // == SchemaVersion at emit time
	Namespace    Namespace       `json:"namespace"`
	Producer     Producer        `json:"producer"`
	Kind         string          `json:"kind"` // e.g. "DiscoveryBundle", "CodeAnalysisBundle"
	EmittedAt    time.Time       `json:"emitted_at"`
	EvidenceRefs []EvidenceRef   `json:"evidence_refs,omitempty"`
	Receipt      *Receipt        `json:"receipt,omitempty"`
	Payload      json.RawMessage `json:"payload"`
}

// NewEnvelope wraps a payload record of the given kind. The payload is
// JSON-encoded; callers typically pass one of the record types in this package.
func NewEnvelope(ns Namespace, p Producer, kind string, payload any) (Envelope, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		Version:   SchemaVersion,
		Namespace: ns,
		Producer:  p,
		Kind:      kind,
		EmittedAt: time.Now().UTC(),
		Payload:   json.RawMessage(b),
	}, nil
}

// Decode unmarshals the envelope payload into dst.
func (e Envelope) Decode(dst any) error { return json.Unmarshal(e.Payload, dst) }
