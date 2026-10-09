package wmdesign

import (
	"fmt"
	"math"
	"strings"
)

// ConnectorRoutingPolicy retains the reviewed allocation routing policy.
// Clearance applies to centerlines and allocation envelopes, not glyph ink.
type ConnectorRoutingPolicy struct {
	Clearance  float64           `json:"clearance_pt"`
	Exclusions map[string]string `json:"exclusions,omitempty"`
	Reserved   []Rect            `json:"reserved,omitempty"`
}

func validateConnectorRoutingPolicy(p *ConnectorRoutingPolicy) error {
	if p == nil {
		return nil
	}
	if math.IsNaN(p.Clearance) || math.IsInf(p.Clearance, 0) || p.Clearance < 0 || p.Clearance > 72 || len(p.Exclusions) > 128 || len(p.Reserved) > 64 {
		return fmt.Errorf("scene.invalid_connector_routing_policy")
	}
	for name, reason := range p.Exclusions {
		if name == "" || len(name) > 512 || strings.TrimSpace(reason) == "" || len(reason) > 4096 {
			return fmt.Errorf("scene.invalid_routing_exclusion")
		}
	}
	for _, r := range p.Reserved {
		if !intakeFinite(r.X, r.Y, r.W, r.H) || r.W <= 0 || r.H <= 0 {
			return fmt.Errorf("scene.invalid_routing_reserved_allocation")
		}
	}
	return nil
}
