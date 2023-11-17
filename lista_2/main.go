package main

import (
	"math/rand"
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

func runner(v Vertex, explorerCount *atomic.Uint64, quit *atomic.Bool, maxExplorers int, logChanel chan<- LogPayload) {
	logger := v.CreateLogger(logChanel)

	for !quit.Load() {
		spawnTimer := time.NewTimer(spawnExplorerTick)
		exploreTimer := time.NewTimer(moveExplorerTick)

		if v.explorer == nil {
			// we don't currently have an explorer, so we can either spawn one or accept one from a neighbor
			select {
			case e := <-v.self:
				v.explorer = e
				logger.LogExplorerReceived(v.explorer.id)
			case <-spawnTimer.C:
				if rand.Float64() < spawnExplorerRate && explorerCount.Load() < uint64(maxExplorers-1) {
					id := explorerCount.Add(1)
					v.explorer = &Explorer{id: int(id)}
					logger.LogExplorerSpawned(v.explorer.id)
				}
			}
		} else {
			// we have an explorer, so we can try to move it to a neighbor
			<-exploreTimer.C

			if rand.Float64() < moveExplorerRate {
				// try to move the explorer to a neighbor
				select {
				case v.north <- v.explorer:
					logger.LogExplorerSend(v.explorer.id, North)
					v.explorer = nil
				case v.south <- v.explorer:
					logger.LogExplorerSend(v.explorer.id, South)
					v.explorer = nil
				case v.east <- v.explorer:
					logger.LogExplorerSend(v.explorer.id, East)
					v.explorer = nil
				case v.west <- v.explorer:
					logger.LogExplorerSend(v.explorer.id, West)
					v.explorer = nil
				default:
					// no neighbor is available, so we just keep the explorer
				}
			}
		}

		spawnTimer.Stop()
		exploreTimer.Stop()
	}
}

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

	vertices := CreateLattice(n, m)
	logChannel := make(chan LogPayload, logBuffer)
	loggerDone := make(chan bool)
	cameraChanel := make(chan CameraMessage, cameraBuffer)
	cameraDone := make(chan bool)

	wg := sync.WaitGroup{}
	wg.Add(n * m)

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			go func(v Vertex) {
				runner(v, &explorerCount, &quit, maxExplorers, logChannel)
				wg.Done()
			}(vertices[i][j])
		}
	}

	go func() {
		loggerRun(logChannel, cameraChanel)
		loggerDone <- true
	}()

	go func() {
		camera := NewCamera(cameraChanel, n, m)
		camera.Start()
		cameraDone <- true
	}()

	time.Sleep(runTime)
	quit.Store(true)
	wg.Wait()

	close(logChannel)

	<-loggerDone
	<-cameraDone
}
