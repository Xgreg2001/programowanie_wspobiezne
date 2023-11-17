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
	logger   *VertexLogger
	id       int
	x        int
	y        int
	occupied bool
	in       chan ExplorerMessage
	out      chan ExplorerMessage
}

func (v Vertex) run(wg *sync.WaitGroup, explorerCount *atomic.Uint64, quit *atomic.Bool, maxExplorers int, logChannel chan<- LogMessage, lattice *Lattice) {
	v.AttachLogger(logChannel)

	for !quit.Load() {
		spawnTimer := time.NewTimer(spawnExplorerTick)

		if !v.occupied {
			// we don't currently have an explorer, so we can either spawn one or accept one from a neighbor
			select {
			case msg := <-v.in:
				switch msg.msgType {
				case ExplorerMessageEnter:
					v.occupied = true
					v.LogExplorerReceived(msg.id)
				default:
					fmt.Fprintln(os.Stderr, "ERROR: We should only receive ExplorerMessageEnter here")
				}
			case <-spawnTimer.C:
				if rand.Float64() < spawnExplorerRate {
					spawnExplorer(wg, quit, lattice, explorerCount, maxExplorers, &v, logChannel)
				}
			}
		} else {
			select {
			case msg := <-v.out:
				if msg.msgType == ExplorerMessageLeave {
					v.occupied = false
					v.LogExplorerLeft(msg.id)
				} else {
					fmt.Fprintln(os.Stderr, "ERROR: We should only receive ExplorerMessageLeave here")
				}
			case <-spawnTimer.C:
				// this ensures that thread don't hang after all explorers close
			}

		}
		spawnTimer.Stop()
	}
}

func spawnExplorer(wg *sync.WaitGroup, quit *atomic.Bool, lattice *Lattice, explorerCount *atomic.Uint64, maxExplorers int, v *Vertex, logChannel chan<- LogMessage) {
	for {
		currentCount := explorerCount.Load()
		if currentCount < uint64(maxExplorers-1) {
			if explorerCount.CompareAndSwap(currentCount, currentCount+1) {
				id := currentCount + 1
				v.occupied = true
				explorer := &Explorer{id: int(id), x: v.x, y: v.y, lattice: lattice}
				explorer.updateChannels()
				go func() {
					wg.Add(1)
					explorer.run(quit, logChannel)
					wg.Done()
				}()
				v.LogExplorerSpawned(explorer.id)
				break
			} else {
				continue
			}
		} else {
			break
		}
	}
}

func CreateLattice(n, m int) Lattice {
	// create all channels first and then create all vertices
	incomingChannels := make([]chan ExplorerMessage, n*m)
	outgoingChannels := make([]chan ExplorerMessage, n*m)

	for i := 0; i < n*m; i++ {
		incomingChannels[i] = make(chan ExplorerMessage)
		outgoingChannels[i] = make(chan ExplorerMessage)
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
