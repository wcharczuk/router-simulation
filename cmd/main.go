package main

import (
	"fmt"
	"time"

	util "github.com/blendlabs/go-util"
	"github.com/wcharczuk/simulation"
)

func simulate(router simulation.Router) {
	println()
	println(router.Name(), "Starting 10 second simulation")
	sim := simulation.New(router)
	sim.SetServerCount(8)
	sim.SetServerWorkerCount(8)
	sim.SetCachedResourceFetchDuration(30 * time.Millisecond)
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
	println(router.Name(), "Simulation Results")
	fmt.Printf("Throughput %0.2f rps\n", float64(len(sim.Requests))/float64(sim.SimulationLength()/time.Second))
	fmt.Printf("Average Total Time %v\n", util.Math.MeanOfDuration(totalTimes))
	fmt.Printf("Average Work Time %v\n", util.Math.MeanOfDuration(workTimes))
	fmt.Printf("Average Routing Time %v\n", util.Math.MeanOfDuration(routingTimes))
	fmt.Printf("Not Routed %d, Not Completed %d\n", notRouted, notCompleted)
	fmt.Printf("Cache Miss Rate %d/%d ~= %0.2f%%\n", misses, len(sim.Requests), float64(misses)/float64(len(sim.Requests))*100)
}

func main() {
	simulate(new(simulation.RoundRobinRouter))
	simulate(new(simulation.HashedRouter))
}
