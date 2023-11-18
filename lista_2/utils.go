package main

import (
	"time"
)

type MessageType int

type Message struct {
	msgType         MessageType
	expId           int
	responseChannel chan Message
}

const (
	MsgExplorerEnter MessageType = iota
	MsgExplorerEnterConfirm
	MsgExplorerEnterDeny
	MsgExplorerEnterHazard
	MsgExplorerLeave
	MsgExplorerReady
	MsgWildLocatorDied
	MsgWildLocatorEvict
	MsgWildLocatorEnter
	MsgWildLocatorEnterConfirm
	MsgWildLocatorEvictConfirm
	MsgWildLocatorEvictDeny
)

func trySendMessage(channel chan<- Message, message Message) bool {
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

func tryRecievMessage(channel <-chan Message) *Message {
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
