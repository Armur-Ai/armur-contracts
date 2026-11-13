package contracts

import (
	"fmt"
	"strings"
)

// This file adds referential-integrity validation for the exchange records that
// carry a graph. CheckEnvelope guards *schema* compatibility (can I read this
// major?); ValidateSnapshot guards *structural* soundness (do the edges and
// path steps actually reference things that exist?). A consumer runs it as a
// pre-flight so a malformed producer is rejected with a precise reason rather
// than silently mis-processed.

// ProblemKind classifies a structural problem in a record.
type ProblemKind string

const (
	ProblemEmptyID           ProblemKind = "empty_id"
	ProblemDuplicateNode     ProblemKind = "duplicate_node"
	ProblemDuplicateEdge     ProblemKind = "duplicate_edge"
	ProblemDanglingEdge      ProblemKind = "dangling_edge"      // edge endpoint references no node
	ProblemBadExploitability ProblemKind = "bad_exploitability" // outside [0,1]
	ProblemDanglingStep      ProblemKind = "dangling_step"      // path step references no edge
	ProblemEmptyObjective    ProblemKind = "empty_objective"
	ProblemEmptyValue        ProblemKind = "empty_value"    // a discovery component carries no value
	ProblemEmptyObserver     ProblemKind = "empty_observer" // a component names no collector
	ProblemDuplicateRef      ProblemKind = "duplicate_ref"  // two components share a ref
)

// Problem is one structural issue found in a record.
type Problem struct {
	Kind   ProblemKind `json:"kind"`
	Ref    string      `json:"ref"` // the offending id (node/edge/path), if any
	Detail string      `json:"detail"`
}

func (p Problem) String() string {
	if p.Ref != "" {
		return fmt.Sprintf("%s (%s): %s", p.Kind, p.Ref, p.Detail)
	}
	return fmt.Sprintf("%s: %s", p.Kind, p.Detail)
}

// ValidateSnapshot checks an AssessmentSnapshot for referential integrity and
// returns every problem it finds (empty when sound). It does NOT judge security
// posture — only whether the graph is internally consistent:
//   - node ids are present and unique;
//   - edge ids are present and unique, and both endpoints reference a real node;
//   - edge exploitability, when set, is in [0,1];
//   - every path step references a real edge, and each path has an objective.
func ValidateSnapshot(s AssessmentSnapshot) []Problem {
	var problems []Problem

	nodeIDs := make(map[string]bool, len(s.Nodes))
	for _, n := range s.Nodes {
		if n.ID == "" {
			problems = append(problems, Problem{Kind: ProblemEmptyID, Detail: "a node has an empty id"})
			continue
		}
		if nodeIDs[n.ID] {
			problems = append(problems, Problem{Kind: ProblemDuplicateNode, Ref: n.ID, Detail: "node id appears more than once"})
			continue
		}
		nodeIDs[n.ID] = true
	}

	edgeIDs := make(map[string]bool, len(s.Edges))
	for _, e := range s.Edges {
		if e.ID == "" {
			problems = append(problems, Problem{Kind: ProblemEmptyID, Detail: fmt.Sprintf("an edge (%s->%s) has an empty id", e.From, e.To)})
		} else if edgeIDs[e.ID] {
			problems = append(problems, Problem{Kind: ProblemDuplicateEdge, Ref: e.ID, Detail: "edge id appears more than once"})
		} else {
			edgeIDs[e.ID] = true
		}
		if e.From == "" || !nodeIDs[e.From] {
			problems = append(problems, Problem{Kind: ProblemDanglingEdge, Ref: e.ID, Detail: fmt.Sprintf("edge source %q is not a node", e.From)})
		}
		if e.To == "" || !nodeIDs[e.To] {
			problems = append(problems, Problem{Kind: ProblemDanglingEdge, Ref: e.ID, Detail: fmt.Sprintf("edge target %q is not a node", e.To)})
		}
		if e.Exploitability < 0 || e.Exploitability > 1 {
			problems = append(problems, Problem{Kind: ProblemBadExploitability, Ref: e.ID, Detail: fmt.Sprintf("exploitability %.3f is outside [0,1]", e.Exploitability)})
		}
	}

	for pi, path := range s.Paths {
		if strings.TrimSpace(path.Objective) == "" {
			problems = append(problems, Problem{Kind: ProblemEmptyObjective, Ref: fmt.Sprintf("paths[%d]", pi), Detail: "path has no objective"})
		}
		for si, step := range path.Steps {
			if step.EdgeID == "" || !edgeIDs[step.EdgeID] {
				problems = append(problems, Problem{
					Kind:   ProblemDanglingStep,
					Ref:    fmt.Sprintf("paths[%d].steps[%d]", pi, si),
					Detail: fmt.Sprintf("step references edge %q which is not in the graph", step.EdgeID),
				})
			}
		}
	}
	return problems
}

// ValidateDiscoveryBundle checks a DiscoveryBundle for structural soundness and
// returns every problem it finds (empty when sound). A consumer (e.g. the swarm
// warm-start seam) runs it before trusting imported discovery:
//   - every component carries a value and names the collector that observed it;
//   - component refs, when present, are unique.
//
// It does not police the kind vocabulary — producers add component kinds over
// time, and an unknown kind is forward-compatible, not an error.
func ValidateDiscoveryBundle(b DiscoveryBundle) []Problem {
	var problems []Problem
	refs := make(map[string]bool, len(b.Components))
	for i, c := range b.Components {
		where := c.Ref
		if where == "" {
			where = fmt.Sprintf("components[%d]", i)
		}
		if strings.TrimSpace(c.Value) == "" {
			problems = append(problems, Problem{Kind: ProblemEmptyValue, Ref: where, Detail: "component has an empty value"})
		}
		if strings.TrimSpace(c.ObservedBy) == "" {
			problems = append(problems, Problem{Kind: ProblemEmptyObserver, Ref: where, Detail: "component names no observing collector"})
		}
		if c.Ref != "" {
			if refs[c.Ref] {
				problems = append(problems, Problem{Kind: ProblemDuplicateRef, Ref: c.Ref, Detail: "component ref appears more than once"})
			}
			refs[c.Ref] = true
		}
	}
	return problems
}

// Validate returns nil when the bundle is structurally sound, else an error
// summarising the problems.
func (b DiscoveryBundle) Validate() error {
	return problemsError("DiscoveryBundle", ValidateDiscoveryBundle(b))
}

// problemsError renders a problem list as a single error (nil when empty).
func problemsError(kind string, problems []Problem) error {
	if len(problems) == 0 {
		return nil
	}
	parts := make([]string, len(problems))
	for i, p := range problems {
		parts[i] = p.String()
	}
	return fmt.Errorf("invalid %s (%d problem(s)): %s", kind, len(problems), strings.Join(parts, "; "))
}

// Validate returns nil when the snapshot is structurally sound, else an error
// summarising the problems — the convenient one-call form for a consumer that
// just wants to reject a malformed snapshot.
func (s AssessmentSnapshot) Validate() error {
	return problemsError("AssessmentSnapshot", ValidateSnapshot(s))
}
