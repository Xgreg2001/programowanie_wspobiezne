package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Fork struct {
	id  int
	sem Semaphore
}

type Philosopher struct {
	id int
}

func main() {

	n := 5

	// room := NewSemaphore(int64(n - 1))
	forks := []Fork{}
	eating := []string{}
	eating_mu := sync.Mutex{}
	monitor := NewMonitor(n)

	done := atomic.Bool{}
	wg := sync.WaitGroup{}

	for i := 0; i < n; i++ {
		forks = append(forks, Fork{id: i, sem: *NewSemaphore(1)})
	}

	wg.Add(n)

	for i := 0; i < n; i++ {
		phil := Philosopher{id: i}

		go func() {

			phil.run_with_monitor(&done, monitor, forks, &eating, &eating_mu, n)
			// phil.run_semaphores(&done, forks, &eating, &eating_mu, n)
			// phil.run_asymetric(&done, forks, &eating, &eating_mu, n)
			// phil.run_with_room(&done, room, forks, &eating, &eating_mu, n)
			wg.Done()

		}()
	}

	time.Sleep(10 * time.Second)

	done.Store(true)

	wg.Wait()

}

func (p Philosopher) run_semaphores(done *atomic.Bool, forks []Fork, eating *[]string, eating_mu *sync.Mutex, n int) {
	for !done.Load() {
		// Think
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		forks[p.id].sem.Wait()
		forks[(p.id+1)%n].sem.Wait()

		// Eat

		eating_mu.Lock()
		elem := fmt.Sprintf("(%d, %d, %d)", forks[p.id].id, p.id, forks[(p.id+1)%n].id)
		*eating = append(*eating, elem)
		fmt.Println(eating)
		eating_mu.Unlock()

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		eating_mu.Lock()

		var idx int
		for i, v := range *eating {
			if v == elem {
				idx = i
				break
			}
		}

		*eating = append((*eating)[:idx], (*eating)[(idx+1):]...)

		fmt.Println(eating)
		eating_mu.Unlock()

		forks[(p.id+1)%n].sem.Signal()
		forks[p.id].sem.Signal()
	}
}

func (p Philosopher) run_asymetric(done *atomic.Bool, forks []Fork, eating *[]string, eating_mu *sync.Mutex, n int) {
	for !done.Load() {
		// Think
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		if p.id == 0 {
			forks[(p.id+1)%n].sem.Wait()
			forks[p.id].sem.Wait()
		} else {
			forks[p.id].sem.Wait()
			forks[(p.id+1)%n].sem.Wait()
		}

		// Eat

		eating_mu.Lock()
		elem := fmt.Sprintf("(%d, %d, %d)", forks[p.id].id, p.id, forks[(p.id+1)%n].id)
		*eating = append(*eating, elem)
		fmt.Println(eating)
		eating_mu.Unlock()

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		eating_mu.Lock()

		var idx int
		for i, v := range *eating {
			if v == elem {
				idx = i
				break
			}
		}

		*eating = append((*eating)[:idx], (*eating)[(idx+1):]...)

		fmt.Println(eating)
		eating_mu.Unlock()

		if p.id == 0 {
			forks[p.id].sem.Signal()
			forks[(p.id+1)%n].sem.Signal()
		} else {
			forks[(p.id+1)%n].sem.Signal()
			forks[p.id].sem.Signal()
		}
	}
}

func (p Philosopher) run_with_room(done *atomic.Bool, room *Semaphore, forks []Fork, eating *[]string, eating_mu *sync.Mutex, n int) {
	for !done.Load() {
		// Think
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		room.Wait()
		forks[p.id].sem.Wait()
		forks[(p.id+1)%n].sem.Wait()

		// Eat

		eating_mu.Lock()
		elem := fmt.Sprintf("(%d, %d, %d)", forks[p.id].id, p.id, forks[(p.id+1)%n].id)
		*eating = append(*eating, elem)
		fmt.Println(eating)
		eating_mu.Unlock()

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		eating_mu.Lock()

		var idx int
		for i, v := range *eating {
			if v == elem {
				idx = i
				break
			}
		}

		*eating = append((*eating)[:idx], (*eating)[(idx+1):]...)

		fmt.Println(eating)
		eating_mu.Unlock()

		forks[(p.id+1)%n].sem.Signal()
		forks[p.id].sem.Signal()
		room.Signal()
	}
}

func (p Philosopher) run_with_monitor(done *atomic.Bool, monitor *PhilMonitor, forks []Fork, eating *[]string, eating_mu *sync.Mutex, n int) {
	for !done.Load() {
		// Think
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		monitor.TakeFork(p.id)

		// Eat

		eating_mu.Lock()
		elem := fmt.Sprintf("(%d, %d, %d)", forks[p.id].id, p.id, forks[(p.id+1)%n].id)
		*eating = append(*eating, elem)
		fmt.Println(eating)
		eating_mu.Unlock()

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		eating_mu.Lock()

		var idx int
		for i, v := range *eating {
			if v == elem {
				idx = i
				break
			}
		}

		*eating = append((*eating)[:idx], (*eating)[(idx+1):]...)

		fmt.Println(eating)
		eating_mu.Unlock()

		monitor.ReleaseFork(p.id)
	}
}
