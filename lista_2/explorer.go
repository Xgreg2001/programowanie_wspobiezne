package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"time"
)

type Explorer struct {
	logger   *ExplorerLogger
	id       int
	lattice  *Lattice
	x        int
	y        int
	responds chan Message
	current  chan<- Message
	north    chan<- Message
	south    chan<- Message
	east     chan<- Message
	west     chan<- Message
}

type MessageType int

type Message struct {
	msgType         MessageType
	expId           int
	responseChannel chan<- Message
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
	ticker := time.NewTicker(tickTime)
	defer ticker.Stop()

	for !quit.Load() {
		<-ticker.C

		if rand.Float64() < moveExplorerRate {
			var moved bool
			var alive bool
			msg := Message{msgType: ExplorerMessageEnter, expId: e.id, responseChannel: e.responds}
			select {
			case e.north <- msg:
				alive, moved = e.handleResponse(quit, North)
			case e.south <- msg:
				alive, moved = e.handleResponse(quit, South)
			case e.east <- msg:
				alive, moved = e.handleResponse(quit, East)
			case e.west <- msg:
				alive, moved = e.handleResponse(quit, West)
			default:
				// no neighbor is available, so we reset the timer
				moved = false
				alive = true
			}

			if !alive {
				break
			}

			if moved {
				e.current <- Message{msgType: ExplorerMessageLeave, expId: e.id}
				e.updateChannels()
			}
		}
	}
}

func (e *Explorer) handleResponse(quit *atomic.Bool, direction LogDirection) (bool, bool) {
	moved := false

	res := tryRecievMessage(e.responds, quit)

	if res == nil {
		return true, false
	}

	switch res.msgType {
	case ExplorerMessageEnterConfirm:
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
	case ExplorerMessegeEnterHazard:
		e.LogExplorerDied()
		trySendMessage(e.current, Message{msgType: ExplorerMessageLeave, expId: e.id}, quit)
		return false, moved
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
