// Package httpadapter is the authenticated-HTTP implementation of
// contracts.Adapter: a Client that speaks to a Handler over HTTP. It behaves
// identically to the in-process and file adapters (verified by the shared
// conformance suite), so a product can be embedded, dropped a file, or reached
// over the network without changing semantics.
//
// Transport mapping: Submit = POST /submit, Fetch = GET /fetch?ref=…,
// Negotiate = GET /negotiate. An unknown ref is 404 -> ErrAbsentPeer; an
// unsupported kind is 422 -> ErrUnsupportedKind.
package httpadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	contracts "github.com/Armur-Ai/armur-contracts"
)

// Handler returns an http.Handler backed by the given store (any Adapter, e.g.
// an inproc or file adapter). It is the server half of the transport.
func Handler(store contracts.Adapter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var env contracts.Envelope
		if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rec, err := store.Submit(r.Context(), env)
		if err == contracts.ErrUnsupportedKind {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, rec)
	})
	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		ref := contracts.JobRef(r.URL.Query().Get("ref"))
		env, err := store.Fetch(r.Context(), ref)
		if err == contracts.ErrAbsentPeer {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, env)
	})
	mux.HandleFunc("/negotiate", func(w http.ResponseWriter, r *http.Request) {
		cs, err := store.Negotiate(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, cs)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// Client is the HTTP client half, implementing contracts.Adapter.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient builds a client for the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTP: http.DefaultClient}
}

func (c *Client) Submit(ctx context.Context, env contracts.Envelope) (contracts.Receipt, error) {
	body, err := json.Marshal(env)
	if err != nil {
		return contracts.Receipt{}, err
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/submit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return contracts.Receipt{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnprocessableEntity {
		return contracts.Receipt{}, contracts.ErrUnsupportedKind
	}
	if resp.StatusCode != http.StatusOK {
		return contracts.Receipt{}, fmt.Errorf("httpadapter: submit status %d", resp.StatusCode)
	}
	var rec contracts.Receipt
	return rec, json.NewDecoder(resp.Body).Decode(&rec)
}

func (c *Client) Fetch(ctx context.Context, ref contracts.JobRef) (contracts.Envelope, error) {
	u := c.BaseURL + "/fetch?ref=" + url.QueryEscape(string(ref))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return contracts.Envelope{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return contracts.Envelope{}, contracts.ErrAbsentPeer
	}
	if resp.StatusCode != http.StatusOK {
		return contracts.Envelope{}, fmt.Errorf("httpadapter: fetch status %d", resp.StatusCode)
	}
	var env contracts.Envelope
	return env, json.NewDecoder(resp.Body).Decode(&env)
}

func (c *Client) Negotiate(ctx context.Context) (contracts.CapabilitySet, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/negotiate", nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return contracts.CapabilitySet{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return contracts.CapabilitySet{}, fmt.Errorf("httpadapter: negotiate status %d: %s", resp.StatusCode, b)
	}
	var cs contracts.CapabilitySet
	return cs, json.NewDecoder(resp.Body).Decode(&cs)
}

var _ contracts.Adapter = (*Client)(nil)
