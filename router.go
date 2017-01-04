package simulation

// Router is a type that sends requests to relevant backend servers.
type Router interface {
	Name() string
	SetServers([]*Server)
	Route(req *Request) *Server
}
