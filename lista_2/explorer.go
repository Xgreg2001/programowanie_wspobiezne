package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"time"
)

type Explorer struct {
	logger  *ExplorerLogger
	id      int
	lattice *Lattice
	x       int
	y       int
	current chan Message
	north   chan Message
	south   chan Message
	east    chan Message
	west    chan Message
}

type MessageType int

type Message struct {
	msgType MessageType
	expId   int
}

const (
	ExplorerMessageEnter MessageType = iota
	ExplorerMessageEnterConfirm
	ExplorerMessegeEnterHazard
	ExplorerMessageLeave
	ExplorerMessageReady
)

func (e *Explorer) run(quit *atomic.Bool, logChannel chan<- LogMessage) {
	e.AttachLogger(logChannel)

L:
	for !quit.Load() {
		timer := time.NewTimer(tickTime)

		<-timer.C

		if rand.Float64() < moveExplorerRate {
			moved := false
			// try to move to one of the neighboring vertices
			msg := Message{msgType: ExplorerMessageEnter, expId: e.id}
			select {
			case e.north <- msg:
				res := <-e.north
				switch res.msgType {
				case ExplorerMessageEnterConfirm:
					e.LogExplorerMoved(North)
					e.y -= 1
					moved = true
				case ExplorerMessegeEnterHazard:
					e.LogExplorerDied()
					e.current <- Message{msgType: ExplorerMessageLeave, expId: e.id}
					break L
				default:
					fmt.Fprintln(os.Stderr, "ERROR: this type of message should not be handled here:", res)
				}
			case e.south <- msg:
				res := <-e.south
				switch res.msgType {
				case ExplorerMessageEnterConfirm:
					e.LogExplorerMoved(South)
					e.y += 1
					moved = true
				case ExplorerMessegeEnterHazard:
					e.LogExplorerDied()
					e.current <- Message{msgType: ExplorerMessageLeave, expId: e.id}
					break L
				default:
					fmt.Fprintln(os.Stderr, "ERROR: this type of message should not be handled here:", res)
				}
			case e.east <- msg:
				res := <-e.east
				switch res.msgType {
				case ExplorerMessageEnterConfirm:
					e.LogExplorerMoved(East)
					e.x += 1
					moved = true
				case ExplorerMessegeEnterHazard:
					e.LogExplorerDied()
					e.current <- Message{msgType: ExplorerMessageLeave, expId: e.id}
					break L
				default:
					fmt.Fprintln(os.Stderr, "ERROR: this type of message should not be handled here:", res)
				}
			case e.west <- msg:
				res := <-e.west
				switch res.msgType {
				case ExplorerMessageEnterConfirm:
					e.LogExplorerMoved(West)
					e.x -= 1
					moved = true
				case ExplorerMessegeEnterHazard:
					e.LogExplorerDied()
					e.current <- Message{msgType: ExplorerMessageLeave, expId: e.id}
					break L
				default:
					fmt.Fprintln(os.Stderr, "ERROR: this type of message should not be handled here:", res)
				}
			default:
				// no neighbor is available, so we reset the timer
			}

			if moved {
				e.current <- Message{msgType: ExplorerMessageLeave, expId: e.id}
				e.updateChannels()
			}
		}

		timer.Stop()
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
