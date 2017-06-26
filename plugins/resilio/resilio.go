package resilio

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
)

type Resilio struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Resilio, error) {
	manifest, err := plugins.LoadManifest("resilio")

	if err != nil {
		return nil, err
	}

	return &Resilio{Base: plugins.Base{DockerClient: client, Name: "resilio", Version: 1, Manifest: manifest}, imageName: "bytesized/resilio-sync"}, nil
}

func (self *Resilio) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&ResilioRPC{base: self, BaseRPC: *rpc})
}

type ResilioOpts struct {
	plugins.BaseOpts
}

func (self *Resilio) Install(opts *ResilioOpts) error {
	var err error

	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":        self.Name,
		"config_folder": opts.ConfigFolder,
		"username":      opts.Username,
	}).Debug("Plugin options")

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/resilio/data/sync.conf", opts.ConfigFolder+"/sync.conf", opts)
	if err != nil {
		return err
	}

	p, err := core.GetFreePort()
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"8888/tcp":  []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
		"55555/tcp": []docker.PortBinding{docker.PortBinding{HostPort: p}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_resilio_" + opts.WebPort})

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
