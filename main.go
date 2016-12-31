package main

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	util "github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/collections"
)

// generateWorkTime uses exprand to generate a reasonable work length
// on the order of milliseconds.
func generateWorkTime() time.Duration {
	randMillis := rand.ExpFloat64() * 20
	return time.Duration(randMillis) * time.Millisecond
}

// generateID returns a new random, 8 character, identifier
func generateID() string {
	return util.String.RandomString(8)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// Request is an incoming work item
type Request struct {
	Arrival   time.Time
	Routed    time.Time
	Served    time.Time
	Completed time.Time

	WorkTime time.Duration

	IsCacheMiss bool

	ID       string
	Resource string
}

// NewServer creates a new server.
func NewServer() *Server {
	return &Server{
		ID:        generateID(),
		Resources: collections.NewSetOfString(),
		Requests:  make(chan *Request, 32),
	}
}

// Server is a server in the dispatch pool.
type Server struct {
	ID string

	resourceLock sync.RWMutex
	Resources    collections.SetOfString

	Requests chan *Request
}

// Run runs the server.
func (s *Server) Run(abort chan bool) {
	for x := 0; x < 4; x++ {
		go s.serve(abort)
	}
}

func (s *Server) serve(abort chan bool) {
	for {
		select {
		case <-abort:
			return
		case req := <-s.Requests:
			s.HandleRequest(req)
		}
	}
}

// HandleRequest processes a request.
func (s *Server) HandleRequest(req *Request) {
	s.resourceLock.RLock()
	if s.Resources.Contains(req.Resource) {
		s.resourceLock.RUnlock()
		req.Completed = time.Now()
		return
	}
	s.resourceLock.RUnlock()
	s.resourceLock.Lock()
	req.IsCacheMiss = true
	time.Sleep(req.WorkTime)
	s.Resources.Add(req.Resource)
	req.Completed = time.Now()
	s.resourceLock.Unlock()
}

// Router is a type that sends requests to relevant backend servers.
type Router interface {
	SetServers([]*Server)
	Route(req *Request) *Server
}

type hashedRouter struct {
	Servers []*Server
}

func (hr *hashedRouter) SetServers(servers []*Server) {
	hr.Servers = servers
}

func (hr *hashedRouter) Route(req *Request) *Server {
	hv := hash(req.Resource)
	index := int(hv) % len(hr.Servers)
	return hr.Servers[index]
}

type roundRobinRouter struct {
	Index   int32
	Servers []*Server
}

func (rrr *roundRobinRouter) SetServers(servers []*Server) {
	rrr.Servers = servers
}

func (rrr *roundRobinRouter) Route(req *Request) *Server {
	server := rrr.Servers[rrr.Index]
	atomic.AddInt32(&rrr.Index, 1)
	if rrr.Index >= int32(len(rrr.Servers)) {
		rrr.Index = 0
	}
	return server
}

// NewResourceSet returns a new set of resources.
func NewResourceSet() []string {
	var resources []string
	for x := 0; x < 1024; x++ {
		resources = append(resources, generateID())
	}
	return resources
}

// NewSimulation returns a new simulation.
func NewSimulation(router Router) *Simulation {
	servers := []*Server{
		NewServer(),
		NewServer(),
		NewServer(),
		NewServer(),
		NewServer(),
		NewServer(),
		NewServer(),
		NewServer(),
	}
	router.SetServers(servers)
	return &Simulation{
		abort:     make(chan bool),
		Router:    router,
		Servers:   servers,
		Resources: NewResourceSet(),
	}
}

// Simulation is the collection of state for the whole simulation.
type Simulation struct {
	Router    Router
	Servers   []*Server
	Resources []string
	Requests  []*Request
	abort     chan bool
}

func (s *Simulation) selectRandomResource() string {
	return s.Resources[rand.Intn(len(s.Resources))]
}

func (s *Simulation) createRequest() *Request {
	req := &Request{
		ID:       generateID(),
		Resource: s.selectRandomResource(),
		WorkTime: generateWorkTime(),
	}
	s.Requests = append(s.Requests, req)
	return req
}

func (s *Simulation) generateArrival() {
	req := s.createRequest()
	req.Arrival = time.Now()
	server := s.Router.Route(req)
	req.Routed = time.Now()
	server.Requests <- req
}

// Run run's the simulation
func (s *Simulation) Run() {
	for _, server := range s.Servers {
		go server.Run(s.abort)
	}

	go func() {
		for {
			select {
			case <-s.abort:
				return
			default:
				s.generateArrival()
			}
		}
	}()
}

// Stop stops the simulation.
func (s *Simulation) Stop() {
	s.abort <- true
}

func simulate(label string, router Router) {
	println()
	println(label, "Starting 10 second simulation")
	sim := NewSimulation(router)
	sim.Run()
	time.Sleep(10 * time.Second)
	sim.Stop()

	var totalTimes []time.Duration
	var workTimes []time.Duration
	var routingTimes []time.Duration
	var misses int
	for _, req := range sim.Requests {
		totalTimes = append(totalTimes, req.Completed.Sub(req.Arrival))
		workTimes = append(workTimes, req.Completed.Sub(req.Routed))
		routingTimes = append(routingTimes, req.Routed.Sub(req.Arrival))
		if req.IsCacheMiss {
			misses++
		}
	}
	println(label, "Simulation Results")
	fmt.Printf("Throughput %0.2f rps\n", float64(len(sim.Requests))/10.0)
	fmt.Printf("Average Total Time %v\n", util.Math.MeanOfDuration(totalTimes))
	fmt.Printf("Average Work Time %v\n", util.Math.MeanOfDuration(workTimes))
	fmt.Printf("Average Routing Time %v\n", util.Math.MeanOfDuration(routingTimes))
	fmt.Printf("Cache Miss Rate %d/%d ~= %0.2f%%\n", misses, len(sim.Requests), float64(misses)/float64(len(sim.Requests))*100)
}

func main() {
	simulate("Round Robin", new(roundRobinRouter))
	simulate("Hashed", new(hashedRouter))
}
