package main

import "sync"

type ReaderWriterMonitor struct {
	readers     int
	writing     bool
	wantToWrite int
	wantToRead  int
	prio        bool
	mu          sync.Mutex
	okToRead    sync.Cond
	okToWrite   sync.Cond
}

func NewMonitor() *ReaderWriterMonitor {
	return &ReaderWriterMonitor{mu: sync.Mutex{}, okToRead: sync.Cond{L: &sync.Mutex{}}, okToWrite: sync.Cond{L: &sync.Mutex{}}}
}

func (m *ReaderWriterMonitor) StartRead() {
	m.okToRead.L.Lock()
	m.mu.Lock()
	m.wantToRead += 1
	for m.writing || (m.wantToWrite > 0 && !m.prio) {
		m.mu.Unlock()
		m.okToRead.Wait()
		m.mu.Lock()
	}
	m.readers += 1
	m.wantToRead -= 1
	m.okToRead.Signal()
	m.mu.Unlock()
	m.okToRead.L.Unlock()
}

func (m *ReaderWriterMonitor) StopRead() {
	m.mu.Lock()
	m.readers -= 1
	if m.readers == 0 {
		m.okToWrite.Signal()
	}
	m.prio = false
	m.mu.Unlock()
}

func (m *ReaderWriterMonitor) StartWrite() {
	m.okToWrite.L.Lock()
	m.mu.Lock()
	m.wantToWrite += 1
	for m.readers > 0 || m.writing || (m.wantToRead > 0 && m.prio) {
		m.mu.Unlock()
		m.okToWrite.Wait()
		m.mu.Lock()
	}
	m.writing = true
	m.wantToWrite -= 1
	m.mu.Unlock()
	m.okToWrite.L.Unlock()
}

func (m *ReaderWriterMonitor) StopWrite() {
	m.mu.Lock()
	m.writing = false
	m.prio = true
	m.okToRead.Signal()
	m.okToWrite.Signal()
	m.mu.Unlock()
}
