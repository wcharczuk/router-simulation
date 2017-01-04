package simulation

import (
	"math/rand"
	"time"

	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/collections"
)

// New returns a new simulation.
func New(router Router) *Simulation {
	return &Simulation{
		abort:  make(chan bool),
		Router: router,
	}
}

// Simulation is the collection of state for the whole simulation.
type Simulation struct {
	Router    Router
	Servers   []*Server
	Resources []string
	Requests  []*Request

	simulationLength               time.Duration
	serverCount                    int
	serverWorkerCount              int
	serverIdentifierLength         int
	cachedResourceCount            int
	cachedResourceIdentifierLength int
	cachedResourceFetchDuration    time.Duration

	abort chan bool
}

// SimulationLength is the length of the simulation.
func (s Simulation) SimulationLength() time.Duration {
	if s.simulationLength == 0 {
		return 10 * time.Second
	}
	return s.simulationLength
}

// SetSimulationLength sets the length of the simulation
func (s *Simulation) SetSimulationLength(length time.Duration) {
	s.simulationLength = length
}

func (s Simulation) ServerCount() int {
	if s.serverCount == 0 {
		return 8
	}
	return s.serverCount
}

func (s *Simulation) SetServerCount(count int) {
	s.serverCount = count
}

func (s Simulation) ServerWorkerCount() int {
	if s.serverWorkerCount == 0 {
		return 8
	}
	return s.serverWorkerCount
}

func (s *Simulation) SetServerWorkerCount(count int) {
	s.serverWorkerCount = count
}

func (s Simulation) ServerIdentifierLength() int {
	if s.serverIdentifierLength == 0 {
		return 4
	}
	return s.serverIdentifierLength
}

func (s *Simulation) SetServerIdentifierLength(length int) {
	s.serverIdentifierLength = length
}

func (s Simulation) CachedResourceCount() int {
	if s.cachedResourceCount == 0 {
		return 1 << 10
	}
	return s.cachedResourceCount
}

func (s *Simulation) SetCachedResourceCount(count int) {
	s.cachedResourceCount = count
}

func (s Simulation) CachedResourceIdentifierLength() int {
	if s.cachedResourceIdentifierLength == 0 {
		return 8
	}
	return s.cachedResourceIdentifierLength
}

func (s *Simulation) CachedResourceFetchDuration() time.Duration {
	if s.cachedResourceFetchDuration == 0 {
		return 20 * time.Millisecond
	}
	return s.cachedResourceFetchDuration
}

func (s *Simulation) SetCachedResourceFetchDuration(fetchDuration time.Duration) {
	s.cachedResourceFetchDuration = fetchDuration
}

// Initialize initializes the simulation.
func (s *Simulation) Initialize() {
	var servers []*Server
	for x := 0; x < s.ServerCount(); x++ {
		servers = append(servers, s.newServer())
	}
	s.Router.SetServers(servers)
	s.Servers = servers
	s.Resources = s.newResourceSet()
}

// NewServer creates a new server.
func (s Simulation) newServer() *Server {
	return &Server{
		ID:        util.String.RandomString(s.ServerIdentifierLength()),
		Resources: collections.NewSetOfString(),
		Requests:  make(chan *Request, 32),
	}
}

func (s Simulation) newResourceSet() []string {
	var resources []string
	for x := 0; x < s.CachedResourceCount(); x++ {
		resources = append(resources, util.String.RandomString(s.CachedResourceIdentifierLength()))
	}
	return resources
}

func (s *Simulation) selectRandomResource() string {
	return s.Resources[rand.Intn(len(s.Resources))]
}

func (s *Simulation) createRequest() *Request {
	req := &Request{
		Resource: s.selectRandomResource(),
		WorkTime: GenerateWorkTime(s.CachedResourceFetchDuration()),
	}
	s.Requests = append(s.Requests, req)
	return req
}

func (s *Simulation) arrival() {
	req := s.createRequest()
	req.Arrival = time.Now()
	server := s.Router.Route(req)
	req.Routed = time.Now()
	server.Requests <- req
}

// Run starts and stops the simulation.
func (s *Simulation) Run() {
	s.Initialize()
	s.Start()
	time.Sleep(s.SimulationLength())
	s.Stop()
}

// Start start's the simulation.
func (s *Simulation) Start() {
	for _, server := range s.Servers {
		server.Run(s.abort)
	}

	// the below is essentially a firehose of requests
	// we try to saturate the system to see what the
	// throughput at max is.
	go func() {
		for {
			select {
			case <-s.abort:
				return
			default:
				s.arrival()
			}
		}
	}()
}

// Stop stops the simulation.
func (s *Simulation) Stop() {
	s.abort <- true
}
