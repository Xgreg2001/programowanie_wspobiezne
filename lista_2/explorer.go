package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Explorer struct {
	logger  *ExplorerLogger
	id      int
	lattice *Lattice
	x       int
	y       int
	self    chan Message
	current chan<- Message
	north   chan<- Message
	south   chan<- Message
	east    chan<- Message
	west    chan<- Message
}

func spawnExplorer(wg *sync.WaitGroup, lattice *Lattice, explorerStats *ExplorerStats, maxExplorers int, v *Vertex, logChannel chan<- LogMessage) {
	explorerStats.mu.Lock()
	if explorerStats.count < maxExplorers {
		expId := explorerStats.nextId
		explorerStats.nextId += 1
		if explorerStats.nextId == 100 {
			explorerStats.nextId = 1
		}
		explorerStats.count += 1
		explorerStats.mu.Unlock()

		explorer := Explorer{id: expId, x: v.x, y: v.y, lattice: lattice, self: make(chan Message)}
		v.hasExplorer = true
		v.LogExplorerSpawned(expId)
		wg.Add(1)

		go func() {
			// setup explorer and run it
			explorer.updateChannels()
			explorer.AttachLogger(logChannel)
			explorer.run()

			// cleanup after the finish
			explorerStats.mu.Lock()
			explorerStats.count -= 1
			explorerStats.mu.Unlock()

			wg.Done()
		}()
	} else {
		explorerStats.mu.Unlock()
	}
}

func (e *Explorer) run() {
	ticker := time.NewTicker(tickTime)
	defer ticker.Stop()

	for !shouldQuit.Load() {
		<-ticker.C

		if rand.Float64() < moveExplorerRate {
			var moved bool
			var alive bool
			msg := Message{msgType: MsgExplorerEnter, expId: e.id, responseChannel: e.self}
			select {
			case e.north <- msg:
				alive, moved = e.handleResponse(North)
			case e.south <- msg:
				alive, moved = e.handleResponse(South)
			case e.east <- msg:
				alive, moved = e.handleResponse(East)
			case e.west <- msg:
				alive, moved = e.handleResponse(West)
			default:
				// no neighbor is available, so we reset the timer
				moved = false
				alive = true
			}

			if !alive {
				break
			}

			if moved {
				trySendMessage(e.current, Message{msgType: MsgExplorerLeave, expId: e.id})
				e.updateChannels()
			}
		}
	}
}

func (e *Explorer) handleResponse(direction LogDirection) (bool, bool) {
	moved := false

	res := tryRecievMessage(e.self)

	if res == nil {
		return true, false
	}

	switch res.msgType {
	case MsgExplorerEnterConfirm:
		e.LogExplorerMoved(direction)
		switch direction {
		case North:
			e.y -= 1
		case South:
			e.y += 1
		case West:
			e.x -= 1
		case East:
			e.x += 1
		default:
			fmt.Fprintln(os.Stderr, "ERROR: Incorrect direction parameter!")
		}
		moved = true
	case MsgExplorerEnterHazard:
		e.LogExplorerDied()
		trySendMessage(e.current, Message{msgType: MsgExplorerLeave, expId: e.id})
		return false, moved
	case MsgExplorerEnterDeny:
		// I guess we couldn't enter XD
		moved = false
		return true, moved
	default:
		fmt.Fprintln(os.Stderr, "ERROR: this type of message should not be handled here:", res)
	}
	return true, moved
}

func (e *Explorer) updateChannels() {
	e.current = e.lattice.vertices[e.y][e.x].out

	if e.y > 0 {
		e.north = e.lattice.vertices[e.y-1][e.x].in
	} else {
		e.north = nil
	}

	if e.y < e.lattice.m-1 {
		e.south = e.lattice.vertices[e.y+1][e.x].in
	} else {
		e.south = nil
	}

	if e.x > 0 {
		e.west = e.lattice.vertices[e.y][e.x-1].in
	} else {
		e.west = nil
	}

	if e.x < e.lattice.n-1 {
		e.east = e.lattice.vertices[e.y][e.x+1].in
	} else {
		e.east = nil
	}
}
