package headphones

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
)

type Headphones struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Headphones, error) {
	manifest, err := plugins.LoadManifest("headphones")

	if err != nil {
		return nil, err
	}

	return &Headphones{Base: plugins.Base{DockerClient: client, Name: "headphones", Version: 1, Manifest: manifest}, imageName: "bytesized/headphones"}, nil
}

func (self *Headphones) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&HeadphonesRPC{base: self, BaseRPC: *rpc})
}

type HeadphonesOpts struct {
	plugins.BaseOpts
}

func (self *Headphones) Install(opts *HeadphonesOpts) error {
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

	err = self.WriteTemplate("plugins/headphones/data/config.ini", opts.ConfigFolder+"/config.ini", opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"8181/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_headphones_" + opts.WebPort})

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
