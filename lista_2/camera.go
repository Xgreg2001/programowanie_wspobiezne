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
	cameraChannel <-chan CameraMessage
	board         [][]string
	n             int
	m             int
	crossedEdges  [][]bool
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
	CamHazardSpawned
	CamHazardRemoved
	CamExplorerRemoved
)

func RecordSpawnExplorer(expId, x, y int) CameraMessage {
	return CameraMessage{expId: expId, x: x, y: y, messageType: CamExplorerSpawned}
}

func RecordMoveExplorer(expId, fromX, fromY, toX, toY int) CameraMessage {
	return CameraMessage{expId: expId, x: fromX, y: fromY, xHelper: toX, yHelper: toY, messageType: CamExplorerMoved}
}

func RecordSpawnHazard(x, y int) CameraMessage {
	return CameraMessage{messageType: CamHazardSpawned, x: x, y: y}
}

func RecordRemoveHazard(x, y int) CameraMessage {
	return CameraMessage{messageType: CamHazardRemoved, x: x, y: y}
}

func RecordRemoveExplorer(expId, x, y int) CameraMessage {
	return CameraMessage{messageType: CamExplorerRemoved, x: x, y: y, expId: expId}
}

func (c Camera) PrintBoard() {
	c.PrintBoardSeparator()
	bottomRow := "+"
	for y := 0; y < c.m; y++ {
		fmt.Print("|")
		for x := 0; x < c.n; x++ {
			vertId := y*c.n + x

			if c.board[y][x] != "" {
				fmt.Printf("%2s", c.board[y][x])
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
		case msg, ok := <-c.cameraChannel:
			switch msg.messageType {
			case CamExplorerSpawned:
				c.board[msg.y][msg.x] = fmt.Sprintf("%02d", msg.expId)
			case CamExplorerMoved:
				c.board[msg.y][msg.x] = ""
				c.board[msg.yHelper][msg.xHelper] = fmt.Sprintf("%02d", msg.expId)

				fromId := msg.y*c.n + msg.x
				toId := msg.yHelper*c.n + msg.xHelper
				c.crossedEdges[fromId][toId] = true
				c.crossedEdges[toId][fromId] = true
			case CamHazardSpawned:
				c.board[msg.y][msg.x] = "##"
			case CamHazardRemoved:
				c.board[msg.y][msg.x] = ""
			case CamExplorerRemoved:
				c.board[msg.y][msg.x] = ""
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

func NewCamera(cameraChannel <-chan CameraMessage, n, m int) Camera {
	board := make([][]string, n)
	for y := 0; y < m; y++ {
		board[y] = make([]string, n)
	}

	numbVert := n * m
	crossedEdges := make([][]bool, numbVert)
	for i := 0; i < numbVert; i++ {
		crossedEdges[i] = make([]bool, numbVert)
		for j := 0; j < numbVert; j++ {
			crossedEdges[i][j] = false
		}
	}

	return Camera{cameraChannel: cameraChannel, board: board, n: n, m: m, crossedEdges: crossedEdges}
}
