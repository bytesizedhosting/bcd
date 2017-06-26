package rocketchat

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
	"os"
)

type Rocketchat struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Rocketchat, error) {
	manifest, err := plugins.LoadManifest("rocketchat")

	if err != nil {
		return nil, err
	}

	return &Rocketchat{Base: plugins.Base{DockerClient: client, Name: "Rocketchat", Version: 1, Manifest: manifest}, imageName: "bytesized/rocketchat"}, nil
}

func (self *Rocketchat) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&RocketchatRPC{base: self, BaseRPC: *rpc})
}

type RocketchatOpts struct {
	plugins.BaseOpts
	DatabaseFolder string `json:"database_folder,omitempty"`
	DataFolder     string `json:"data_folder,omitempty"`
}

func (self *Rocketchat) Install(opts *RocketchatOpts) error {
	var err error

	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	if opts.DatabaseFolder == "" {
		opts.DatabaseFolder = opts.User.HomeDir +
			"/config/rocketchat-db"
	}

	if opts.DataFolder == "" {
		opts.DataFolder = opts.User.HomeDir + "/data/"
	}

	log.WithFields(log.Fields{
		"plugin":          self.Name,
		"data_folder":     opts.DataFolder,
		"database_folder": opts.DatabaseFolder,
		"username":        opts.Username,
	}).Debug("Rocketchat options")

	err = os.MkdirAll(opts.DatabaseFolder, 0755)
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
		Binds: []string{opts.DataFolder +
			":/app/uploads",
			opts.DatabaseFolder + ":/database"},
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_rocketchat_" + opts.WebPort})

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
