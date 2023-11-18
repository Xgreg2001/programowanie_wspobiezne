package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Lattice struct {
	vertices [][]Vertex
	n        int
	m        int
}

type Vertex struct {
	logger                    *VertexLogger
	id                        int
	x                         int
	y                         int
	hasExplorer               bool
	hasWildLocator            bool
	hazardous                 bool
	currentWildLocatorChannel chan Message
	in                        chan Message
	out                       chan Message
	inWild                    chan Message
	outWild                   chan Message
}

func (v Vertex) run(explorerWg *sync.WaitGroup, explorerStats *ExplorerStats, wildLocatorWg *sync.WaitGroup, maxExplorers int, logChannel chan<- LogMessage, lattice *Lattice) {
	v.AttachLogger(logChannel)
	ticker := time.NewTicker(tickTime)
	hazardTimer := time.NewTimer(hazardLifeTime)
	hazardTimer.Stop()

	for !shouldQuit.Load() {

		if !v.hasExplorer && !v.hasWildLocator {
			// we don't currently have an explorer or wild locator so we can either spawn one of them or accept one from a neighbor
			select {
			case msg := <-v.in:
				switch msg.msgType {
				case MsgExplorerEnter:
					v.handleMsgExplorerEnter(msg)
				default:
					fmt.Fprintln(os.Stderr, "ERROR: We should only receive MsgExplorerEnter here")
				}
			case msg := <-v.inWild:
				if msg.msgType == MsgWildLocatorEnter {
					response := Message{msgType: MsgWildLocatorEnterConfirm}
					ok := trySendMessage(msg.responseChannel, response)
					if ok {
						v.hasWildLocator = true
						v.currentWildLocatorChannel = msg.responseChannel
					}
				} else {
					fmt.Fprintln(os.Stderr, "ERROR: We should only receive MsgWildLocatorEnter here")
				}
			case <-ticker.C:
				r := rand.Float64()
				if !v.hazardous {
					if r < spawnExplorerRate {
						spawnExplorer(explorerWg, lattice, explorerStats, maxExplorers, &v, logChannel)
						continue
					}

					r -= spawnExplorerRate
					if r < spawnHazardRate {
						v.hazardous = true
						hazardTimer.Reset(hazardLifeTime)
						v.LogHazardSpawned()
						continue
					}

					r -= spawnHazardRate
					if r < spawnWildLocatorRate {
						spawnWildLocator(wildLocatorWg, lattice, &v, logChannel)
						continue
					}
				} else {
					// we can't spawn explorers or hazards if we already have a hazard
					if r < spawnWildLocatorRate {
						spawnWildLocator(wildLocatorWg, lattice, &v, logChannel)
						continue
					}
				}
			case <-hazardTimer.C:
				v.hazardous = false
				v.LogHazardDisappeared()
			}
		} else if v.hasExplorer {
			select {
			case msg := <-v.out:
				if msg.msgType == MsgExplorerLeave {
					v.hasExplorer = false
					v.LogExplorerLeft(msg.expId)
				} else {
					fmt.Fprintln(os.Stderr, "ERROR: We should only receive MsgExplorerLeave here:", msg)
				}
			case <-ticker.C:
				// this ensures that thread don't hang after all explorers close
			}

		} else if v.hasWildLocator {
			// we have a wild locator so we need to listen to its messages as well we need to be able to accept a incoming explorer
			select {
			case msg := <-v.outWild:
				if msg.msgType == MsgWildLocatorDied {
					v.hasWildLocator = false
					v.currentWildLocatorChannel = nil
				} else {
					fmt.Fprintln(os.Stderr, "ERROR: We should only recieve MsgWildLocatorDied here:", msg)
				}
			case msg := <-v.in:
				if msg.msgType == MsgExplorerEnter {
					evicted := v.tryEvictLocator(msg)

					if evicted {
						v.handleMsgExplorerEnter(msg)
					} else {
						trySendMessage(msg.responseChannel, Message{msgType: MsgExplorerEnterDeny})
					}
				} else {
					fmt.Fprintln(os.Stderr, "ERROR: We should only recieve MsgExplorerEnter here:", msg)
				}

			case <-ticker.C:
				//this ensures we don't hang after all other threads close
			}

		}
	}
	ticker.Stop()
}

func (v *Vertex) tryEvictLocator(msg Message) bool {
	request := Message{msgType: MsgWildLocatorEvict}
	send := trySendMessage(v.currentWildLocatorChannel, request)

	if !send {
		return false
	}

	respond := tryRecievMessage(v.outWild)
	evicted := false

	switch respond.msgType {
	case MsgWildLocatorEvictConfirm:
		v.hasWildLocator = false
		v.currentWildLocatorChannel = nil
		evicted = true
	case MsgWildLocatorEvictDeny:
		// there is nothing we can do ;-;
	default:
		fmt.Fprintln(os.Stderr, "ERROR: locator didn't confirm or deny eviction request:", respond)
	}

	return evicted
}

func (v *Vertex) handleMsgExplorerEnter(msg Message) {
	if !v.hazardous {
		response := Message{msgType: MsgExplorerEnterConfirm}
		ok := trySendMessage(msg.responseChannel, response)
		if ok {
			v.hasExplorer = true
			v.LogExplorerReceived(msg.expId)
		}
	} else {
		response := Message{msgType: MsgExplorerEnterHazard}
		ok := trySendMessage(msg.responseChannel, response)
		if ok {
			v.hazardous = false
			v.LogMsgExplorerEnteredHazard(msg.expId)
		}
	}
}

func CreateLattice(n, m int) Lattice {
	// create all channels first and then create all vertices
	incomingChannels := make([]chan Message, n*m)
	outgoingChannels := make([]chan Message, n*m)
	incomingWildChannels := make([]chan Message, n*m)
	outgoingWildChannels := make([]chan Message, n*m)

	for i := 0; i < n*m; i++ {
		incomingChannels[i] = make(chan Message)
		outgoingChannels[i] = make(chan Message)
		incomingWildChannels[i] = make(chan Message)
		outgoingWildChannels[i] = make(chan Message)

	}

	vertices := make([][]Vertex, n)
	for y := 0; y < m; y++ {
		vertices[y] = make([]Vertex, n)
		for x := 0; x < n; x++ {
			id := y*n + x
			vertices[y][x] = Vertex{
				id:      id,
				x:       x,
				y:       y,
				in:      incomingChannels[id],
				out:     outgoingChannels[id],
				inWild:  incomingWildChannels[id],
				outWild: outgoingWildChannels[id],
			}
		}
	}

	return Lattice{vertices: vertices, n: n, m: m}
}
