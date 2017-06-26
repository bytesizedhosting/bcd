package plexrequests

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
	"os"
)

type Plexrequests struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Plexrequests, error) {
	manifest, err := plugins.LoadManifest("plexrequests")

	if err != nil {
		return nil, err
	}

	return &Plexrequests{Base: plugins.Base{DockerClient: client, Name: "plexrequests", Version: 1, Manifest: manifest}, imageName: "bytesized/plex-requests"}, nil
}

func (self *Plexrequests) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&PlexrequestsRPC{base: self, BaseRPC: *rpc})
}

type PlexrequestsOpts struct {
	plugins.BaseOpts
}

func (self *Plexrequests) Install(opts *PlexrequestsOpts) error {
	var err error
	err = opts.SetDefault(self.Name)

	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":       self.Name,
		"datafolder":   opts.DataFolder,
		"configfolder": opts.ConfigFolder,
	}).Debug("Plexrequests options")

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
		"3000/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_plexrequests_" + opts.WebPort})

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
