package main

type Semaphore struct {
	semC chan struct{}
}

func NewSemaphore(n int64) *Semaphore {
	w := &Semaphore{semC: make(chan struct{}, n)}
	return w
}

func (s *Semaphore) Wait() {
	s.semC <- struct{}{}
}

func (s *Semaphore) Signal() {
	<-s.semC
}
