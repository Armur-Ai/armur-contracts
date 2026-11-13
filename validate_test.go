package contracts

import "testing"

func soundSnapshot() AssessmentSnapshot {
	return AssessmentSnapshot{
		SnapshotID: "s1",
		Nodes: []GraphNode{
			{ID: "entry", Type: "network"},
			{ID: "a", Type: "service"},
			{ID: "objective", Type: "data"},
		},
		Edges: []GraphEdge{
			{ID: "e1", From: "entry", To: "a", Kind: "reaches", Exploitability: 0.5},
			{ID: "e2", From: "a", To: "objective", Kind: "escalates_to", Exploitability: 1.0},
		},
		Paths: []AttackPath{
			{Objective: "objective", Steps: []PathStep{{EdgeID: "e1"}, {EdgeID: "e2"}}, Complete: true},
		},
	}
}

func TestValidateSnapshot_SoundIsClean(t *testing.T) {
	if problems := ValidateSnapshot(soundSnapshot()); len(problems) != 0 {
		t.Fatalf("expected no problems, got %v", problems)
	}
	if err := soundSnapshot().Validate(); err != nil {
		t.Fatalf("sound snapshot should validate: %v", err)
	}
}

func TestValidateSnapshot_DanglingEdge(t *testing.T) {
	s := soundSnapshot()
	s.Edges = append(s.Edges, GraphEdge{ID: "e3", From: "a", To: "ghost", Kind: "reaches"})
	problems := ValidateSnapshot(s)
	found := false
	for _, p := range problems {
		if p.Kind == ProblemDanglingEdge && p.Ref == "e3" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a dangling_edge problem for e3->ghost, got %v", problems)
	}
	if s.Validate() == nil {
		t.Error("Validate() must return an error for a dangling edge")
	}
}

func TestValidateSnapshot_DanglingStep(t *testing.T) {
	s := soundSnapshot()
	s.Paths[0].Steps = append(s.Paths[0].Steps, PathStep{EdgeID: "nope"})
	problems := ValidateSnapshot(s)
	found := false
	for _, p := range problems {
		if p.Kind == ProblemDanglingStep {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a dangling_step problem, got %v", problems)
	}
}

func TestValidateSnapshot_DuplicatesAndBadWeight(t *testing.T) {
	s := soundSnapshot()
	s.Nodes = append(s.Nodes, GraphNode{ID: "a", Type: "service"})             // dup node id
	s.Edges = append(s.Edges, GraphEdge{ID: "e1", From: "a", To: "objective"}) // dup edge id
	s.Edges[0].Exploitability = 1.5                                            // out of range
	kinds := map[ProblemKind]bool{}
	for _, p := range ValidateSnapshot(s) {
		kinds[p.Kind] = true
	}
	for _, want := range []ProblemKind{ProblemDuplicateNode, ProblemDuplicateEdge, ProblemBadExploitability} {
		if !kinds[want] {
			t.Errorf("expected a %s problem; got kinds %v", want, kinds)
		}
	}
}

func TestValidateSnapshot_EmptyObjective(t *testing.T) {
	s := soundSnapshot()
	s.Paths[0].Objective = "   "
	problems := ValidateSnapshot(s)
	found := false
	for _, p := range problems {
		if p.Kind == ProblemEmptyObjective {
			found = true
		}
	}
	if !found {
		t.Errorf("blank objective should be flagged, got %v", problems)
	}
}

func TestValidateDiscoveryBundle(t *testing.T) {
	good := DiscoveryBundle{
		BundleID: "b1",
		Components: []ComponentObservation{
			{Ref: "httpx:https://a/x", Kind: "endpoint", Value: "https://a/x", ObservedBy: "httpx-import"},
		},
	}
	if err := good.Validate(); err != nil {
		t.Fatalf("sound bundle should validate: %v", err)
	}

	bad := DiscoveryBundle{
		Components: []ComponentObservation{
			{Ref: "r1", Kind: "endpoint", Value: "", ObservedBy: "x"},          // empty value
			{Ref: "r1", Kind: "endpoint", Value: "https://b", ObservedBy: "x"}, // dup ref r1
		},
	}
	kinds := map[ProblemKind]bool{}
	for _, p := range ValidateDiscoveryBundle(bad) {
		kinds[p.Kind] = true
	}
	for _, want := range []ProblemKind{ProblemEmptyValue, ProblemDuplicateRef} {
		if !kinds[want] {
			t.Errorf("expected a %s problem; got %v", want, kinds)
		}
	}
	// A missing ObservedBy is provenance-only and must NOT fail the gate.
	okMissingObserver := DiscoveryBundle{Components: []ComponentObservation{
		{Ref: "r9", Kind: "endpoint", Value: "https://a", ObservedBy: ""},
	}}
	if err := okMissingObserver.Validate(); err != nil {
		t.Errorf("missing ObservedBy should not fail validation: %v", err)
	}
	if bad.Validate() == nil {
		t.Error("Validate() must return an error for the malformed bundle")
	}
}
