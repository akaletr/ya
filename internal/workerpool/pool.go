package workerpool

import (
	"sync"
)

// Pool воркера
type Pool struct {
	Jobs    []*Job
	Workers []*Worker

	concurrency   int
	collector     chan *Job
	runBackground chan bool
	wg            sync.WaitGroup
}

// AddJob добавляет job в pool
func (p *Pool) AddJob(job *Job) {
	p.collector <- job
}

// NewPool инициализирует новый пул с заданными задачами и
func NewPool(concurrency int) *Pool {
	return &Pool{
		Jobs:        nil,
		concurrency: concurrency,
		collector:   make(chan *Job, 1000),
	}
}

// Run запускает всю работу в Pool и блокирует ее до тех пор,
// пока она не будет закончена.
//func (p *Pool) Run() {
//	for i := 1; i <= p.concurrency; i++ {
//		worker := NewWorker(p.collector, i)
//		worker.Start(&p.wg)
//	}
//
//	for i := range p.Jobs {
//		p.collector <- p.Jobs[i]
//	}
//	close(p.collector)
//
//	p.wg.Wait()
//}

// RunBackground запускает pool в фоне
func (p *Pool) RunBackground() {
	for i := 1; i <= p.concurrency; i++ {
		worker := NewWorker(p.collector, i)
		p.Workers = append(p.Workers, worker)
		go worker.StartBackground()
	}

	for i := range p.Jobs {
		p.collector <- p.Jobs[i]
	}

	p.runBackground = make(chan bool)
	<-p.runBackground
}

// Stop останавливает запущенных в фоне worker-ов
func (p *Pool) Stop() {
	for i := range p.Workers {
		p.Workers[i].Stop()
	}
	p.runBackground <- true
}
