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
	LogMsgExplorerSpawned LogType = iota
	LogMsgExplorerMoved
	LogMsgExplorerReceived
	LogMsgExplorerLeft
	LogMsgHazardSpawned
	LogMsgHazardDisappeared
	LogMsgExplorerDied
	LogMsgExplorerEnteredHazard
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
		v.logger.logChannel <- MakeLogMsgExplorerSpawned(v.id, v.x, v.y, expId)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Spawned: ", expId)
	}
}

func (e Explorer) LogExplorerMoved(direction LogDirection) {
	if e.logger != nil {
		switch direction {
		case North:
			e.logger.logChannel <- MakeLogMsgExplorerMoved(e.x, e.y, e.x, e.y-1, e.id, direction)
		case South:
			e.logger.logChannel <- MakeLogMsgExplorerMoved(e.x, e.y, e.x, e.y+1, e.id, direction)
		case East:
			e.logger.logChannel <- MakeLogMsgExplorerMoved(e.x, e.y, e.x+1, e.y, e.id, direction)
		case West:
			e.logger.logChannel <- MakeLogMsgExplorerMoved(e.x, e.y, e.x-1, e.y, e.id, direction)
		default:
			panic("Can't log explorer send with no direction")
		}
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Moved: ", e.id)
	}
}

func (e Explorer) LogExplorerDied() {
	if e.logger != nil {
		e.logger.logChannel <- MakeLogMsgExplorerDied(e.id, e.x, e.y)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Died: ", e.id)
	}
}

func (v Vertex) LogExplorerReceived(expId int) {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogMsgExplorerReceived(v.id, v.x, v.y, expId)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Received: ", expId)
	}
}

func (v Vertex) LogExplorerLeft(expId int) {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogMsgExplorerLeft(v.id, v.x, v.y, expId)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Received: ", expId)
	}
}

func (v Vertex) LogMsgExplorerEnteredHazard(expId int) {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogMsgExplorerEnteredHazard(v.id, v.x, v.y, expId)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on explorer Entered Hazard: ", expId)
	}
}

func (v Vertex) LogHazardSpawned() {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogMsgHazardSpawned(v.id, v.x, v.y)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on Hazard Spawned")
	}
}
func (v Vertex) LogHazardDisappeared() {
	if v.logger != nil {
		v.logger.logChannel <- MakeLogMsgHazardDisappeared(v.id, v.x, v.y)
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: no logger attached on Hazard Disapeard")
	}
}
func (l LogMessage) String() string {
	result := fmt.Sprintf("ID: %2d [%s] ", l.vertexId, l.timestamp.Format(time.StampMicro))
	switch l.logType {
	case LogMsgExplorerSpawned:
		result += fmt.Sprintf("E-ID: %2d %15s (%2d,%2d)", l.expId, "spawned at", l.fromY, l.fromX)
	case LogMsgExplorerMoved:
		result += fmt.Sprintf("E-ID: %2d %15s (%2d,%2d) %2s (%2d,%2d) [%s]", l.expId, "send from", l.fromY, l.fromX, "to", l.toY, l.toX, l.direction)
	case LogMsgExplorerReceived:
		result += fmt.Sprintf("E-ID: %2d %15s (%2d,%2d)", l.expId, "received at", l.toY, l.toX)
	case LogMsgExplorerLeft:
		result += fmt.Sprintf("E-ID: %2d %15s (%2d,%2d)", l.expId, "left from", l.fromY, l.fromX)
	case LogMsgHazardSpawned:
		result += fmt.Sprintf("HAZARD:  %15s (%2d,%2d)", "spawned at", l.toY, l.toX)
	case LogMsgHazardDisappeared:
		result += fmt.Sprintf("HAZARD:  %15s (%2d,%2d)", "disappeared at", l.toY, l.toX)
	case LogMsgExplorerEnteredHazard:
		result += fmt.Sprintf("E-ID: %2d %15s (%2d,%2d)", l.expId, "entered hazard", l.toY, l.toX)
	case LogMsgExplorerDied:
		result += fmt.Sprintf("E-ID: %2d %15s", l.expId, "died")
	default:
		result += fmt.Sprint("No such log type")
	}
	return result
}

func MakeLogMsgBlueprint() LogMessage {
	return LogMessage{timestamp: time.Now(), direction: None}
}

func MakeLogMsgExplorerSpawned(vertId, x, y, expId int) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgExplorerSpawned
	msg.fromX = x
	msg.fromY = y
	msg.expId = expId
	msg.vertexId = vertId
	return msg
}

func MakeLogMsgExplorerMoved(fromX, fromY, toX, toY, expId int, direction LogDirection) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgExplorerMoved
	msg.direction = direction
	msg.fromX = fromX
	msg.fromY = fromY
	msg.toX = toX
	msg.toY = toY
	msg.expId = expId
	return msg
}

func MakeLogMsgExplorerReceived(vertId, atX, atY, expId int) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgExplorerReceived
	msg.toX = atX
	msg.toY = atY
	msg.expId = expId
	msg.vertexId = vertId
	return msg
}

func MakeLogMsgExplorerLeft(vertId, fromX, fromY, expId int) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgExplorerLeft
	msg.fromX = fromX
	msg.fromY = fromY
	msg.vertexId = vertId
	msg.expId = expId
	return msg
}

func MakeLogMsgHazardSpawned(vertId, atX, atY int) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgHazardSpawned
	msg.toX = atX
	msg.toY = atY
	msg.vertexId = vertId
	return msg
}

func MakeLogMsgHazardDisappeared(vertId, atX, atY int) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgHazardDisappeared
	msg.toX = atX
	msg.toY = atY
	msg.vertexId = vertId
	return msg
}

func MakeLogMsgExplorerEnteredHazard(vertId, atX, atY, expId int) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgExplorerEnteredHazard
	msg.toX = atX
	msg.toY = atY
	msg.vertexId = vertId
	msg.expId = expId
	return msg
}

func MakeLogMsgExplorerDied(expId, atX, atY int) LogMessage {
	msg := MakeLogMsgBlueprint()
	msg.logType = LogMsgExplorerDied
	msg.toX = atX
	msg.toY = atY
	msg.expId = expId
	return msg
}

func loggerRun(logChanel <-chan LogMessage, cameraChannel chan<- CameraMessage) {
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
		case LogMsgExplorerSpawned:
			cameraChannel <- RecordSpawnExplorer(log.expId, log.fromX, log.fromY)
		case LogMsgExplorerMoved:
			cameraChannel <- RecordMoveExplorer(log.expId, log.fromX, log.fromY, log.toX, log.toY)
		case LogMsgHazardSpawned:
			cameraChannel <- RecordSpawnHazard(log.toX, log.toY)
		case LogMsgHazardDisappeared:
			cameraChannel <- RecordRemoveHazard(log.toX, log.toY)
		case LogMsgExplorerEnteredHazard:
			cameraChannel <- RecordRemoveHazard(log.toX, log.toY)
		case LogMsgExplorerDied:
			cameraChannel <- RecordRemoveExplorer(log.expId, log.toX, log.toY)
		}

	}

	close(cameraChannel)
}
