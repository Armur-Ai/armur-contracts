package contracts

import "time"

// RemediationRequest (Engira → Dhantra): an authorized request to repair, with
// explicit scope and allowed change classes. A CVSS score is NOT an
// authorization to modify anything; missing repo mapping or policy blocks the
// affected change, not unrelated read-only analysis.
type RemediationRequest struct {
	RequestID             string            `json:"request_id"`
	CaseRef               string            `json:"case_ref,omitempty"`
	FindingRefs           []string          `json:"finding_refs"` // immutable refs or attributed imported claims
	Source                SourceRevisionRef `json:"source"`
	AllowedChanges        []string          `json:"allowed_change_classes"` // e.g. dependency_bump|code_fix|config
	Exclusions            []string          `json:"exclusions,omitempty"`
	Requirements          string            `json:"requirements,omitempty"` // security + legitimate-behavior requirements + authority
	Policy                string            `json:"approval_policy,omitempty"`
	TargetEnv             string            `json:"target_environment,omitempty"`
	RequestedVerification []string          `json:"requested_verification,omitempty"` // rescan|retest
}

// RemediationPlan (Dhantra): the intended repair, distinguishing root-cause
// repair vs compensating control vs temporary mitigation.
type RemediationPlan struct {
	PlanID         string   `json:"plan_id"`
	Digest         string   `json:"digest"` // immutable plan digest
	RequestRef     string   `json:"request_ref"`
	Claim          string   `json:"claim"`    // precise claim being repaired
	Strategy       string   `json:"strategy"` // root_cause|compensating_control|mitigation
	BaseRevisions  []string `json:"base_revisions"`
	PermittedFiles []string `json:"permitted_files,omitempty"`
	RequiredTests  []string `json:"required_tests,omitempty"`
	Assumptions    []string `json:"assumptions,omitempty"`
	MissingInputs  []string `json:"missing_inputs,omitempty"`
	ApprovalGates  []string `json:"approval_gates,omitempty"`
}

// ChangeArtifact (Dhantra): the actual produced change. Never forge a human
// signature or claim the founder authored an automated change.
type ChangeArtifact struct {
	ChangeID         string   `json:"change_id"`
	BaseRevision     string   `json:"base_revision"`
	PatchDigest      string   `json:"patch_digest"`
	ChangedPaths     []string `json:"changed_paths"`
	LockfileChanges  []string `json:"lockfile_changes,omitempty"`
	ProposedCommits  []string `json:"proposed_commits,omitempty"`
	EvidenceRefs     []string `json:"source_evidence_refs,omitempty"`
	AgentAttribution string   `json:"agent_attribution"` // the automated agent; never a human signature
	ReviewState      string   `json:"review_state,omitempty"`
	TargetRepo       string   `json:"target_repo"`
}

// VerificationRun (a.k.a. RepairResult): keeps source-validation, build-success,
// deployment-observation and runtime-verification as SEPARATE fields. There is
// deliberately no single unqualified fixed=true.
type VerificationRun struct {
	RunID              string        `json:"run_id"`
	PlanRef            string        `json:"plan_ref,omitempty"`
	ChangeRef          string        `json:"change_ref,omitempty"`
	SourceValidated    string        `json:"source_validated"` // pass|fail|inconclusive|skipped
	BuildSucceeded     string        `json:"build_succeeded"`
	DeploymentObserved string        `json:"deployment_observed"`
	RuntimeVerified    string        `json:"runtime_verified"`
	Classification     string        `json:"classification,omitempty"` // repair|mitigation
	Baseline           string        `json:"baseline_ref,omitempty"`
	Candidate          string        `json:"candidate_ref,omitempty"`
	Evidence           []EvidenceRef `json:"evidence,omitempty"`
	Limits             []string      `json:"limits,omitempty"`
}

// DeploymentReceipt: a PR merge or CI status cannot manufacture a receipt;
// production deployment is never inferred from a branch name.
type DeploymentReceipt struct {
	ReceiptID      string        `json:"receipt_id"`
	ArtifactDigest string        `json:"artifact_digest"`
	TargetEnv      string        `json:"target_environment"`
	Authority      string        `json:"authority"`
	DeployedAt     time.Time     `json:"deployed_at"`
	State          string        `json:"state"` // observed|partial|unknown
	Health         []EvidenceRef `json:"health_evidence,omitempty"`
}

// ApprovalRecord binds an authenticated actor to an exact, immutable payload
// digest. A change to the payload, destination or effects invalidates approval.
type ApprovalRecord struct {
	Actor         string    `json:"actor"`
	PayloadDigest string    `json:"payload_digest"`
	Action        string    `json:"action"`
	Target        string    `json:"target"`
	Effects       []string  `json:"permitted_effects,omitempty"`
	PolicyVersion string    `json:"policy_version,omitempty"`
	Expiry        time.Time `json:"expiry,omitempty"`
}
