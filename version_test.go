package contracts

import "testing"

func TestParseVersion(t *testing.T) {
	ok := map[string]Version{
		"0.1.0":   {0, 1, 0},
		"1.2.3":   {1, 2, 3},
		"2.0":     {2, 0, 0},
		" 1.4.0 ": {1, 4, 0},
	}
	for in, want := range ok {
		got, err := ParseVersion(in)
		if err != nil || got != want {
			t.Errorf("ParseVersion(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "1", "x.y.z", "1.-1.0", "1.2.3.4-pre"} {
		if _, err := ParseVersion(bad); err == nil {
			t.Errorf("ParseVersion(%q): expected error", bad)
		}
	}
}

func TestCheckSchema(t *testing.T) {
	// Pin expectations to the current SchemaVersion (0.x line).
	cases := []struct {
		producer string
		want     CompatLevel
		wantErr  bool
	}{
		{SchemaVersion, CompatExact, false},
		{"0.0.9", CompatConsumerAhead, false},  // consumer (0.1) knows more
		{"0.99.0", CompatProducerAhead, false}, // producer emits fields we ignore
		{"1.0.0", CompatIncompatible, true},    // different major → reject
		{"", CompatIncompatible, true},
		{"garbage", CompatIncompatible, true},
	}
	for _, c := range cases {
		lvl, err := CheckSchema(c.producer)
		if (err != nil) != c.wantErr || lvl != c.want {
			t.Errorf("CheckSchema(%q) = %v, err=%v; want %v, wantErr=%v",
				c.producer, lvl, err, c.want, c.wantErr)
		}
		if c.want == CompatIncompatible && lvl.OK() {
			t.Errorf("CheckSchema(%q): OK() should be false", c.producer)
		}
	}
}

func TestCheckEnvelope(t *testing.T) {
	env, err := NewEnvelope(NSHayagriva, Producer{Product: "hayagriva"}, "DiscoveryBundle", DiscoveryBundle{BundleID: "b"})
	if err != nil {
		t.Fatal(err)
	}
	lvl, err := CheckEnvelope(env)
	if err != nil || lvl != CompatExact {
		t.Fatalf("CheckEnvelope(self-emitted) = %v, %v; want exact, nil", lvl, err)
	}
	// A foreign envelope from a future major must be rejected.
	env.Version = "2.0.0"
	if lvl, err := CheckEnvelope(env); err == nil || lvl.OK() {
		t.Fatalf("CheckEnvelope(major+1) = %v, %v; want incompatible, err", lvl, err)
	}
}
