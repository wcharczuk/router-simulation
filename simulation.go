package simulation

import (
	"math/rand"
	"sort"
	"time"

	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/collections"
)

// New returns a new simulation.
func New(router Router) *Simulation {
	return &Simulation{
		Events: collections.NewRingBuffer(),
		abort:  make(chan bool),
		Router: router,
	}
}

type simulationEvent struct {
	At     time.Duration
	Action func(*Simulation)
}

type simulationEvents []interface{}

func (se simulationEvents) Len() int { return len(se) }
func (se simulationEvents) Less(a, b int) bool {
	return se[a].(simulationEvent).At < se[b].(simulationEvent).At
}
func (se simulationEvents) Swap(a, b int) { se[a], se[b] = se[b], se[a] }

// Simulation is the collection of state for the whole simulation.
type Simulation struct {
	Router    Router
	Servers   []*Server
	Resources []string
	Requests  []*Request
	Events    *collections.RingBuffer

	Started time.Time
	Stopped time.Time

	simulationLength            time.Duration
	serverCount                 int
	serverWorkerCount           int
	cachedResourceCount         int
	cachedResourceFetchDuration time.Duration

	abort chan bool
}

// AddEvent adds an event to the simulation at a given offset from start.
func (s *Simulation) AddEvent(at time.Duration, action func(*Simulation)) {
	events := s.Events.AsSlice()
	events = append(events, simulationEvent{At: at, Action: action})
	sort.Sort(simulationEvents(events))
	s.Events = collections.NewRingBufferFromSlice(events)
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

// ServerCount is the number of servers.
func (s Simulation) ServerCount() int {
	if s.serverCount == 0 {
		return 8
	}
	return s.serverCount
}

// SetServerCount sets the number of servers.
func (s *Simulation) SetServerCount(count int) {
	s.serverCount = count
}

// ServerWorkerCount is the number of goroutines per server.
func (s Simulation) ServerWorkerCount() int {
	if s.serverWorkerCount == 0 {
		return 8
	}
	return s.serverWorkerCount
}

// SetServerWorkerCount sets the number of goroutines per server.
func (s *Simulation) SetServerWorkerCount(count int) {
	s.serverWorkerCount = count
}

// CachedResourceCount is the number of cached resources the system tracks.
func (s Simulation) CachedResourceCount() int {
	if s.cachedResourceCount == 0 {
		return 1 << 10
	}
	return s.cachedResourceCount
}

// SetCachedResourceCount sets the cached resource count.
func (s *Simulation) SetCachedResourceCount(count int) {
	s.cachedResourceCount = count
}

// CachedResourceFetchDuration is the time it takes to complete a cache miss.
func (s Simulation) CachedResourceFetchDuration() time.Duration {
	if s.cachedResourceFetchDuration == 0 {
		return 20 * time.Millisecond
	}
	return s.cachedResourceFetchDuration
}

// SetCachedResourceFetchDuration sets the cache miss penalty.
func (s *Simulation) SetCachedResourceFetchDuration(fetchDuration time.Duration) {
	s.cachedResourceFetchDuration = fetchDuration
}

// Initialize initializes the simulation.
func (s *Simulation) Initialize() {
	var servers []*Server
	for x := 0; x < s.ServerCount(); x++ {
		servers = append(servers, s.CreateServer())
	}
	s.Router.SetServers(servers)
	s.Servers = servers
	s.Resources = s.newResourceSet()
}

// CreateServer creates a new server.
func (s Simulation) CreateServer() *Server {
	return &Server{
		ID:        util.UUIDv4().ToShortString(),
		Resources: collections.NewSetOfString(),
		Requests:  make(chan *Request, s.ServerWorkerCount()),
	}
}

func (s Simulation) newResourceSet() []string {
	var resources []string
	for x := 0; x < s.CachedResourceCount(); x++ {
		resources = append(resources, util.UUIDv4().ToShortString())
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

func (s *Simulation) processEvent() {
	if s.Events.Len() == 0 {
		return
	}
	event := s.Events.Peek().(simulationEvent)
	if event.At < time.Now().Sub(s.Started) {
		event.Action(s)
		s.Events.Dequeue()
	}
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

	// arrivals
	go func() {
		for {
			select {
			case <-s.abort:
				return
			default:
				req := s.createRequest()
				req.Arrival = time.Now()
				server := s.Router.Route(req)
				req.Routed = time.Now()
				server.Requests <- req
			}
		}
	}()

	// events
	go func() {
		for {
			select {
			case <-s.abort:
				return
			default:
				s.processEvent()
			}
		}
	}()

	s.Started = time.Now()
}

// Stop stops the simulation.
func (s *Simulation) Stop() {
	s.abort <- true
	s.Stopped = time.Now()
}
