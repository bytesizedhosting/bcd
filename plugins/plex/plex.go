package plex

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
)

type Plex struct {
	plugins.Base
	imageName string
}

type PlexOpts struct {
	plugins.BaseOpts
	PlexEmail    string `json:"plex_email,omitempty"`
	PlexPassword string `json:"plex_password,omitempty"`
	PlexPass     string `json:"plex_pass"`
}

func New(client *docker.Client) (*Plex, error) {
	manifest, err := plugins.LoadManifest("plex")
	if err != nil {
		return nil, err
	}

	return &Plex{Base: plugins.Base{DockerClient: client, Name: "Plex", Version: 1, Manifest: manifest}, imageName: "bytesized/plex"}, nil
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
		dockerImage = dockerImage + ":pass"
	}

	log.WithFields(log.Fields{
		"plugin":       self.Name,
		"datafolder":   opts.DataFolder,
		"configfolder": opts.ConfigFolder,
		"username":     opts.Username,
		"password":     opts.PlexPassword,
		"plexpass":     opts.PlexPass,
	}).Debug("Plex options")

	log.Debugln("Pulling docker image:", dockerImage)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: dockerImage}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"32400/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	log.Infoln("Creating docker container")
	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid, "GUID=" + opts.User.Gid, "PLEX_USERNAME=" + opts.PlexEmail, "PLEX_PASSWORD=" + opts.PlexPassword, "PLEX_EXTERNALPORT=" + opts.WebPort, "RUN_AS_ROOT=FALSE"}, Image: dockerImage}
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
