package httpadapter

import (
	"net/http/httptest"
	"testing"

	contracts "github.com/Armur-Ai/armur-contracts"
	"github.com/Armur-Ai/armur-contracts/adapter/inproc"
	"github.com/Armur-Ai/armur-contracts/conformance"
)

func TestHTTPConformance(t *testing.T) {
	conformance.RunAdapterSuite(t, func() contracts.Adapter {
		// Fresh inproc store behind a fresh httptest server per adapter.
		store := inproc.New(contracts.NSHayagriva, "DiscoveryBundle")
		srv := httptest.NewServer(Handler(store))
		t.Cleanup(srv.Close)
		return NewClient(srv.URL)
	})
}
