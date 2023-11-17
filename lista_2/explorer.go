package main

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type Explorer struct {
	logger  *ExplorerLogger
	id      int
	lattice *Lattice
	x       int
	y       int
	current chan<- ExplorerMessage
	north   chan<- ExplorerMessage
	south   chan<- ExplorerMessage
	east    chan<- ExplorerMessage
	west    chan<- ExplorerMessage
}

type ExplorerMessageType int

type ExplorerMessage struct {
	msgType ExplorerMessageType
	id      int
}

const (
	ExplorerMessageEnter ExplorerMessageType = iota
	ExplorerMessageLeave
	ExplorerMessageReady
)

func (e *Explorer) run(quit *atomic.Bool, logChannel chan<- LogMessage) {
	e.AttachLogger(logChannel)

	for !quit.Load() {
		exploreTimer := time.NewTimer(moveExplorerTick)

		<-exploreTimer.C

		if rand.Float64() < moveExplorerRate {
			moved := false
			// try to move to one of the neighboring vertices
			msg := ExplorerMessage{msgType: ExplorerMessageEnter, id: e.id}
			select {
			case e.north <- msg:
				e.LogExplorerMoved(North)
				e.y -= 1
				moved = true
			case e.south <- msg:
				e.LogExplorerMoved(South)
				e.y += 1
				moved = true
			case e.east <- msg:
				e.LogExplorerMoved(East)
				e.x += 1
				moved = true
			case e.west <- msg:
				e.LogExplorerMoved(West)
				e.x -= 1
				moved = true
			default:
				// no neighbor is available, so we reset the timer
			}

			if moved {
				e.current <- ExplorerMessage{msgType: ExplorerMessageLeave, id: e.id}
				e.updateChannels()
			}
		}

		exploreTimer.Stop()
	}
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
