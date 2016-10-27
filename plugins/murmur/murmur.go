package murmur

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
)

type Murmur struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Murmur, error) {
	manifest, err := plugins.LoadManifest("murmur")

	if err != nil {
		return nil, err
	}

	return &Murmur{Base: plugins.Base{DockerClient: client, Name: "murmur", Version: 1, Manifest: manifest}, imageName: "bytesized/murmur"}, nil
}

func (self *Murmur) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&MurmurRPC{base: self, BaseRPC: *rpc})
}

type MurmurOpts struct {
	plugins.BaseOpts
}

func (self *Murmur) Install(opts *MurmurOpts) error {
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

	err = self.WriteTemplate("plugins/murmur/data/murmur.ini", opts.ConfigFolder+"/murmur.ini", opts)
	if err != nil {
		return err
	}

	/*
		log.Debugln("Pulling docker image", self.imageName)
		err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
		if err != nil {
			return err
		}*/

	portBindings := map[docker.Port][]docker.PortBinding{
		"64738/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
		"64738/udp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_murmur_" + opts.WebPort})

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
