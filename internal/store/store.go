// Package store maps NEXO domain entities onto PrismaTec-Core organisms.
// Pattern aligned with smart-wms wms-twin corestore:
// organism Name = "nexo.<kind>.<id>", payload in working memory + episodic history.
package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yecharlot/NexoWork/internal/domain"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
)

const domainPrefix = "nexo"

type Store struct {
	manager *organism.Manager
	nodeID  string
	mu      sync.Mutex
}

func New(manager *organism.Manager, nodeID string) *Store {
	return &Store{manager: manager, nodeID: nodeID}
}

func orgName(kind, id string) string {
	return domainPrefix + "." + kind + "." + id
}

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano()%1_000_000_000_000)
}

func (s *Store) getOrCreateOrganism(kind, id string) (*organism.Organism, error) {
	name := orgName(kind, id)
	org, err := s.manager.Get(name)
	if err != nil {
		if !strings.Contains(err.Error(), "organism not found") {
			return nil, err
		}
		org, err = s.manager.Create(organism.CreateOptions{
			Name: name,
			Capabilities: []organism.Capability{
				"memory.read", "memory.write", "nexo.read", "nexo.write",
			},
			Policy: organism.Policy{Rules: map[string]bool{
				"memory.read": true, "memory.write": true,
				"nexo.read": true, "nexo.write": true,
			}},
			NodeID:    s.nodeID,
			CreatedBy: s.nodeID,
		})
		if err != nil {
			return nil, fmt.Errorf("create organism %s: %w", name, err)
		}
	}
	if org.Placement.Primary == s.nodeID && org.Status != organism.StatusRunning {
		org, err = s.manager.Start(org.ID)
		if err != nil {
			return nil, err
		}
	}
	return org, nil
}

func (s *Store) savePayload(kind, id string, payload any, episodeType string) (*organism.Organism, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	org, err := s.getOrCreateOrganism(kind, id)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if _, err := s.manager.PutMemory(org.ID, "payload", string(raw)); err != nil {
		// fallback by name
		if _, err2 := s.manager.PutMemory(orgName(kind, id), "payload", string(raw)); err2 != nil {
			return nil, err
		}
	}
	if _, err := s.manager.PutMemory(org.ID, "kind", kind); err != nil {
		_ = err
	}
	if _, err := s.manager.PutMemory(org.ID, "domain_id", id); err != nil {
		_ = err
	}
	if _, err := s.manager.AppendEpisode(org.ID, episodeType, map[string]any{
		"kind": kind, "id": id,
	}, raw); err != nil {
		_ = err
	}
	return s.manager.Get(org.ID)
}

func (s *Store) loadPayload(kind, id string, dst any) error {
	org, err := s.manager.Get(orgName(kind, id))
	if err != nil {
		return err
	}
	val := ""
	if org.Memory.Working != nil {
		val = org.Memory.Working["payload"]
	}
	if val == "" {
		return fmt.Errorf("not found: %s/%s", kind, id)
	}
	return json.Unmarshal([]byte(val), dst)
}

func (s *Store) listKind(kind string) []json.RawMessage {
	out := make([]json.RawMessage, 0)
	prefix := domainPrefix + "." + kind + "."
	for _, org := range s.manager.List() {
		if org == nil {
			continue
		}
		if !strings.HasPrefix(org.Name, prefix) && !strings.HasPrefix(org.ID, prefix) {
			continue
		}
		val := ""
		if org.Memory.Working != nil {
			val = org.Memory.Working["payload"]
		}
		if val == "" {
			continue
		}
		var generic map[string]any
		if json.Unmarshal([]byte(val), &generic) == nil {
			generic["root_cid"] = org.RootCID
			if b, err := json.Marshal(generic); err == nil {
				out = append(out, b)
				continue
			}
		}
		out = append(out, json.RawMessage(val))
	}
	sort.Slice(out, func(i, j int) bool { return string(out[i]) > string(out[j]) })
	return out
}

func (s *Store) CreateRequest(r domain.ServiceRequest) (domain.ServiceRequest, error) {
	if r.ID == "" {
		r.ID = newID("req")
	}
	now := time.Now().UTC()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now
	if r.Status == "" {
		r.Status = domain.StatusActive
	}
	org, err := s.savePayload(domain.KindRequest, r.ID, r, "nexo.request.created")
	if err != nil {
		return r, err
	}
	if org != nil {
		r.RootCID = org.RootCID
	}
	return r, nil
}

func (s *Store) CreateProvider(p domain.ServiceProvider) (domain.ServiceProvider, error) {
	if p.ID == "" {
		p.ID = newID("prv")
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = domain.StatusActive
	}
	org, err := s.savePayload(domain.KindProvider, p.ID, p, "nexo.provider.created")
	if err != nil {
		return p, err
	}
	if org != nil {
		p.RootCID = org.RootCID
	}
	return p, nil
}

func (s *Store) CreateMatch(m domain.Match) (domain.Match, error) {
	if m.ID == "" {
		m.ID = newID("mch")
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
	if m.Status == "" {
		m.Status = domain.StatusMatched
	}
	org, err := s.savePayload(domain.KindMatch, m.ID, m, "nexo.match.created")
	if err != nil {
		return m, err
	}
	if org != nil {
		m.RootCID = org.RootCID
	}
	var req domain.ServiceRequest
	if err := s.loadPayload(domain.KindRequest, m.RequestID, &req); err == nil {
		req.Status = domain.StatusMatched
		req.UpdatedAt = now
		_, _ = s.savePayload(domain.KindRequest, req.ID, req, "nexo.request.matched")
	}
	return m, nil
}

func (s *Store) UpdateRequestStatus(id, status string) (domain.ServiceRequest, error) {
	var r domain.ServiceRequest
	if err := s.loadPayload(domain.KindRequest, id, &r); err != nil {
		return r, err
	}
	r.Status = status
	r.UpdatedAt = time.Now().UTC()
	org, err := s.savePayload(domain.KindRequest, r.ID, r, "nexo.request.status")
	if err != nil {
		return r, err
	}
	if org != nil {
		r.RootCID = org.RootCID
	}
	return r, nil
}

func (s *Store) ListRequests() ([]domain.ServiceRequest, error) {
	raw := s.listKind(domain.KindRequest)
	out := make([]domain.ServiceRequest, 0, len(raw))
	for _, b := range raw {
		var r domain.ServiceRequest
		if json.Unmarshal(b, &r) == nil {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Store) ListProviders() ([]domain.ServiceProvider, error) {
	raw := s.listKind(domain.KindProvider)
	out := make([]domain.ServiceProvider, 0, len(raw))
	for _, b := range raw {
		var p domain.ServiceProvider
		if json.Unmarshal(b, &p) == nil {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Store) ListMatches() ([]domain.Match, error) {
	raw := s.listKind(domain.KindMatch)
	out := make([]domain.Match, 0, len(raw))
	for _, b := range raw {
		var m domain.Match
		if json.Unmarshal(b, &m) == nil {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *Store) GetRequest(id string) (domain.ServiceRequest, error) {
	var r domain.ServiceRequest
	err := s.loadPayload(domain.KindRequest, id, &r)
	return r, err
}

func (s *Store) GetProvider(id string) (domain.ServiceProvider, error) {
	var p domain.ServiceProvider
	err := s.loadPayload(domain.KindProvider, id, &p)
	return p, err
}
