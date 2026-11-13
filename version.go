package contracts

import (
	"fmt"
	"strconv"
	"strings"
)

// This file adds a compatibility gate on top of SchemaVersion so a consumer can
// decide, deterministically, whether an envelope it received was produced
// against a schema it can safely decode — instead of every product
// re-implementing an ad-hoc version check (or, as happened in practice, pinning
// drifting versions with no check at all).
//
// The rule follows the package's additive-versioning contract (see doc.go):
//
//   - Same MAJOR  → compatible. Minor bumps only ADD optional fields, so a
//     consumer can always decode a same-major envelope; unknown newer fields are
//     ignored, missing newer fields decode to zero values.
//   - Different MAJOR → incompatible. A major bump is the one case allowed to
//     remove or repurpose fields, so the payload may no longer mean what the
//     consumer expects. Reject rather than silently mis-decode.

// Version is a parsed semantic version (major.minor.patch). Only the numeric
// core is modelled; pre-release/build metadata is not used by the schema.
type Version struct{ Major, Minor, Patch int }

// ParseVersion parses a "MAJOR.MINOR.PATCH" (or "MAJOR.MINOR") string. It is
// deliberately strict: a malformed version is an error, never a silent zero.
func ParseVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, fmt.Errorf("contracts: empty schema version")
	}
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return Version{}, fmt.Errorf("contracts: malformed schema version %q (want MAJOR.MINOR[.PATCH])", s)
	}
	out := Version{}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("contracts: malformed schema version %q", s)
		}
		switch i {
		case 0:
			out.Major = n
		case 1:
			out.Minor = n
		case 2:
			out.Patch = n
		}
	}
	return out, nil
}

// String renders the version back to "MAJOR.MINOR.PATCH".
func (v Version) String() string { return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch) }

// CompatLevel is the outcome of comparing a producer's schema version against
// this build's SchemaVersion.
type CompatLevel int

const (
	// CompatExact — identical major.minor. No field-set difference.
	CompatExact CompatLevel = iota
	// CompatConsumerAhead — same major, this consumer's minor is higher than the
	// producer's. The consumer understands every field the producer could emit;
	// newer optional fields simply arrive empty. Fully safe.
	CompatConsumerAhead
	// CompatProducerAhead — same major, the producer's minor is higher. The
	// producer may include optional fields this consumer does not model; they are
	// ignored on decode. Safe under the additive rule, but worth surfacing.
	CompatProducerAhead
	// CompatIncompatible — different major. Reject.
	CompatIncompatible
)

func (c CompatLevel) String() string {
	switch c {
	case CompatExact:
		return "exact"
	case CompatConsumerAhead:
		return "consumer-ahead"
	case CompatProducerAhead:
		return "producer-ahead"
	case CompatIncompatible:
		return "incompatible"
	default:
		return "unknown"
	}
}

// OK reports whether the level permits decoding (anything but incompatible).
func (c CompatLevel) OK() bool { return c != CompatIncompatible }

// CheckSchema compares a producer's declared schema version against this build's
// SchemaVersion. The returned error is non-nil only when the versions are
// incompatible (different major) or the producer version is unparseable; the
// CompatLevel is still meaningful in the error-free cases and callers may log a
// CompatProducerAhead as a soft warning.
func CheckSchema(producerVersion string) (CompatLevel, error) {
	pv, err := ParseVersion(producerVersion)
	if err != nil {
		return CompatIncompatible, err
	}
	cv, err := ParseVersion(SchemaVersion)
	if err != nil {
		// SchemaVersion is a compile-time constant; this should never happen.
		return CompatIncompatible, fmt.Errorf("contracts: invalid SchemaVersion %q: %w", SchemaVersion, err)
	}
	if pv.Major != cv.Major {
		return CompatIncompatible, fmt.Errorf(
			"contracts: incompatible schema major: producer %s, consumer %s", pv, cv)
	}
	switch {
	case pv.Minor == cv.Minor:
		return CompatExact, nil
	case cv.Minor > pv.Minor:
		return CompatConsumerAhead, nil
	default:
		return CompatProducerAhead, nil
	}
}

// CheckEnvelope applies CheckSchema to an envelope's Version field. Use it at
// the consumption boundary (e.g. right after decoding a DiscoveryBundle
// envelope) to fail fast on an incompatible producer rather than mis-decoding.
func CheckEnvelope(e Envelope) (CompatLevel, error) {
	return CheckSchema(e.Version)
}
