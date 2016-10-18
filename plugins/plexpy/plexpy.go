package plexpy

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
	"os"
)

type Plexpy struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Plexpy, error) {
	manifest, err := plugins.LoadManifest("plexpy")

	if err != nil {
		return nil, err
	}

	return &Plexpy{Base: plugins.Base{DockerClient: client, Name: "plexpy", Version: 1, Manifest: manifest}, imageName: "bytesized/plexpy"}, nil
}

func (self *Plexpy) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&PlexpyRPC{base: self, BaseRPC: *rpc})
}

type PlexpyOpts struct {
	plugins.BaseOpts
}

func (self *Plexpy) Install(opts *PlexpyOpts) error {
	var err error
	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":       self.Name,
		"datafolder":   opts.DataFolder,
		"configfolder": opts.ConfigFolder,
		"username":     opts.Username,
	}).Debug("Plexpy options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/plexpy/data/config.ini", opts.ConfigFolder+"/config.ini", opts)
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
		Binds: []string{
			opts.ConfigFolder + ":/config",
		},
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_plexpy_" + opts.WebPort})

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
