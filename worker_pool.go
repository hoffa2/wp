package wp

import (
	"io"
	"sync"
)

// Pool Defines the interface of a worker pool
type Pool struct {
	nWorkers    int
	errCallback bool
	errPrint    io.Writer
	jobfunc     func(in interface{}) error
	jobs        chan interface{}
	sync.WaitGroup
	WorkerPool chan chan interface{}
	quit       chan bool
}

type worker struct {
	jobfunc    func(in interface{}) error
	WorkerPool chan chan interface{}
	jobChannel chan interface{}
	*sync.WaitGroup
	errPrint io.Writer
}

func newWorker(WorkerPool chan chan interface{},
	jobfunc func(in interface{}) error, s *sync.WaitGroup, w io.Writer) *worker {
	return &worker{
		jobfunc:    jobfunc,
		WaitGroup:  s,
		errPrint:   w,
		WorkerPool: WorkerPool,
		jobChannel: make(chan interface{}),
	}
}

// SetErrOutput set to which errors are written

func (w *worker) Start() {
	go func() {
		w.WorkerPool <- w.jobChannel
		for {
			select {
			case job, open := <-w.jobChannel:
				if !open {
					return
				}
				err := w.jobfunc(job)
				if err != nil && w.errPrint != nil {
					w.errPrint.Write([]byte(err.Error()))
				}
				w.WaitGroup.Done()
				w.WorkerPool <- w.jobChannel
			}
		}
	}()
}

// NewPool Creates a new Worker pool - numWorkers is the pool size
// and jobfunc is a callback to function that is executed
func NewPool(numWorkers int, w io.Writer, jobfunc func(in interface{}) error) *Pool {
	return &Pool{
		nWorkers:   numWorkers,
		jobfunc:    jobfunc,
		errPrint:   w,
		jobs:       make(chan interface{}, numWorkers),
		WorkerPool: make(chan chan interface{}, numWorkers),
	}
}

// Start starts the worker pool
// Usage go pool.Start()
func (p *Pool) Start() {
	for i := 0; i < p.nWorkers; i++ {
		w := newWorker(p.WorkerPool, p.jobfunc, &p.WaitGroup, p.errPrint)
		w.Start()
	}

	go func() {
		for {
			select {
			case job := <-p.jobs:
				go func(job interface{}) {
					jobChan := <-p.WorkerPool
					jobChan <- job
				}(job)
			case <-p.quit:
				return
			}
		}
	}()
}

// Wait blocks until there are no more
// busy workers
func (p *Pool) Wait() {
	p.WaitGroup.Wait()
}

// Add queue a job to the worker pool
// with your argument struct
func (p *Pool) Add(job interface{}) {
	p.WaitGroup.Add(1)
	go func() {
		p.jobs <- job
	}()
}

// Quit shuts the pool down after
// all jobs are done
func (p *Pool) Quit() {
	go func() {
		p.quit <- true
		for {
			select {
			case wp := <-p.WorkerPool:
				close(wp)
			default:
				close(p.WorkerPool)
				return
			}
		}
	}()
}
