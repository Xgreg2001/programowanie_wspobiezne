package main

import "sync"

type PhilMonitor struct {
	mu      sync.Mutex
	forks   []int
	okToEat []sync.Cond
	n       int
}

func NewMonitor(n int) *PhilMonitor {
	forks := make([]int, n)
	for i := range forks {
		forks[i] = 2
	}
	okToEat := make([]sync.Cond, n)
	for i := range okToEat {
		okToEat[i] = sync.Cond{L: &sync.Mutex{}}
	}
	return &PhilMonitor{mu: sync.Mutex{}, forks: forks, okToEat: okToEat, n: n}
}

func (m *PhilMonitor) TakeFork(i int) {
	m.mu.Lock()
	m.okToEat[i].L.Lock()
	for m.forks[i] != 2 {
		m.mu.Unlock()
		m.okToEat[i].Wait()
		m.mu.Lock()
	}
	m.forks[(i+1)%m.n] = m.forks[(i+1)%m.n] - 1
	m.forks[(i-1+m.n)%m.n] = m.forks[(i-1+m.n)%m.n] - 1
	m.okToEat[i].L.Unlock()
	m.mu.Unlock()
}

func (m *PhilMonitor) ReleaseFork(i int) {
	m.mu.Lock()
	m.forks[(i+1)%m.n] = m.forks[(i+1)%m.n] + 1
	m.forks[(i-1+m.n)%m.n] = m.forks[(i-1+m.n)%m.n] + 1
	if m.forks[(i+1)%m.n] == 2 {
		m.okToEat[(i+1)%m.n].Signal()
	}
	if m.forks[(i-1+m.n)%m.n] == 2 {
		m.okToEat[(i-1+m.n)%m.n].Signal()
	}
	m.mu.Unlock()
}
