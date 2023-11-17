package main

import (
	"fmt"
	"time"
)

const (
	TERM_RESET = "\033[0m"
	TERM_RED   = "\033[31m"
)

type Camera struct {
	cameraChanel <-chan CameraMessage
	board        [][]Explorer
	n            int
	m            int
	crossedEdges [][]bool
}

type CameraMessage struct {
	messageType CameraMessageType
	x           int
	y           int
	expId       int
	xHelper     int
	yHelper     int
}

type CameraMessageType int

const (
	CamExplorerSpawned CameraMessageType = iota
	CamExplorerMoved
)

func RecordSpawnExplorer(expId, x, y int) CameraMessage {
	return CameraMessage{expId: expId, x: x, y: y, messageType: CamExplorerSpawned}
}

func RecordMoveExplorer(expId, fromX, fromY, toX, toY int) CameraMessage {
	return CameraMessage{expId: expId, x: fromX, y: fromY, xHelper: toX, yHelper: toY, messageType: CamExplorerMoved}
}

func (c Camera) PrintBoard() {
	c.PrintBoardSeparator()
	bottomRow := "+"
	for y := 0; y < c.m; y++ {
		fmt.Print("|")
		for x := 0; x < c.n; x++ {
			vertId := y*c.n + x

			if c.board[y][x].id != 0 {
				fmt.Printf("%02d", c.board[y][x].id)
			} else {
				fmt.Printf("  ")
			}

			if x < c.n-1 {
				if c.crossedEdges[vertId][vertId+1] {
					fmt.Printf("%s|%s", TERM_RED, TERM_RESET)
				} else {
					fmt.Printf(" ")
				}
			} else {
				fmt.Println("|")
			}

			if y < c.m-1 {
				if c.crossedEdges[vertId][vertId+c.n] {
					bottomRow += fmt.Sprintf("%s--%s+", TERM_RED, TERM_RESET)
				} else {
					bottomRow += "  +"
				}
			}
		}
		if y < c.m-1 {
			fmt.Println(bottomRow)
			bottomRow = "+"
		}
	}
	c.PrintBoardSeparator()
	c.ClearEdges()
}

func (c Camera) Start() {
	ticker := time.NewTicker(cameraTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.PrintBoard()
		case msg, ok := <-c.cameraChanel:
			switch msg.messageType {
			case CamExplorerSpawned:
				c.board[msg.y][msg.x] = Explorer{id: msg.expId}
			case CamExplorerMoved:
				c.board[msg.y][msg.x] = Explorer{}
				c.board[msg.yHelper][msg.xHelper] = Explorer{id: msg.expId}

				fromId := msg.y*c.n + msg.x
				toId := msg.yHelper*c.n + msg.xHelper
				c.crossedEdges[fromId][toId] = true
				c.crossedEdges[toId][fromId] = true
			}

			if !ok {
				return
			}
		}
	}

}

func (c Camera) ClearEdges() {
	for i := 0; i < c.n*c.m; i++ {
		for j := 0; j < c.n*c.m; j++ {
			c.crossedEdges[i][j] = false
		}
	}
}

func (c Camera) PrintBoardSeparator() {
	fmt.Print("+")
	for i := 0; i < c.n; i++ {
		fmt.Print("--+")
	}
	fmt.Println()
}

func NewCamera(cameraChanel <-chan CameraMessage, n, m int) Camera {
	board := make([][]Explorer, n)
	for y := 0; y < m; y++ {
		board[y] = make([]Explorer, n)
		for x := 0; x < n; x++ {
			board[y][x] = Explorer{}
		}
	}

	numbVert := n * m
	crossedEdges := make([][]bool, numbVert)
	for i := 0; i < numbVert; i++ {
		crossedEdges[i] = make([]bool, numbVert)
		for j := 0; j < numbVert; j++ {
			crossedEdges[i][j] = false
		}
	}

	return Camera{cameraChanel: cameraChanel, board: board, n: n, m: m, crossedEdges: crossedEdges}
}
