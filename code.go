package contracts

import "time"

// CodeAnalysisBundle is Trishul's output: an immutable record of one
// code-analysis run across the three pillars (SAST / secrets / dependency),
// plus reliability/maintainability as separately-typed issues. Missing
// analyzers or parsing failures do NOT produce a clean scan.
type CodeAnalysisBundle struct {
	RunID        string               `json:"run_id"`
	Source       SourceRevisionRef    `json:"source"`
	Pillars      []string             `json:"pillars"`     // enabled: sast|secrets|dependency
	Analyzers    []AnalyzerPin        `json:"analyzers"`   // analyzer/rule/config/model versions
	Incremental  bool                 `json:"incremental"` // a changed-file scan is not full assurance
	Coverage     AnalysisCoverage     `json:"coverage"`
	SourceClaims []SourceClaim        `json:"source_claims,omitempty"`
	Secrets      []SecretCandidate    `json:"secrets,omitempty"`
	Dependencies *DependencyInventory `json:"dependencies,omitempty"`
	Issues       []CodeIssue          `json:"issues,omitempty"` // reliability/maintainability/architecture
	Metrics      []QualityMetric      `json:"metrics,omitempty"`
	StartedAt    time.Time            `json:"started_at"`
	FinishedAt   time.Time            `json:"finished_at"`
}

// AnalyzerPin records a versioned analyzer/rule set for reproducibility.
type AnalyzerPin struct {
	Name         string `json:"name"`
	Version      string `json:"version,omitempty"`
	RulesVersion string `json:"rules_version,omitempty"`
	ConfigDigest string `json:"config_digest,omitempty"`
}

// AnalysisCoverage records what was analyzed/skipped/failed so absence is never
// mistaken for assurance.
type AnalysisCoverage struct {
	Languages []string `json:"languages,omitempty"`
	Analyzed  int      `json:"analyzed_files"`
	Skipped   int      `json:"skipped_files"`
	Failed    int      `json:"failed_files"`
	Notes     []string `json:"notes,omitempty"`
}

// ClaimKind enumerates the explicit kinds of source claim.
type ClaimKind string

const (
	ClaimSourceWeakness     ClaimKind = "source_weakness"
	ClaimSecretCandidate    ClaimKind = "secret_candidate"
	ClaimDependencyAdvisory ClaimKind = "dependency_advisory"
	ClaimOther              ClaimKind = "other"
)

// SourceClaim is an attributed native-tool claim about source code. It is NOT a
// confirmed finding until adjudicated (see ClaimStatus). A source claim can be
// supported without asserting runtime exploitability.
type SourceClaim struct {
	Fingerprint string         `json:"fingerprint"` // versioned, collision-aware; line shifts != new root cause
	Tool        string         `json:"tool"`
	RuleID      string         `json:"rule_id,omitempty"`
	NativeID    string         `json:"native_id,omitempty"`
	Kind        ClaimKind      `json:"kind"`
	Severity    string         `json:"severity,omitempty"`
	AnalyzedRev string         `json:"analyzed_rev"` // commit id analyzed
	Locations   []CodeLocation `json:"locations,omitempty"`
	ProposedCWE []string       `json:"proposed_cwe,omitempty"`
	Confidence  string         `json:"confidence,omitempty"` // method/limits, not a bare number
	Status      ClaimStatus    `json:"status"`               // defaults to declared
	Evidence    []EvidenceRef  `json:"evidence,omitempty"`
	Suppression *Suppression   `json:"suppression,omitempty"`
}

// CodeLocation is a source position.
type CodeLocation struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
}

// Suppression carries owner/reason/scope/expiry and stays visible in exports.
type Suppression struct {
	Owner  string    `json:"owner"`
	Reason string    `json:"reason"`
	Scope  string    `json:"scope,omitempty"`
	Expiry time.Time `json:"expiry,omitempty"`
}

// SecretCandidate is a detected potential credential. Shared messages carry a
// redacted snippet and a keyed comparison token — never the raw secret. Pattern
// match, exposure, verified validity and revocation are separate states.
type SecretCandidate struct {
	Detector     string        `json:"detector"`
	Version      string        `json:"version,omitempty"`
	Location     CodeLocation  `json:"location"`
	RedactedCtx  string        `json:"redacted_context,omitempty"`
	Class        string        `json:"class,omitempty"`         // synthetic|unknown when supported
	CompareToken string        `json:"compare_token,omitempty"` // keyed, restricted; for dedup only
	ScopeRange   string        `json:"scope_range,omitempty"`   // working tree | history range | artifact
	Status       ClaimStatus   `json:"status"`                  // never auto-usable as a session
	Evidence     []EvidenceRef `json:"evidence,omitempty"`
}

// DependencyInventory records what software is included, with completeness.
// Distinguish declared/resolved/built/installed/runtime-observed.
type DependencyInventory struct {
	Source        SourceRevisionRef `json:"source"`
	Parser        string            `json:"parser"`
	ParserVersion string            `json:"parser_version,omitempty"`
	SourceKind    string            `json:"source_kind"` // manifest|lockfile|sbom|installed
	Components    []Dependency      `json:"components"`
	Completeness  string            `json:"completeness,omitempty"`
}

// Dependency is one package with identity and edges.
type Dependency struct {
	PURL      string             `json:"purl,omitempty"`
	Ecosystem string             `json:"ecosystem"`
	Name      string             `json:"name"`
	Version   string             `json:"version,omitempty"` // resolved version or commit
	Direct    bool               `json:"direct"`
	Scope     string             `json:"scope,omitempty"` // runtime|dev|build
	Checksums map[string]string  `json:"checksums,omitempty"`
	State     string             `json:"state,omitempty"` // declared|resolved|built|installed|observed
	Reach     *ReachabilityClaim `json:"reach,omitempty"`
}

// ReachabilityClaim: import presence is not proof of executing vulnerable code.
// "Not observed reachable" is not a universal safe verdict.
type ReachabilityClaim struct {
	Method    string        `json:"method"`    // static|dynamic|unknown
	Reachable string        `json:"reachable"` // yes|no|unknown
	Evidence  []EvidenceRef `json:"evidence,omitempty"`
	Limits    []string      `json:"limits,omitempty"`
}

// IssueKind keeps quality issues correctly typed so they never acquire
// CWE/CVE/CVSS merely to fit a security report.
type IssueKind string

const (
	IssueSecurityWeakness   IssueKind = "security_weakness"
	IssueReliability        IssueKind = "reliability"
	IssueMaintainability    IssueKind = "maintainability"
	IssueArchitecturePolicy IssueKind = "architecture_policy"
	IssueInformational      IssueKind = "informational"
)

// CodeIssue is an attributed non-secret, non-dependency code issue. Only
// security_weakness issues may become canonical security claims (via Akira).
type CodeIssue struct {
	Kind       IssueKind      `json:"kind"`
	Rule       string         `json:"rule,omitempty"`
	Source     string         `json:"source_rev"`
	Locations  []CodeLocation `json:"locations,omitempty"`
	Impact     string         `json:"impact,omitempty"`
	Confidence string         `json:"confidence,omitempty"`
	Evidence   []EvidenceRef  `json:"evidence,omitempty"`
}

// QualityMetric is a measured metric (e.g. duplication ratio, complexity). A
// threshold is configured policy, not "percentage secure".
type QualityMetric struct {
	Name    string  `json:"name"`
	Version string  `json:"version,omitempty"`
	Value   float64 `json:"value"`
	Unit    string  `json:"unit,omitempty"`
	Scope   string  `json:"scope,omitempty"`
	Limits  string  `json:"limits,omitempty"`
}
