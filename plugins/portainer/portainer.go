package portainer

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
)

type Portainer struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Portainer, error) {
	manifest, err := plugins.LoadManifest("portainer")

	if err != nil {
		return nil, err
	}

	return &Portainer{Base: plugins.Base{DockerClient: client, Name: "portainer", Version: 1, Manifest: manifest}, imageName: "portainer/portainer:latest"}, nil
}

func (self *Portainer) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&PortainerRPC{base: self, BaseRPC: *rpc})
}

type PortainerOpts struct {
	plugins.BaseOpts
	SocketPath string `json:"socket_path,omitempty"`
}

func (self *Portainer) Install(opts *PortainerOpts) error {
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

	portBindings := map[docker.Port][]docker.PortBinding{
		"9000/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        []string{opts.SocketPath + ":/var/run/docker.sock", opts.ConfigFolder + ":/data"},
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_portainer_" + opts.WebPort})

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
