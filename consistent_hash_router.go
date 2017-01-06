package simulation

import (
	"github.com/blendlabs/go-util/collections"
)

// NewConsistentHashRouter returns a new consistent hash router.
func NewConsistentHashRouter(resourceTotal int, c float64) *ConsistentHashRouter {
	return &ConsistentHashRouter{
		ResourceTotal: resourceTotal,
		C:             c,
	}
}

// ConsistentHashRouter implemenets a chr router.
type ConsistentHashRouter struct {
	C              float64
	ResourceTotal  int
	Servers        []*Server
	ResourceCounts map[string]collections.SetOfString
}

// Name returns the router name.
func (chr ConsistentHashRouter) Name() string { return "Consistent Hash" }

// SetServers sets the server pool.
func (chr *ConsistentHashRouter) SetServers(servers []*Server) {
	chr.Servers = servers
	chr.ResourceCounts = make(map[string]collections.SetOfString)
	for _, s := range servers {
		chr.ResourceCounts[s.ID] = collections.NewSetOfString()
	}
}

func (chr *ConsistentHashRouter) computeLoad(serverID string) float64 {
	resources, hasResources := chr.ResourceCounts[serverID]
	if !hasResources {
		return 0.0
	}
	m := float64(chr.ResourceTotal)
	n := float64(len(chr.Servers))
	max := (chr.C * m) / n
	specificLoad := float64(resources.Len()) / float64(chr.ResourceTotal)
	return specificLoad / max
}

// Route returns which server to route the request to.
func (chr *ConsistentHashRouter) Route(req *Request) *Server {
	hv := HashIdentifier(req.Resource)
	index := int(hv) % len(chr.Servers)
	server := chr.Servers[index]
	for chr.computeLoad(server.ID) >= 1.0 {
		index++
		if index >= len(chr.Servers) {
			index = 0
		}
		server = chr.Servers[index]
	}
	// potentially duplicative
	chr.ResourceCounts[server.ID].Add(req.Resource)
	return server
}
