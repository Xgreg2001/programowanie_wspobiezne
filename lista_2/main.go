package main

import (
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	spawnExplorerTick = 50 * time.Millisecond
	moveExplorerTick  = 50 * time.Millisecond
	spawnExplorerRate = 0.01
	moveExplorerRate  = 0.10
	logBuffer         = 100
	runTime           = 10 * time.Second
	cameraTick        = 300 * time.Millisecond
	cameraBuffer      = 100
)

func main() {
	explorerCount := atomic.Uint64{}
	quit := atomic.Bool{}

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

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			go func(v Vertex) {
				v.run(&explorerWg, &explorerCount, &quit, maxExplorers, logChannel, &lattice)
				vertexWg.Done()
			}(lattice.vertices[i][j])
		}
	}

	time.Sleep(runTime)
	quit.Store(true)
	vertexWg.Wait()
	explorerWg.Wait()

	close(logChannel)

	<-loggerDone
	<-cameraDone
}
