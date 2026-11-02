// Package contracts defines the shared, language-neutral evidence vocabulary
// exchanged between the six Armur products (Pentest Swarm, Akira, Hayagriva,
// Engira, Trishul, Dhantra).
//
// These are the Go bindings of the canonical schema. The same records are
// mirrored in TypeScript for Akira (which extends @akira/schema). There is one
// shared vocabulary — products MUST NOT fork a competing generic Finding type
// (spec §5.1).
//
// Design rules (from the six-product spec):
//
//   - Additive versioning. Extend records with new optional fields; do not
//     repurpose or remove fields. SchemaVersion tracks the vocabulary.
//   - Trust/receipt semantics. A native scanner's output is an attributed
//     SourceClaim, never a confirmed finding, until an adjudicator supports a
//     canonical claim. Keep "declared tool result", "adjudicated source
//     weakness" and "runtime-confirmed violation" distinguishable — there is no
//     single unqualified fixed=true / vulnerable=true field anywhere.
//   - Three equivalent transports. A record exchanged over local file/CLI,
//     in-process Go, or authenticated HTTP carries identical semantics; see
//     Adapter and the conformance suite in transport.go.
//   - No shared writable database. Each product owns its own store; these types
//     describe exchange records only, moved via durable outbox/inbox receipts
//     with idempotency, never a distributed transaction.
//
// Status: v0 — the core records and capability interfaces. Remaining spec
// records (FlowTrace, SourceBuildDeploymentBinding, ChangeAssessment, etc.) are
// added additively.
package contracts

// SchemaVersion is the version of this shared vocabulary. Bump the minor for
// additive changes, the major for an incompatible one (avoid the latter).
const SchemaVersion = "0.1.0"
