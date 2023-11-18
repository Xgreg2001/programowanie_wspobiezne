package main

import (
	"sync/atomic"
	"time"
)

func trySendMessage(channel chan<- Message, message Message, shouldQuit *atomic.Bool) bool {
	for !shouldQuit.Load() {
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case channel <- message:
			return true
		case <-timer.C:
			// recheck quit variable
		}
	}
	return false
}

func tryRecievMessage(channel <-chan Message, shouldQuit *atomic.Bool) *Message {
	for !shouldQuit.Load() {
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case response := <-channel:
			// we recieved a response so we can procced
			return &response
		case <-timer.C:
			// rerun quit variable check
		}
	}
	return nil
}
