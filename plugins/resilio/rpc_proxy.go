// THIS FILE IS AUTO-GENERATED DO NOT EDIT

package resilio

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/jobs"
	"github.com/bytesizedhosting/bcd/plugins"
)

type ResilioRPC struct {
	base *Resilio
	plugins.BaseRPC
}
func (self *ResilioRPC) Reinstall(opts *ResilioOpts, job *jobs.Job) error {
	err := self.base.Uninstall(&plugins.AppConfig{ContainerId: opts.ContainerId})
	if err != nil {
		log.Infoln("Could not remove Docker container but since this is a reinstall we don't care.")
	}
	self.Install(opts, job)
	return nil
}

func (self *ResilioRPC) Install(opts *ResilioOpts, job *jobs.Job) error {
	*job = *jobs.New(opts)
	log.Debugln("Resilio options:", opts)
	go func() {
		err := self.base.Install(opts)
		job.Options = *opts

		if err != nil {
			log.Debugln("Resilio installation received an error:", err)
			job.ErrorString = err.Error()
			job.Error = err
			job.Status = jobs.FAILED
		} else {
			log.Infoln("Resilio installation completed")
			job.Status = jobs.FINISHED
		}

		jobs.Storage.Set(job.Id, job)
	}()

	return nil
}
