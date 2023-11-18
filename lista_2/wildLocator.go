package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type WildLocator struct {
	logger  *WildLocatorLogger
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

func spawnWildLocator(wg *sync.WaitGroup, lattice *Lattice, v *Vertex, logChannel chan<- LogMessage) {
	wildLocator := WildLocator{x: v.x, y: v.y, lattice: lattice, self: make(chan Message)}
	v.hasWildLocator = true
	v.currentWildLocatorChannel = wildLocator.self
	v.LogWildLocatorSpawned()

	wg.Add(1)

	go func() {
		// setup wildLocator and run it
		wildLocator.updateChannels()
		wildLocator.AttachLogger(logChannel)
		wildLocator.run()

		wg.Done()
	}()
}

func (w *WildLocator) run() {
	ticker := time.NewTicker(tickTime)
	defer ticker.Stop()

	timer := time.NewTimer(WildLocatorLifeTime)
	defer timer.Stop()

	alive := true

	for !shouldQuit.Load() && alive {
		select {
		case <-timer.C:
			// our time to live ended
			trySendMessage(w.current, Message{msgType: MsgWildLocatorDied})
			w.LogWildLocatorDied()
			alive = false
		case <-ticker.C:
			// we should recheck quit variable
		case msg := <-w.self:
			// we got a message from vertex we are in handle it correctly
			switch msg.msgType {
			case MsgWildLocatorEvict:
				moved := w.tryToMove()

				if moved {
					trySendMessage(w.current, Message{msgType: MsgWildLocatorEvictConfirm})
					w.updateChannels()
				} else {
					trySendMessage(w.current, Message{msgType: MsgWildLocatorEvictDeny})
				}
			default:
				fmt.Fprintln(os.Stderr, "ERROR: unrecognized message type received by wild locator:", msg)
			}
		}
	}
}

func (w *WildLocator) tryToMove() bool {
	msg := Message{msgType: MsgWildLocatorEnter, responseChannel: w.self}
	var moved bool
	select {
	case w.north <- msg:
		moved = w.handleResponse(North)
	case w.south <- msg:
		moved = w.handleResponse(South)
	case w.east <- msg:
		moved = w.handleResponse(East)
	case w.west <- msg:
		moved = w.handleResponse(West)
	default:
	}
	return moved
}

func (w *WildLocator) handleResponse(direction LogDirection) bool {
	res := tryRecievMessage(w.self)

	if res == nil {
		return false
	}

	if res.msgType != MsgWildLocatorEnterConfirm {
		fmt.Fprintln(os.Stderr, "ERROR: the vertex didn't confirm entry")
		return false
	}

	w.LogWildLocatorMoved(direction)
	switch direction {
	case North:
		w.y -= 1
	case South:
		w.y += 1
	case West:
		w.x -= 1
	case East:
		w.x += 1
	}

	return true
}

func (w *WildLocator) updateChannels() {
	w.current = w.lattice.vertices[w.y][w.x].outWild

	if w.y > 0 {
		w.north = w.lattice.vertices[w.y-1][w.x].inWild
	} else {
		w.north = nil
	}

	if w.y < w.lattice.m-1 {
		w.south = w.lattice.vertices[w.y+1][w.x].inWild
	} else {
		w.south = nil
	}

	if w.x > 0 {
		w.west = w.lattice.vertices[w.y][w.x-1].inWild
	} else {
		w.west = nil
	}

	if w.x < w.lattice.n-1 {
		w.east = w.lattice.vertices[w.y][w.x+1].inWild
	} else {
		w.east = nil
	}
}
