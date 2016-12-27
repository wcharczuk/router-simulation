package main

import (
	"math/rand"
	"time"

	"github.com/blendlabs/go-util/collections"
)

type client struct {
	ArrivalInterval time.Duration
	WorkTime        time.Duration

	Resource  string
	Created   time.Time
	Routed    time.Time
	Completed time.Time

	ServedBy string
}

type server struct {
	ID   string
	Work []*client
}

func (s *server) processWork(current time.Time, tickLength time.Duration) time.Time {
	var remainingWork []*client
	stopAt := current.Add(tickLength)
	for _, workItem := range s.Work {
		if current.Add(workItem.WorkTime).Before(stopAt) {
			workItem.Completed = current.Add(workItem.WorkTime)
			workItem.ServedBy = s.ID
			current = current.Add(workItem.WorkTime)
		} else {
			remainingWork = append(remainingWork, workItem)
		}
	}

	return current
}

type router struct {
	Arrivals   *collections.RingBuffer
	Dispatcher func(*simulation, *client) *server
}

func newSimluation(tickLength, simLength time.Duration, serverCount, resourceCount int, clientArrivalAvg, clientWorkAvg time.Duration, dispatcher func(*simulation, *client) *server) *simulation {
	sim := &simulation{
		TickLength: tickLength,
		Start:      time.Now(),
		End:        time.Now().Add(simLength),
	}

	sim.Router = &router{
		Arrivals:   collections.NewRingBuffer(),
		Dispatcher: dispatcher,
	}

	sim.Servers = map[string]*server{}
	for x := 0; x < serverCount; x++ {
		id := generateID()
		sim.Servers[id] = &server{
			ID: id,
		}
	}

	for x := 0; x < resourceCount; x++ {
		sim.Resources = append(sim.Resources, generateID())
	}
	return sim
}

type simulation struct {
	TickLength time.Duration
	Start      time.Time
	Current    time.Time
	End        time.Time

	Router  *router
	Servers map[string]*server
	Clients []*client

	Resources []string
}

func (s *simulation) getRandomResource() string {
	randomIndex := rand.Intn(len(s.Resources))
	return s.Resources[randomIndex]
}

func (s *simulation) createClient() *client {
	q := generateArrivalTimeInterval()
	return &client{
		Created:  s.Current.Add(q),
		Resource: s.getRandomResource(),
	}
}

func (s *simulation) tick() {
	if s.Current.After(s.End) {
		return
	}

	// schedule new arrivals
	for intraTick := time.Duration(0); intraTick < s.TickLength; {
		c := s.createClient()
		s.Router.Arrivals.Enqueue(c)
		intraTick += c.ArrivalInterval
	}

	// route arrivals in router queue
	for {
		req := s.Router.Arrivals.Peek().(*client)
		if req.Created.Before(s.Current) {
			req = s.Router.Arrivals.Dequeue().(*client)
			req.Routed = s.Current
			srv := s.Router.Dispatcher(s, req)
			srv.Work = append(srv.Work, req)
		} else {
			break
		}
	}

	// process server work
	for _, srv := range s.Servers {
		s.Current = srv.processWork(s.Current, s.TickLength)
	}

	s.Current = s.Current.Add(s.TickLength)
}

func main() {
	// read simulation parameters
	// loop simulation
	// ...
	// print results
}
