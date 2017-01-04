package simulation

import "sync/atomic"

// RoundRobinRouter routes based on an even distribution through the server registry.
// In respect to the cached identifier it is "random" routing.
type RoundRobinRouter struct {
	Index   int32
	Servers []*Server
}

// Name returns the router name.
func (rrr RoundRobinRouter) Name() string {
	return "Round Robin"
}

// SetServers sets the server registry.
func (rrr *RoundRobinRouter) SetServers(servers []*Server) {
	rrr.Servers = servers
}

// Route returns which server to send the request to.
func (rrr *RoundRobinRouter) Route(req *Request) *Server {
	server := rrr.Servers[rrr.Index]
	atomic.AddInt32(&rrr.Index, 1)
	if rrr.Index >= int32(len(rrr.Servers)) {
		rrr.Index = 0
	}
	return server
}
