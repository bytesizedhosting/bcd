package plex

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
)

type Plex struct {
	plugins.Base
	imageName string
}

type PlexOpts struct {
	plugins.BaseOpts
	PlexClaim string `json:"plex_claim,omitempty"`
	PlexPass  string `json:"plex_pass"`
}

func New(client *docker.Client) (*Plex, error) {
	manifest, err := plugins.LoadManifest("plex")
	if err != nil {
		return nil, err
	}

	return &Plex{Base: plugins.Base{DockerClient: client, Name: "Plex", Version: 1, Manifest: manifest}, imageName: "plexinc/pms-docker"}, nil
}

func (self *Plex) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&PlexRPC{base: self, BaseRPC: *rpc})
}

func (self *Plex) Install(opts *PlexOpts) error {
	var err error
	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	dockerImage := self.imageName

	if opts.PlexPass == "1" {
		dockerImage = dockerImage + ":plexpass"
	} else {
		dockerImage = dockerImage + ":latest"
	}

	log.WithFields(log.Fields{
		"plugin":       self.Name,
		"datafolder":   opts.DataFolder,
		"configfolder": opts.ConfigFolder,
		"claim":        opts.PlexClaim,
		"plexpass":     opts.PlexPass,
	}).Debug("Plex options")

	log.Debugln("Pulling docker image:", dockerImage)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: dockerImage}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	hostConfig := docker.HostConfig{
		NetworkMode: "host",
		Binds:       plugins.DefaultBindings(opts),
	}

	log.Infoln("Creating docker container")
	conf := docker.Config{Env: []string{"PLEX_UID=" + opts.User.Uid, "PLEX_GID=" + opts.User.Gid, "PLEX_CLAIM=" + opts.PlexClaim}, Image: dockerImage}
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_plex_" + opts.WebPort})

	if err != nil {
		return err
	}
	opts.ContainerId = c.ID

	log.Infoln("Starting docker container")
	err = self.DockerClient.StartContainer(c.ID, nil)
	if err != nil {
		return err
	}

	return nil
}
