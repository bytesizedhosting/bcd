package Rclone

import (
        log "github.com/Sirupsen/logrus"
        "github.com/fsouza/go-dockerclient"
        "os"
)

func NewRcloneRPC(parent appPlugin) *RcloneRPC {
        return &RcloneRPC{parent}
}

type BaseRPC struct {
        base appPlugin
}

type ActionOpts struct {
        ContainerId   string   `json:"container_id"`
        DeleteFolders []string `json:"delete_folders"`
}

func (self *RcloneRPC) Start(opts *RcloneOpts, success *bool) error {
        containerId := opts.ContainerId
        log.WithFields(log.Fields{
                "container_id": containerId,
                "name":         self.base.GetName(),
        }).Info("Starting container")

        err := self.base.Start(&AppConfig{ContainerId: containerId})

        if err != nil {
                return err
        }
        *success = true

        log.WithFields(log.Fields{
                "container_id": containerId,
                "name":         self.base.GetName(),
        }).Info("Container started")

        return nil
}

func (self *RcloneRPC) Status(opts *RcloneOpts, state *docker.State) error {
        containerId := opts.ContainerId
        s, err := self.base.Status(&AppConfig{ContainerId: containerId})
        if err != nil {
                return err
        }

        *state = *s
        return nil
}

func (self *RcloneRPC) Stop(opts *RcloneOpts, success *bool) error {
        containerId := opts.ContainerId
        log.WithFields(log.Fields{
                "container_id": containerId,
                "name":         self.base.GetName(),
        }).Info("Stopping container")

        err := self.base.Stop(&AppConfig{ContainerId: containerId})

        if err != nil {
                return err
        }
        *success = true

        log.WithFields(log.Fields{
                "container_id": containerId,
                "name":         self.base.GetName(),
        }).Info("Container stopped")

        return nil
}
func (self *RcloneRPC) Restart(opts *RcloneOpts, success *bool) error {
        containerId := opts.ContainerId
        err := self.base.Restart(&AppConfig{ContainerId: containerId})

        if err != nil {
                return err
        }
        *success = true

        return nil
}
func (self *RcloneRPC) Uninstall(opts *RcloneOpts, success *bool) error {
        containerId := opts.ContainerId

        log.WithFields(log.Fields{
                "container_id": containerId,
                "name":         self.base.GetName(),
        }).Info("Removing container")

        err := self.base.Uninstall(&AppConfig{ContainerId: containerId})

        if err != nil {
                return err
        }
        *success = true

        if len(opts.DeleteFolders) > 0 {
                for _, folder := range opts.DeleteFolders {
                        log.WithFields(log.Fields{"folder": folder}).Info("Removing folder")
                        err := os.RemoveAll(folder)
                        if err != nil {
                                log.Infof("Could not delete folder '%s': '%s'", folder, err)
                        }
                }
        }

        log.WithFields(log.Fields{
                "container_id": containerId,
                "name":         self.base.GetName(),
        }).Info("Container removed")

        return nil
}
