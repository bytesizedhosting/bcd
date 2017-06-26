// THIS FILE IS AUTO-GENERATED DO NOT EDIT

package rocketchat

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/jobs"
	"github.com/bytesizedhosting/bcd/plugins"
)

type RocketchatRPC struct {
	base *Rocketchat
	plugins.BaseRPC
}
func (self *RocketchatRPC) Reinstall(opts *RocketchatOpts, job *jobs.Job) error {
	err := self.base.Uninstall(&plugins.AppConfig{ContainerId: opts.ContainerId})
	if err != nil {
		log.Infoln("Could not remove Docker container but since this is a reinstall we don't care.")
	}
	self.Install(opts, job)
	return nil
}

func (self *RocketchatRPC) Install(opts *RocketchatOpts, job *jobs.Job) error {
	*job = *jobs.New(opts)
	log.Debugln("Rocketchat options:", opts)
	go func() {
		err := self.base.Install(opts)
		job.Options = *opts

		if err != nil {
			log.Debugln("Rocketchat installation received an error:", err)
			job.ErrorString = err.Error()
			job.Error = err
			job.Status = jobs.FAILED
		} else {
			log.Infoln("Rocketchat installation completed")
			job.Status = jobs.FINISHED
		}

		jobs.Storage.Set(job.Id, job)
	}()

	return nil
}
