package contracts

import "time"

// This file mirrors Akira's canonical assessment/graph vocabulary (its zod
// @akira/schema: Node/Edge/Technique/Scenario in; AttackPath/RunResult out) so
// Go producers/consumers speak the same language. Akira remains the authority
// for graph semantics and path reasoning; these are the exchange shapes.

// GraphNode mirrors Akira's Node (asset/identity/secret/capability...).
type GraphNode struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // human|agent|model|tool|data|service|network|credential|secret|...
	Label       string            `json:"label"`
	Criticality float64           `json:"criticality,omitempty"` // 0..1
	Meta        map[string]string `json:"meta,omitempty"`
}

// GraphEdge mirrors Akira's Edge with technique annotations + exploitability.
type GraphEdge struct {
	ID             string   `json:"id"`
	From           string   `json:"from"`
	To             string   `json:"to"`
	Kind           string   `json:"kind"` // trust|data-flow|tool-grant|network-reach|identity-assumes|...
	TechniqueIDs   []string `json:"technique_ids,omitempty"`
	Exploitability float64  `json:"exploitability,omitempty"` // 0..1
	Verified       bool     `json:"verified"`                 // proven (not just modeled)
	Requires       []string `json:"requires,omitempty"`
}

// AttackPath is a reasoned route entry→objective. It is Akira's output, not a
// shortest-path over arbitrary relationships. "No path found within limits" is
// not assurance of security.
type AttackPath struct {
	Objective   string     `json:"objective"`
	Steps       []PathStep `json:"steps"`
	Probability float64    `json:"probability,omitempty"`
	Complete    bool       `json:"complete"`           // false if search hit its bounds
	Unknowns    []string   `json:"unknowns,omitempty"` // missing/conflicting preconditions
}

// PathStep is one hop with its technique + proof linkage.
type PathStep struct {
	EdgeID      string `json:"edge_id"`
	TechniqueID string `json:"technique_id,omitempty"`
	ClaimRef    string `json:"claim_ref,omitempty"` // supporting claim/finding, if any
	Verified    bool   `json:"verified"`
}

// ActionSpec is a declarative action in Akira's catalog (preconditions →
// effects). Path search composes ActionSpecs; it is not template replay.
type ActionSpec struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Requires    []string `json:"requires,omitempty"` // typed preconditions
	Grants      []string `json:"grants,omitempty"`   // postconditions
	TechniqueID string   `json:"technique_id,omitempty"`
	ReviewState string   `json:"review_state,omitempty"`
}

// ChainTemplate is a named, versioned declarative chain of ActionSpecs.
type ChainTemplate struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CVEs      []string `json:"cves,omitempty"`
	Product   string   `json:"product,omitempty"`
	ActionIDs []string `json:"action_ids"`
	Version   string   `json:"version,omitempty"`
}

// AssessmentSnapshot is an immutable, versioned view Akira exports (and Engira
// imports immutably) tying evidence to graph/paths at a point in time.
type AssessmentSnapshot struct {
	SnapshotID string       `json:"snapshot_id"`
	Revision   string       `json:"revision"`
	Nodes      []GraphNode  `json:"nodes"`
	Edges      []GraphEdge  `json:"edges"`
	Paths      []AttackPath `json:"paths,omitempty"`
	TakenAt    time.Time    `json:"taken_at"`
}
