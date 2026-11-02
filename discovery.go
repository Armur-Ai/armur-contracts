package contracts

import "time"

// DiscoveryBundle is Hayagriva's output: provenance-preserving discovery
// evidence (hosts, services, endpoints, API specs, app inventory). Scanner- or
// probe-reported vulnerabilities are SourceClaims, never confirmed findings —
// discovery observes surface, it does not adjudicate weaknesses.
type DiscoveryBundle struct {
	BundleID   string                 `json:"bundle_id"`
	Scope      []string               `json:"scope"` // the authorized scope this run honored
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	Collectors []CollectorRun         `json:"collectors"` // which collectors ran, with capability/coverage
	Components []ComponentObservation `json:"components"`
	Coverage   string                 `json:"coverage,omitempty"` // human note on completeness
	Limits     []string               `json:"limits,omitempty"`   // missing binaries / skipped areas (a gap, not a clean scan)
}

// CollectorRun records one discovery collector's execution with honest
// capability/coverage. A missing optional binary is an explicit gap, not a
// successful empty result.
type CollectorRun struct {
	Name         string `json:"name"` // e.g. "httpx", "katana", "openapi"
	Version      string `json:"version,omitempty"`
	Ran          bool   `json:"ran"`
	Unavailable  string `json:"unavailable,omitempty"` // reason if it couldn't run
	ItemsEmitted int    `json:"items_emitted"`
}

// ComponentObservation is one observed asset/surface element with provenance.
type ComponentObservation struct {
	Ref        string            `json:"ref"`   // stable observation id
	Kind       string            `json:"kind"`  // host|service|endpoint|api|form|object_ref|technology
	Value      string            `json:"value"` // the host/url/endpoint/etc.
	Attributes map[string]string `json:"attributes,omitempty"`
	ObservedBy string            `json:"observed_by"` // collector name
	ObservedAt time.Time         `json:"observed_at"`
	Evidence   []EvidenceRef     `json:"evidence,omitempty"`
	Trust      TrustStatus       `json:"trust"`
}
