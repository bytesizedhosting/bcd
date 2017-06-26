package radarr

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
	"os"
)

type Radarr struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Radarr, error) {
	manifest, err := plugins.LoadManifest("radarr")

	if err != nil {
		return nil, err
	}

	return &Radarr{Base: plugins.Base{DockerClient: client, Name: "radarr", Version: 1, Manifest: manifest}, imageName: "bytesized/radarr"}, nil
}

func (self *Radarr) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&RadarrRPC{base: self, BaseRPC: *rpc})
}

type RadarrOpts struct {
	plugins.BaseOpts
}

func (self *Radarr) Install(opts *RadarrOpts) error {
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
		"username":     opts.Username,
	}).Debug("Radarr options")

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
		"7878/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_radarr_" + opts.WebPort})

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
