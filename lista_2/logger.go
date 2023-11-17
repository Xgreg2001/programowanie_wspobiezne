package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type LogType int

type LogDirection int

func (l LogDirection) String() string {
	switch l {
	case None:
		return ""
	case North:
		return "N"
	case South:
		return "S"
	case East:
		return "E"
	case West:
		return "W"
	default:
		return "No such direction"
	}
}

type VertexLogger struct {
	vert      *Vertex
	logChanel chan<- LogPayload
}

func (v *Vertex) CreateLogger(logChanel chan<- LogPayload) VertexLogger {
	return VertexLogger{vert: v, logChanel: logChanel}
}

func (l VertexLogger) LogExplorerSpawned(expId int) {
	l.logChanel <- MakeLogExplorerSpawned(l.vert.id, l.vert.x, l.vert.y, expId)
}

func (l VertexLogger) LogExplorerSend(expId int, direction LogDirection) {
	switch direction {
	case North:
		l.logChanel <- MakeLogExplorerSend(l.vert.id, l.vert.x, l.vert.y, l.vert.x, l.vert.y-1, expId, direction)
	case South:
		l.logChanel <- MakeLogExplorerSend(l.vert.id, l.vert.x, l.vert.y, l.vert.x, l.vert.y+1, expId, direction)
	case East:
		l.logChanel <- MakeLogExplorerSend(l.vert.id, l.vert.x, l.vert.y, l.vert.x+1, l.vert.y, expId, direction)
	case West:
		l.logChanel <- MakeLogExplorerSend(l.vert.id, l.vert.x, l.vert.y, l.vert.x-1, l.vert.y, expId, direction)
	default:
		panic("Can't log explorer send with no direction")
	}
}

func (l VertexLogger) LogExplorerReceived(expId int) {
	l.logChanel <- MakeLogExplorerReceived(l.vert.id, l.vert.x, l.vert.y, expId)
}

type LogPayload struct {
	vertexId  int
	logType   LogType
	direction LogDirection
	fromX     int
	fromY     int
	toX       int
	toY       int
	expId     int
	timestamp time.Time
}

const (
	None LogDirection = iota
	North
	South
	East
	West
)

func (l LogPayload) String() string {
	result := fmt.Sprintf("ID: %2d [%s] ", l.vertexId, l.timestamp.Format(time.StampMicro))
	switch l.logType {
	case ExplorerSpawned:
		result += fmt.Sprintf("E-ID: %2d %12s (%2d,%2d)", l.expId, "spawned at", l.fromY, l.fromX)
	case ExplorerSend:
		result += fmt.Sprintf("E-ID: %2d %12s (%2d,%2d) %2s (%2d,%2d) [%s]", l.expId, "send from", l.fromY, l.fromX, "to", l.toY, l.toX, l.direction)
	case ExplorerReceived:
		result += fmt.Sprintf("E-ID: %2d %12s (%2d,%2d)", l.expId, "recived at", l.toY, l.toX)
	default:
		result += fmt.Sprint("No such log type")
	}
	return result
}

const (
	ExplorerSpawned LogType = iota
	ExplorerSend
	ExplorerReceived
)

func MakeLogExplorerSpawned(vertId, x, y, expId int) LogPayload {
	return LogPayload{logType: ExplorerSpawned, fromX: x, fromY: y, expId: expId, timestamp: time.Now(), direction: None, vertexId: vertId}
}

func MakeLogExplorerSend(vertId, fromX, fromY, toX, toY, expId int, direction LogDirection) LogPayload {
	return LogPayload{logType: ExplorerSend, fromX: fromX, fromY: fromY, toX: toX, toY: toY, expId: expId, timestamp: time.Now(), direction: direction, vertexId: vertId}
}

func MakeLogExplorerReceived(vertId, atX, atY, expId int) LogPayload {
	return LogPayload{logType: ExplorerReceived, toX: atX, toY: atY, expId: expId, timestamp: time.Now(), direction: None, vertexId: vertId}
}

func loggerRun(logChanel <-chan LogPayload, cameraChanel chan<- CameraMessage) {
	f, err := os.Create("log.txt")
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	w := bufio.NewWriter(f)
	defer func(w *bufio.Writer) {
		err := w.Flush()
		if err != nil {
			panic(err)
		}
	}(w)

	for log := range logChanel {
		_, err := w.WriteString(log.String() + "\n")
		if err != nil {
			panic(err)
		}

		switch log.logType {
		case ExplorerSpawned:
			cameraChanel <- RecordSpawnExplorer(log.expId, log.fromX, log.fromY)
		case ExplorerSend:
			cameraChanel <- RecordMoveExplorer(log.expId, log.fromX, log.fromY, log.toX, log.toY)
		}
	}

	close(cameraChanel)
}
