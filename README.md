# armur-contracts

Shared, language-neutral **evidence vocabulary** exchanged between the six Armur
products (Pentest Swarm, Akira, Hayagriva, Engira, Trishul, Dhantra).

> **Status: v0, local.** These are the Go bindings of the canonical schema
> (mirrored in TypeScript for Akira's `@akira/schema`). One shared vocabulary —
> products must not fork a competing `Finding` type. See the suite migration
> plan for where this module will live permanently.

## What's here (v0)
- **Envelope + refs** (`envelope.go`): versioned `Envelope`, `Producer`, `Receipt`, `JobRef`, `EvidenceRef`, `SourceRevisionRef`, trust/claim status enums.
- **Discovery** (`discovery.go`): `DiscoveryBundle`, `ComponentObservation` — Hayagriva → Akira.
- **Code** (`code.go`): `CodeAnalysisBundle`, `SourceClaim`, `SecretCandidate`, `DependencyInventory`, `ReachabilityClaim`, `CodeIssue`, `QualityMetric` — Trishul → Akira/Engira.
- **Intelligence** (`intel.go`): `IntelligenceResult`, `AdvisoryRef`, `StandardMapping`, `Score` — Engira.
- **Assessment/graph** (`assessment.go`): `GraphNode`, `GraphEdge`, `AttackPath`, `ActionSpec`, `ChainTemplate`, `AssessmentSnapshot` — mirrors Akira.
- **Remediation** (`remediation.go`): `RemediationRequest`, `RemediationPlan`, `ChangeArtifact`, `VerificationRun`, `DeploymentReceipt`, `ApprovalRecord` — Engira → Dhantra → Trishul/Swarm.
- **Capabilities** (`capabilities.go`): `ScopeGuard`, `RequestGate`, `SecretResolver`, `SessionLeaseProvider`, `EvidenceSink`, `Clock`, `ApprovedHTTPClient`, `ReviewedToolRunner`.
- **Transport** (`transport.go`): the `Adapter` contract implemented identically over file / in-process Go / authenticated HTTP, with explicit failure modes.

## Design rules
Additive versioning · trust/receipt semantics (declared vs adjudicated vs runtime-confirmed stay distinct; no bare `fixed=true`) · three equivalent transports · no shared writable DB.

## Transports

A record (`Envelope`) is exchanged through an `Adapter`. Three implementations
ship, and all pass the shared `conformance` suite — so a product can be embedded
in-process, hand off via a directory, or talk over the network with identical
semantics (idempotency on content `Digest`, `ErrAbsentPeer` on a missing ref,
capability negotiation):

| Adapter | Package | Use |
|---|---|---|
| in-process | `adapter/inproc` | embedding one product inside another; tests/fixtures |
| file | `adapter/file` | durable, offline hand-off (drop a JSON record in a directory) |
| HTTP | `adapter/http` | networked exchange (`Client` ↔ `Handler`) |

```go
store := inproc.New(contracts.NSHayagriva, "DiscoveryBundle")
rec, _ := store.Submit(ctx, env)      // idempotent on Digest(env)
got, _ := store.Fetch(ctx, rec.JobRef)
```

New adapters must pass `conformance.RunAdapterSuite` to be considered conformant.
