// THIS FILE IS AUTO-GENERATED DO NOT EDIT

package headphones

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/jobs"
	"github.com/bytesizedhosting/bcd/plugins"
)

type HeadphonesRPC struct {
	base *Headphones
	plugins.BaseRPC
}
func (self *HeadphonesRPC) Reinstall(opts *HeadphonesOpts, job *jobs.Job) error {
	err := self.base.Uninstall(&plugins.AppConfig{ContainerId: opts.ContainerId})
	if err != nil {
		log.Infoln("Could not remove Docker container but since this is a reinstall we don't care.")
	}
	self.Install(opts, job)
	return nil
}

func (self *HeadphonesRPC) Install(opts *HeadphonesOpts, job *jobs.Job) error {
	*job = *jobs.New(opts)
	log.Debugln("Headphones options:", opts)
	go func() {
		err := self.base.Install(opts)
		job.Options = *opts

		if err != nil {
			log.Debugln("Headphones installation received an error:", err)
			job.ErrorString = err.Error()
			job.Error = err
			job.Status = jobs.FAILED
		} else {
			log.Infoln("Headphones installation completed")
			job.Status = jobs.FINISHED
		}

		jobs.Storage.Set(job.Id, job)
	}()

	return nil
}
