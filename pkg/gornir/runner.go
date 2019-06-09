package gornir

import (
	"context"
	"sync"
)

type JobParameters struct {
	title  string
	logger Logger
	host   *Host
}

func NewJobParameters(title string, logger Logger) *JobParameters {
	return &JobParameters{
		title:  title,
		logger: logger,
	}
}

func (jp *JobParameters) Title() string {
	return jp.title
}

func (jp *JobParameters) Logger() Logger {
	return jp.logger
}

func (jp *JobParameters) Host() *Host {
	return jp.host
}

func (jp *JobParameters) ForHost(host *Host) *JobParameters {
	return &JobParameters{
		title:  jp.title,
		logger: jp.logger.WithField("host", host.Hostname),
		host:   host,
	}
}

// Task is the interface that task plugins need to implement.
// the task is responsible to indicate its completion
// by calling sync.WaitGroup.Done()
type Task interface {
	Run(context.Context, *Host) (interface{}, error)
}

// Runner is the interface of a struct that can implement a strategy
// to run tasks over hosts
type Runner interface {
	Run(context.Context, Task, map[string]*Host, *JobParameters, chan *JobResult) error // Run executes the task over the hosts
	Close() error                                                                       // Close closes and cleans all objects associated with the runner
	Wait() error                                                                        // Wait blocks until all the hosts are done executing the task
}

// JobResult is the result of running a task over a host.
type JobResult struct {
	ctx  context.Context
	jp   *JobParameters
	err  error
	data interface{}
}

// NewJobResult instantiates a new JobResult
func NewJobResult(ctx context.Context, jp *JobParameters, data interface{}, err error) *JobResult {
	return &JobResult{
		ctx:  ctx,
		jp:   jp,
		data: data,
		err:  err,
	}
}

// Context returns the context associated with the task
func (r *JobResult) Context() context.Context {
	return r.ctx
}

func (r *JobResult) JobParameters() *JobParameters {
	return r.jp
}

// Err returns the error the task set, otherwise nil
func (r *JobResult) Err() error {
	return r.err
}

// SetErr stores the error  and also propagates it to the associated Host
func (r *JobResult) SetErr(err error) {
	r.err = err
	r.JobParameters().Host().setErr(err)
}

// Data retrieves arbitrary data stored in the object
func (r *JobResult) Data() interface{} {
	return r.data
}

// SetData let's you store arbitrary data in the object
func (r *JobResult) SetData(data interface{}) {
	r.data = data
}

func TaskWrapper(ctx context.Context, wg *sync.WaitGroup, taskFunc Task, jp *JobParameters, results chan *JobResult) {
	defer wg.Done()
	res, err := taskFunc.Run(ctx, jp.host)
	jp.Host().setErr(err)
	results <- NewJobResult(ctx, jp, res, err)
}
