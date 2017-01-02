package rclone

import (
        log "github.com/Sirupsen/logrus"
        "github.com/fsouza/go-dockerclient"
        "github.com/bytesizedhosting/bcd/plugins"
        "net/rpc"
        "os"
)

type Rclone struct {
        plugins.Base
        imageName string
}

func New(client *docker.Client) (*Rclone, error) {
        manifest, err := plugins.LoadManifest("Rclone")

        if err != nil {
                return nil, err
        }

        return &Rclone{Base: plugins.Base{DockerClient: client, Name: "Rclone", Version: 1, Manifest: manifest}, imageName: "bytesized/rclone"}, nil
}

func (self *Rclone) RegisterRPC(server *rpc.Server) {
        rpc := plugins.NewBaseRPC(self)
        server.Register(&RcloneRPC{base: self, BaseRPC: *rpc})
}

type RcloneOpts struct {
        plugins.BaseOpts
}

func (self *Rclone) Install(opts *RcloneOpts) error {
        var err error

        err = opts.SetDefault(self.Name)
        if err != nil {
                return err
        }

        log.WithFields(log.Fields{
                "plugin":       self.Name,
                "datafolder":   opts.DataFolder,
                "configfolder": opts.ConfigFolder,
                "media_folder": opts.MediaFolder,
        }).Debug("Rclone options")

        err = os.MkdirAll(opts.ConfigFolder, 0755)
        if err != nil {
                return err
        }

        log.Debugln("Pulling docker image", self.imageName)
        err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
        if err != nil {
                return err
        }

        portBindings := map[docker.Port][]docker.PortBinding{
                "8989/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
        }

        hostConfig := docker.HostConfig{
                PortBindings: portBindings,
                Binds:        plugins.DefaultBindings(opts),
        }

        conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

        log.Debugln("Creating docker container")
        c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_rclone_" + opts.WebPort})

        if err != nil {
                return err
        }

        log.Debugln("Starting docker container")

        err = self.DockerClient.StartContainer(c.ID, nil)
        if err != nil {
                return err
        }

        opts.ContainerId = c.ID

        return nil
}
