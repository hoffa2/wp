package util

import "sync"

// Pool Defines the interface of a worker pool
type Pool struct {
	nWorkers int
	jobfunc  func(in interface{})
	jobs     chan interface{}
	sync.WaitGroup
	WorkerPool chan chan interface{}
	quit       chan bool
}

type worker struct {
	jobfunc    func(in interface{})
	WorkerPool chan chan interface{}
	jobChannel chan interface{}
	*sync.WaitGroup
}

func newWorker(WorkerPool chan chan interface{}, jobfunc func(in interface{}), w *sync.WaitGroup) *worker {
	return &worker{
		jobfunc:    jobfunc,
		WaitGroup:  w,
		WorkerPool: WorkerPool,
		jobChannel: make(chan interface{}),
	}
}

func (w *worker) Start() {
	go func() {
		w.WorkerPool <- w.jobChannel
		for {
			select {
			case job, open := <-w.jobChannel:
				if !open {
					return
				}
				w.jobfunc(job)
				w.WaitGroup.Done()
				w.WorkerPool <- w.jobChannel
			}
		}
	}()
}

// NewPool Creates a new Worker pool - numWorkers is the pool size
// and jobfunc is a callback to function that is executed
func NewPool(numWorkers int, jobfunc func(in interface{})) *Pool {
	return &Pool{
		nWorkers:   numWorkers,
		jobfunc:    jobfunc,
		jobs:       make(chan interface{}, numWorkers),
		WorkerPool: make(chan chan interface{}, numWorkers),
	}
}

// Start starts the worker pool
// Usage go pool.Start()
func (p *Pool) Start() {
	for i := 0; i < p.nWorkers; i++ {
		w := newWorker(p.WorkerPool, p.jobfunc, &p.WaitGroup)
		w.Start()
	}

	go func() {
		for {
			select {
			case job := <-p.jobs:
				p.WaitGroup.Add(1)
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
				return
			}
		}
	}()
}
