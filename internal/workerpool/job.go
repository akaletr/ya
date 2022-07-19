package workerpool

import (
	"cmd/shortener/main.go/internal/model"
	"fmt"
)

type Job struct {
	Err  error
	Data model.Note
	f    func(note model.Note) error
}

func NewJob(f func(note model.Note) error, data model.Note) *Job {
	return &Job{f: f, Data: data}
}

func process(workerID int, job *Job) {
	fmt.Printf("Worker %d processes job %v\n", workerID, job.Data)
	job.Err = job.f(job.Data)
}
