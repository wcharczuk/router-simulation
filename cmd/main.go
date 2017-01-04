package main

import (
	"fmt"
	"math/rand"
	"time"

	util "github.com/blendlabs/go-util"
	"github.com/wcharczuk/simulation"
)

func simulate(router simulation.Router) {
	println()

	sim := simulation.New(router)
	sim.SetSimulationLength(16 * time.Second)
	sim.SetServerCount(8)
	sim.SetServerWorkerCount(256)
	sim.SetCachedResourceCount(1 << 10)
	sim.SetCachedResourceFetchDuration(16 * time.Millisecond)

	// double number of servers
	sim.AddEvent(8*time.Second, func(s *simulation.Simulation) {
		fmt.Printf("%v :: Doubling server count\n", time.Now().Sub(s.Started))
		for x := 0; x < s.ServerCount(); x++ {
			s.Servers = append(s.Servers, s.CreateServer())
		}
		s.Router.SetServers(s.Servers)
	})
	sim.AddEvent(12*time.Second, func(s *simulation.Simulation) {
		fmt.Printf("%v :: Killing a (random) server\n", time.Now().Sub(s.Started))
		randomIndex := rand.Intn(len(s.Servers))
		s.Servers = append(s.Servers[0:randomIndex], s.Servers[randomIndex+1:]...)
		s.Router.SetServers(s.Servers)
	})

	println(router.Name(), "Starting", fmt.Sprintf("%v", sim.SimulationLength()), "simulation")
	sim.Run()

	var totalTimes []time.Duration
	var workTimes []time.Duration
	var routingTimes []time.Duration
	var misses int
	var notRouted int
	var notCompleted int
	for _, req := range sim.Requests {
		if req.Arrival.IsZero() {
			continue
		}
		if req.Routed.IsZero() {
			notRouted++
			continue
		}
		if req.Completed.IsZero() {
			notCompleted++
			continue
		}
		totalTimes = append(totalTimes, req.Completed.Sub(req.Arrival))
		workTimes = append(workTimes, req.Completed.Sub(req.Routed))
		routingTimes = append(routingTimes, req.Routed.Sub(req.Arrival))

		if req.IsCacheMiss {
			misses++
		}
	}

	totalRequests := len(sim.Requests) - (notRouted + notCompleted)

	println(router.Name(), "Simulation Results")
	println("---------------------------------")
	fmt.Printf("Throughput           : %0.2f rps\n", float64(len(sim.Requests))/float64(sim.SimulationLength()/time.Second))
	fmt.Printf("Average Routing Time : %v\n", util.Math.MeanOfDuration(routingTimes))
	fmt.Printf("Average Work Time    : %v\n", util.Math.MeanOfDuration(workTimes))
	fmt.Printf("Average Total Time   : %v\n", util.Math.MeanOfDuration(totalTimes))
	fmt.Printf("Not Routed           : %d\n", notRouted)
	fmt.Printf("Not Completed        : %d\n", notCompleted)
	fmt.Printf("Cache Miss Rate %d/%d ~= %0.2f%%\n", misses, totalRequests, float64(misses)/float64(totalRequests)*100)
}

func main() {
	simulate(new(simulation.RoundRobinRouter))
	simulate(new(simulation.HashedRouter))
}
