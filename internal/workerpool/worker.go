package workerpool

import (
	"fmt"
	"sync"
)

// Worker контролирует всю работу
type Worker struct {
	ID      int
	jobChan chan *Job
	quit    chan bool
}

// NewWorker возвращает новый экземпляр worker-а
func NewWorker(channel chan *Job, ID int) *Worker {
	return &Worker{
		ID:      ID,
		jobChan: channel,
	}
}

// запуск worker
func (wr *Worker) Start(wg *sync.WaitGroup) {
	fmt.Printf("Starting worker %d\n", wr.ID)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for job := range wr.jobChan {
			process(wr.ID, job)
		}
	}()
}

// StartBackground запускает worker-а в фоне
func (wr *Worker) StartBackground() {
	fmt.Printf("Starting worker %d\n", wr.ID)

	for {
		select {
		case job := <-wr.jobChan:
			process(wr.ID, job)
		case <-wr.quit:
			return
		}
	}
}

// Остановка quits для воркера
func (wr *Worker) Stop() {
	fmt.Printf("Closing worker %d\n", wr.ID)
	go func() {
		wr.quit <- true
	}()
}
