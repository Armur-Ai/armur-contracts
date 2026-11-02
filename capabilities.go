package contracts

import (
	"context"
	"net/http"
	"time"
)

// Capability interfaces (spec §6). These are REFERENCE definitions of the
// narrow capabilities a collector/runtime is admitted with. Per the spec,
// interfaces should be consumer-owned where that avoids import cycles — a
// product may re-declare the subset it consumes; the transport DTOs above stay
// canonical. Enforcement happens at the real HTTP/dial/tool boundary: an
// interface that hands back a raw, unrestricted client is NOT enforcement.

// ScopeGuard authorizes a target at the dial/HTTP boundary. Admission does not
// grant raw network access; every outbound action is checked here.
type ScopeGuard interface {
	// InScope reports whether target (host[:port] or URL) is within the
	// authorized scope, with a reason when it is not.
	InScope(ctx context.Context, target string) (ScopeDecision, error)
}

// ScopeDecision is the result of a scope check.
type ScopeDecision struct {
	Allowed bool
	Reason  string
}

// RequestGate meters and rate/budget-limits operations. Use operation IDs so
// one request is not double-counted at two layers.
type RequestGate interface {
	// Acquire blocks until the operation may proceed or ctx is cancelled; the
	// returned func releases/records the operation.
	Acquire(ctx context.Context, opID string) (release func(), err error)
}

// SecretResolver resolves an opaque secret reference to a value for immediate,
// scoped use. Raw secrets never travel in exchange records; they are resolved
// here, at the point of use only.
type SecretResolver interface {
	Resolve(ctx context.Context, ref string) (value string, err error)
}

// SessionLeaseProvider issues a bounded, cancellable session lease (e.g. an
// authenticated target session). A suspected credential is never an
// automatically usable session.
type SessionLeaseProvider interface {
	Lease(ctx context.Context, key string) (Lease, error)
}

// Lease is a held session lease; Release returns it.
type Lease interface {
	ID() string
	Release() error
}

// EvidenceSink receives evidence/records as versioned envelopes and returns a
// durable receipt. Delivery is at-least-once; the sink deduplicates.
type EvidenceSink interface {
	Put(ctx context.Context, env Envelope) (Receipt, error)
}

// Clock is an injectable time source (deterministic in tests/fixtures).
type Clock interface {
	Now() time.Time
}

// ApprovedHTTPClient performs only scope-approved HTTP, enforcing policy at the
// request boundary. It intentionally exposes only Do, not the raw transport.
type ApprovedHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ReviewedToolRunner executes only reviewed, operator-installed tools under an
// enforcing isolation profile. Repository config/CI/hooks cannot add tools,
// destinations, or permissions here.
type ReviewedToolRunner interface {
	Run(ctx context.Context, spec ToolSpec) (ToolResult, error)
}

// ToolSpec names a reviewed tool invocation. The runner resolves the actual
// binary/policy; a caller cannot inject an arbitrary command.
type ToolSpec struct {
	Tool string            // reviewed tool name (not a shell command)
	Args []string          // validated args
	Env  map[string]string // allowlisted env only
}

// ToolResult is a tool invocation outcome. A tool failure is not a negative
// (clean) result — callers must distinguish the two.
type ToolResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Failed   bool // the invocation itself failed (vs. ran and found nothing)
}
