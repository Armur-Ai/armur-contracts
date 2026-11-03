package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Digest is the canonical content digest of an exchanged record: a stable hash
// over (producer, kind, payload). Adapters use it for idempotency — the same
// bytes always produce the same digest and therefore the same receipt — so all
// transports agree on record identity.
func Digest(env Envelope) string {
	h := sha256.New()
	h.Write([]byte(env.Producer.Product))
	h.Write([]byte{0})
	h.Write([]byte(env.Kind))
	h.Write([]byte{0})
	b, _ := json.Marshal(env.Payload)
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
