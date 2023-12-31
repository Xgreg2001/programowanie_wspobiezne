package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// TODO: Make error messages more meaningful
// TODO: Split constants to correct files
const (
	tickTime             = 50 * time.Millisecond
	spawnExplorerRate    = 0.05
	moveExplorerRate     = 0.10
	spawnHazardRate      = 0.05
	hazardLifeTime       = 10 * tickTime
	spawnWildLocatorRate = 0.05
	WildLocatorLifeTime  = 10 * tickTime
	logBuffer            = 100
	runTime              = 5 * time.Second
	cameraTick           = 100 * time.Millisecond
	cameraBuffer         = 100
)

type ExplorerStats struct {
	count  int
	nextId int
	mu     sync.Mutex
}

var shouldQuit atomic.Bool = atomic.Bool{}

func main() {
	explorerStats := ExplorerStats{count: 0, nextId: 1}

	n := 10
	m := 10
	err := error(nil)
	args := os.Args[1:]
	if len(args) == 1 {
		n, err = strconv.Atoi(args[0])
		if err != nil {
			panic(err)
		}
		m = n
	} else if len(args) == 2 {
		n, err = strconv.Atoi(args[0])
		if err != nil {
			panic(err)
		}
		m, err = strconv.Atoi(args[1])
		if err != nil {
			panic(err)
		}
	} else if len(args) > 2 {
		panic("Too many arguments")
	}

	maxExplorers := n * m

	lattice := CreateLattice(n, m)
	logChannel := make(chan LogMessage, logBuffer)
	loggerDone := make(chan bool)
	cameraChanel := make(chan CameraMessage, cameraBuffer)
	cameraDone := make(chan bool)

	go func() {
		loggerRun(logChannel, cameraChanel)
		loggerDone <- true
	}()

	go func() {
		camera := NewCamera(cameraChanel, n, m)
		camera.Start()
		cameraDone <- true
	}()

	vertexWg := sync.WaitGroup{}
	vertexWg.Add(n * m)

	explorerWg := sync.WaitGroup{}
	wildLocatorWg := sync.WaitGroup{}

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			go func(v Vertex) {
				v.run(&explorerWg, &explorerStats, &wildLocatorWg, maxExplorers, logChannel, &lattice)
				vertexWg.Done()
			}(lattice.vertices[i][j])
		}
	}

	time.Sleep(runTime)

	fmt.Println("INFO: starting the exit sequence")

	shouldQuit.Store(true)

	vertexWg.Wait()
	fmt.Println("INFO: all vertex routines finished")

	explorerWg.Wait()
	fmt.Println("INFO: all explorer routines finished")

	wildLocatorWg.Wait()
	fmt.Println("INFO: all wild locator routines finished")

	close(logChannel)

	<-loggerDone
	fmt.Println("INFO: logger routine finished")

	<-cameraDone
	fmt.Println("INFO: camera routine finished")
}
