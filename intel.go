package contracts

import "time"

// IntelligenceResult is Engira's output: advisory applicability + scoring +
// standards mapping bound to a specific dependency/component snapshot. Engira
// owns the single authoritative advisory matcher; other products delegate to it
// rather than reimplementing matching.
type IntelligenceResult struct {
	ResultID     string            `json:"result_id"`
	SnapshotRef  string            `json:"snapshot_ref"`            // the intelligence snapshot/version used
	InventoryRef string            `json:"inventory_ref,omitempty"` // the DependencyInventory it references
	Advisories   []AdvisoryRef     `json:"advisories,omitempty"`
	Mappings     []StandardMapping `json:"mappings,omitempty"`
	Scores       []Score           `json:"scores,omitempty"`
	GeneratedAt  time.Time         `json:"generated_at"`
}

// AdvisoryRef is a matched advisory with explicit applicability + rationale.
// A matched advisory is not an exploited finding.
type AdvisoryRef struct {
	AdvisoryID    string      `json:"advisory_id"`   // e.g. CVE / GHSA id
	Component     string      `json:"component"`     // purl or name@version
	Applicability string      `json:"applicability"` // affected|not_affected|unknown
	Rationale     string      `json:"rationale,omitempty"`
	Source        string      `json:"source,omitempty"` // dataset + version
	Trust         TrustStatus `json:"trust"`
}

// Score is a deterministic scoring result preserving its system + version.
type Score struct {
	System  string  `json:"system"`  // e.g. "CVSS"
	Version string  `json:"version"` // e.g. "3.1"
	Vector  string  `json:"vector,omitempty"`
	Value   float64 `json:"value"`
	// A score is never an authorization to modify anything (see RemediationRequest).
}

// StandardMapping attaches a subject to a reference system (CWE/OWASP/ATT&CK)
// with rationale + review status. Multiple or no justified mappings are allowed.
type StandardMapping struct {
	Subject     string      `json:"subject"` // ref to the claim/issue being mapped
	System      string      `json:"system"`  // CWE|OWASP|ATTACK|ATLAS
	Version     string      `json:"version,omitempty"`
	Identifier  string      `json:"identifier"` // e.g. "CWE-89"
	Rationale   string      `json:"rationale,omitempty"`
	ReviewState string      `json:"review_state,omitempty"` // proposed|reviewed
	Trust       TrustStatus `json:"trust"`
}
