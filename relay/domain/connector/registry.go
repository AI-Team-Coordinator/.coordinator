package connector

import (
	"fmt"
	"sort"
	"strings"
)

// Registry holds connectors enabled for this Relay process.
type Registry struct {
	byID map[string]Connector
}

func NewRegistry(connectors ...Connector) (*Registry, error) {
	r := &Registry{byID: make(map[string]Connector)}
	for _, c := range connectors {
		if c == nil {
			continue
		}
		if err := r.Register(c); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Registry) Register(c Connector) error {
	if r == nil || c == nil {
		return fmt.Errorf("connector is nil")
	}
	id := strings.TrimSpace(c.Descriptor().ID)
	if id == "" {
		return fmt.Errorf("connector id is empty")
	}
	if _, dup := r.byID[id]; dup {
		return fmt.Errorf("connector %q already registered", id)
	}
	r.byID[id] = c
	return nil
}

func (r *Registry) Get(id string) (Connector, bool) {
	if r == nil {
		return nil, false
	}
	c, ok := r.byID[strings.TrimSpace(id)]
	return c, ok
}

func (r *Registry) ListDescriptors() []Descriptor {
	if r == nil {
		return nil
	}
	out := make([]Descriptor, 0, len(r.byID))
	for _, c := range r.byID {
		out = append(out, c.Descriptor())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
