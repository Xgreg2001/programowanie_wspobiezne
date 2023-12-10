package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	n := 5  // writers
	m := 10 // readers

	monitor := NewMonitor()
	attendance := []string{}
	att_mu := sync.Mutex{}

	done := atomic.Bool{}
	wg := sync.WaitGroup{}

	wg.Add(n + m)

	for i := 0; i < n; i++ {

		go func(id int) {

			run_writer(&done, monitor, &attendance, &att_mu, id)
			wg.Done()
		}(i)
	}

	for i := 0; i < m; i++ {

		go func(id int) {

			run_reader(&done, monitor, &attendance, &att_mu, id)
			wg.Done()
		}(i)
	}

	time.Sleep(30 * time.Second)

	done.Store(true)

	wg.Wait()

}

func run_reader(done *atomic.Bool, monitor *ReaderWriterMonitor, attendance *[]string, att_mu *sync.Mutex, id int) {
	for !done.Load() {
		// wait
		time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)

		monitor.StartRead()
		// Eat

		att_mu.Lock()
		elem := fmt.Sprintf("(reader: %d)", id)
		*attendance = append(*attendance, elem)
		fmt.Println(attendance)
		att_mu.Unlock()

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		att_mu.Lock()

		var idx int
		for i, v := range *attendance {
			if v == elem {
				idx = i
				break
			}
		}

		*attendance = append((*attendance)[:idx], (*attendance)[(idx+1):]...)

		fmt.Println(attendance)
		att_mu.Unlock()

		monitor.StopRead()
	}
}

func run_writer(done *atomic.Bool, monitor *ReaderWriterMonitor, attendance *[]string, att_mu *sync.Mutex, id int) {
	for !done.Load() {
		// wait
		time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)

		monitor.StartWrite()
		// Eat

		att_mu.Lock()
		elem := fmt.Sprintf("(writer: %d)", id)
		*attendance = append(*attendance, elem)
		fmt.Println(attendance)
		att_mu.Unlock()

		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)

		att_mu.Lock()

		var idx int
		for i, v := range *attendance {
			if v == elem {
				idx = i
				break
			}
		}

		*attendance = append((*attendance)[:idx], (*attendance)[(idx+1):]...)

		fmt.Println(attendance)
		att_mu.Unlock()

		monitor.StopWrite()
	}
}
