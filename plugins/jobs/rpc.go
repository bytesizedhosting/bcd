package jobrpc

import (
	"fmt"
	"github.com/bytesizedhosting/bcd/jobs"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
)

type JobRPC struct {
	plugins.Base
}

func New() *JobRPC {
	return &JobRPC{plugins.Base{Name: "jobs", Version: 1}}
}

func (self *JobRPC) RegisterRPC(server *rpc.Server) {
	server.Register(self)
}
func (self *JobRPC) Get(jobId string, job *jobs.Job) error {
	njob := jobs.Storage.Get(jobId)

	if (*njob == jobs.Job{}) {
		return fmt.Errorf("Job status got lost, most likely during a reboot of the daemon")
	} else {
		*job = *njob
	}
	return nil
}
