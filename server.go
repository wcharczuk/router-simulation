package simulation

import (
	"sync"
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
	s.ResourceLock.Unlock()
}
