package simulation

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/blendlabs/go-util/collections"
)

// Server is a server in the dispatch pool.
type Server struct {
	ID           string
	Requests     chan *Request
	ResourceLock sync.RWMutex
	WorkerCount  int
	Resources    collections.SetOfString

	TotalServed int32
}

// Run runs the server.
func (s *Server) Run(abort chan bool) {
	for x := 0; x < s.WorkerCount; x++ {
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
	atomic.AddInt32(&s.TotalServed, 1)
	s.ResourceLock.RLock()
	if s.Resources.Contains(req.Resource) {
		s.ResourceLock.RUnlock()
		req.Completed = time.Now()
		return
	}
	s.ResourceLock.RUnlock()
	s.ResourceLock.Lock()
	req.IsCacheMiss = true
	time.Sleep(req.WorkTime)
	s.Resources.Add(req.Resource)
	req.Completed = time.Now()
	req.ServedBy = s.ID
	s.ResourceLock.Unlock()
}
