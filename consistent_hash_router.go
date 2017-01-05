package simulation

// NewConsistentHashRouter returns a new consistent hash router.
func NewConsistentHashRouter(c float64) *ConsistentHashRouter {
	return &ConsistentHashRouter{
		C: c,
	}
}

// ConsistentHashRouter implemenets a chr router.
type ConsistentHashRouter struct {
	C          float64
	Servers    []*Server
	Capacities []int
}

// Name returns the router name.
func (chr ConsistentHashRouter) Name() string { return "Consistent Hash" }

// SetServers sets the server pool.
func (chr *ConsistentHashRouter) SetServers(servers []*Server) {
	chr.Servers = servers
	chr.Capacities = make([]int, len(servers))
}

// Route returns which server to route the request to.
func (chr *ConsistentHashRouter) Route(req *Request) *Server {
	return nil
}
