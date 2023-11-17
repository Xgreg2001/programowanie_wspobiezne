package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Lattice struct {
	vertices [][]Vertex
	n        int
	m        int
}

type Vertex struct {
	logger    *VertexLogger
	id        int
	x         int
	y         int
	occupied  bool
	hazardous bool
	in        chan Message
	out       chan Message
}

func (v Vertex) run(wg *sync.WaitGroup, explorerStats *ExplorerStats, quit *atomic.Bool, maxExplorers int, logChannel chan<- LogMessage, lattice *Lattice) {
	v.AttachLogger(logChannel)
	ticker := time.NewTicker(tickTime)
	hazardTimer := time.NewTimer(hazardLifeTime)
	hazardTimer.Stop()

	for !quit.Load() {

		if !v.occupied {
			// we don't currently have an explorer, so we can either spawn one or accept one from a neighbor
			select {
			case msg := <-v.in:
				switch msg.msgType {
				case ExplorerMessageEnter:
					if !v.hazardous {
						v.occupied = true
						v.in <- Message{msgType: ExplorerMessageEnterConfirm}
						v.LogExplorerReceived(msg.expId)
					} else {
						v.hazardous = false
						v.in <- Message{msgType: ExplorerMessegeEnterHazard}
						v.LogMsgExplorerEnteredHazard(msg.expId)
					}
				default:
					fmt.Fprintln(os.Stderr, "ERROR: We should only receive ExplorerMessageEnter here")
				}
			case <-ticker.C:
				r := rand.Float64()
				if r < spawnExplorerRate && !v.hazardous {
					spawnExplorer(wg, quit, lattice, explorerStats, maxExplorers, &v, logChannel)
					continue
				}

				r -= spawnExplorerRate
				if r < spawnHazardRate && !v.hazardous {
					v.hazardous = true
					hazardTimer.Reset(hazardLifeTime)
					v.LogHazardSpawned()
					continue
				}
			case <-hazardTimer.C:
				v.hazardous = false
				v.LogHazardDisappeared()
			}
		} else {
			select {
			case msg := <-v.out:
				if msg.msgType == ExplorerMessageLeave {
					v.occupied = false
					v.LogExplorerLeft(msg.expId)
				} else {
					fmt.Fprintln(os.Stderr, "ERROR: We should only receive ExplorerMessageLeave here")
				}
			case <-ticker.C:
				// this ensures that thread don't hang after all explorers close
			}

		}
	}
	ticker.Stop()
}

func spawnExplorer(wg *sync.WaitGroup, quit *atomic.Bool, lattice *Lattice, explorerStats *ExplorerStats, maxExplorers int, v *Vertex, logChannel chan<- LogMessage) {
	explorerStats.mu.Lock()
	if explorerStats.count < maxExplorers {
		expId := explorerStats.nextId
		explorerStats.nextId += 1
		if explorerStats.nextId == 100 {
			explorerStats.nextId = 1
		}
		explorerStats.count += 1
		explorerStats.mu.Unlock()

		explorer := Explorer{id: expId, x: v.x, y: v.y, lattice: lattice}
		explorer.updateChannels()
		v.occupied = true
		v.LogExplorerSpawned(expId)
		wg.Add(1)

		go func() {
			explorer.run(quit, logChannel)
			explorerStats.mu.Lock()
			explorerStats.count -= 1
			explorerStats.mu.Unlock()
			wg.Done()
		}()
	} else {
		explorerStats.mu.Unlock()
	}
}

func CreateLattice(n, m int) Lattice {
	// create all channels first and then create all vertices
	incomingChannels := make([]chan Message, n*m)
	outgoingChannels := make([]chan Message, n*m)

	for i := 0; i < n*m; i++ {
		incomingChannels[i] = make(chan Message)
		outgoingChannels[i] = make(chan Message)
	}

	vertices := make([][]Vertex, n)
	for y := 0; y < m; y++ {
		vertices[y] = make([]Vertex, n)
		for x := 0; x < n; x++ {
			vertices[y][x] = Vertex{id: y*n + x, x: x, y: y, in: incomingChannels[y*n+x], out: outgoingChannels[y*n+x]}
		}
	}

	return Lattice{vertices: vertices, n: n, m: m}
}
