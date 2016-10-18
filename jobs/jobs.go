package jobs

import (
	"fmt"
	"github.com/bytesizedhosting/bcd/core"
	"sync"
)

const (
	FAILED   = -1
	BUSY     = iota
	FINISHED = iota
)

type Job struct {
	Id          string      `json:"job_id,omitempty"`
	Status      int         `json:"status,omitempty"`
	Options     interface{} `json:"options,omitempty"`
	Error       error       `json:"error,omitempty"`
	ErrorString string      `json:"error_message, omitempty"`
}

// This is a very easy memory storage like solution, we might need something more durable at one point.
type JobStorage struct {
	Jobs map[string]Job
}

var Storage JobStorage
var mutex = &sync.Mutex{}

func init() {
	Storage.Jobs = make(map[string]Job)
}

func (self *JobStorage) Set(jobId string, job *Job) error {
	mutex.Lock()
	self.Jobs[jobId] = *job
	mutex.Unlock()

	return nil
}

func (self *JobStorage) Get(jobId string) *Job {
	job := self.Jobs[jobId]
	return &job
}

func New(opts interface{}) *Job {
	id := fmt.Sprintf("%02X", core.GetRandom(8))
	job := Job{Id: id, Status: BUSY, Options: opts}
	Storage.Jobs[job.Id] = job

	return &job
}
