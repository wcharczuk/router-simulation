package simulation

// HashedRouter is a router that uses a simple hash of the resource id
// to determine which server to route the request to.
type HashedRouter struct {
	Servers []*Server
}

// Name returns the router name.
func (hr HashedRouter) Name() string { return "Hashed" }

// SetServers sets the server pool.
func (hr *HashedRouter) SetServers(servers []*Server) {
	hr.Servers = servers
}

// Route returns which server to route the request to.
func (hr *HashedRouter) Route(req *Request) *Server {
	hv := HashIdentifier(req.Resource)
	index := int(hv) % len(hr.Servers)
	return hr.Servers[index]
}
