package sickrage

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
	"os"
)

type Sickrage struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Sickrage, error) {
	manifest, err := plugins.LoadManifest("sickrage")

	if err != nil {
		return nil, err
	}

	return &Sickrage{Base: plugins.Base{DockerClient: client, Name: "sickrage", Version: 1, Manifest: manifest}, imageName: "bytesized/sickrage"}, nil
}

func (self *Sickrage) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&SickrageRPC{base: self, BaseRPC: *rpc})
}

type SickrageOpts struct {
	plugins.BaseOpts
	DelugeWebUrl   string `json:"deluge_web_url,omitempty"`
	DelugePassword string `json:"deluge_password,omitempty"`
}

func (self *Sickrage) Install(opts *SickrageOpts) error {
	var err error

	err = opts.SetDefault("rtorrent")
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":        self.Name,
		"datafolder":    opts.DataFolder,
		"configfolder":  opts.ConfigFolder,
		"media_folder":  opts.MediaFolder,
		"username":      opts.Username,
		"deluge_pass":   opts.DelugePassword,
		"deluge_weburl": opts.DelugeWebUrl,
	}).Debug("Sickrage options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/sickrage/data/config.ini", opts.ConfigFolder+"/config.ini", opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"8081/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_sickrage_" + opts.WebPort})

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
