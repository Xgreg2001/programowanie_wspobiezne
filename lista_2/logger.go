package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type LogMessage struct {
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
type LogType int

const (
	ExplorerSpawned LogType = iota
	ExplorerMoved
	ExplorerReceived
	ExplorerLeft
)

type LogDirection int

const (
	None LogDirection = iota
	North
	South
	East
	West
)

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
	logChannel chan<- LogMessage
}

type ExplorerLogger struct {
	logChannel chan<- LogMessage
}

func (v *Vertex) AttachLogger(logChannel chan<- LogMessage) {
	v.logger = &VertexLogger{logChannel: logChannel}
}

func (e *Explorer) AttachLogger(logChannel chan<- LogMessage) {
	e.logger = &ExplorerLogger{logChannel: logChannel}
}

func (v Vertex) LogExplorerSpawned(expId int) {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogExplorerSpawned(v.id, v.x, v.y, expId)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Spawned: ", expId)
	}
}

func (e Explorer) LogExplorerMoved(direction LogDirection) {
	if e.logger != nil {
		switch direction {
		case North:
			e.logger.logChannel <- MakeLogExplorerMoved(e.x, e.y, e.x, e.y-1, e.id, direction)
		case South:
			e.logger.logChannel <- MakeLogExplorerMoved(e.x, e.y, e.x, e.y+1, e.id, direction)
		case East:
			e.logger.logChannel <- MakeLogExplorerMoved(e.x, e.y, e.x+1, e.y, e.id, direction)
		case West:
			e.logger.logChannel <- MakeLogExplorerMoved(e.x, e.y, e.x-1, e.y, e.id, direction)
		default:
			panic("Can't log explorer send with no direction")
		}
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Moved: ", e.id)
	}
}

func (v Vertex) LogExplorerReceived(expId int) {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogExplorerReceived(v.id, v.x, v.y, expId)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Received: ", expId)
	}
}

func (v Vertex) LogExplorerLeft(expId int) {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogExplorerLeft(v.id, v.x, v.y, expId)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Received: ", expId)
	}
}

func (l LogMessage) String() string {
	result := fmt.Sprintf("ID: %2d [%s] ", l.vertexId, l.timestamp.Format(time.StampMicro))
	switch l.logType {
	case ExplorerSpawned:
		result += fmt.Sprintf("E-ID: %2d %12s (%2d,%2d)", l.expId, "spawned at", l.fromY, l.fromX)
	case ExplorerMoved:
		result += fmt.Sprintf("E-ID: %2d %12s (%2d,%2d) %2s (%2d,%2d) [%s]", l.expId, "send from", l.fromY, l.fromX, "to", l.toY, l.toX, l.direction)
	case ExplorerReceived:
		result += fmt.Sprintf("E-ID: %2d %12s (%2d,%2d)", l.expId, "received at", l.toY, l.toX)
	case ExplorerLeft:
		result += fmt.Sprintf("E-ID: %2d %12s (%2d,%2d)", l.expId, "left from", l.fromY, l.fromX)
	default:
		result += fmt.Sprint("No such log type")
	}
	return result
}

func MakeLogExplorerSpawned(vertId, x, y, expId int) LogMessage {
	return LogMessage{logType: ExplorerSpawned, fromX: x, fromY: y, expId: expId, timestamp: time.Now(), direction: None, vertexId: vertId}
}

func MakeLogExplorerMoved(fromX, fromY, toX, toY, expId int, direction LogDirection) LogMessage {
	return LogMessage{logType: ExplorerMoved, fromX: fromX, fromY: fromY, toX: toX, toY: toY, expId: expId, timestamp: time.Now(), direction: direction}
}

func MakeLogExplorerReceived(vertId, atX, atY, expId int) LogMessage {
	return LogMessage{logType: ExplorerReceived, toX: atX, toY: atY, expId: expId, timestamp: time.Now(), direction: None, vertexId: vertId}
}

func MakeLogExplorerLeft(vertId, fromX, fromY, expId int) LogMessage {
	return LogMessage{logType: ExplorerLeft, fromX: fromX, fromY: fromY, expId: expId, timestamp: time.Now(), direction: None, vertexId: vertId}
}

func loggerRun(logChanel <-chan LogMessage, cameraChanel chan<- CameraMessage) {
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
		case ExplorerMoved:
			cameraChanel <- RecordMoveExplorer(log.expId, log.fromX, log.fromY, log.toX, log.toY)
		}
	}

	close(cameraChanel)
}
