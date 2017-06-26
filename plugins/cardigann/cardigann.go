package cardigann

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
	"os"
)

type Cardigann struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Cardigann, error) {
	manifest, err := plugins.LoadManifest("cardigann")

	if err != nil {
		return nil, err
	}

	return &Cardigann{Base: plugins.Base{DockerClient: client, Name: "cardigann", Version: 1, Manifest: manifest}, imageName: "bytesized/cardigann"}, nil
}

func (self *Cardigann) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&CardigannRPC{base: self, BaseRPC: *rpc})
}

type CardigannOpts struct {
	plugins.BaseOpts
	DataFolder string `json:"data_folder,omitempty"`
	TvFolder   string `json:"tv_folder,omitempty"`
}

func (self *Cardigann) Install(opts *CardigannOpts) error {
	var err error
	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":       self.Name,
		"datafolder":   opts.DataFolder,
		"configfolder": opts.ConfigFolder,
		"password":     opts.Password,
	}).Debug("Cardigann options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/cardigann/data/config.json", opts.ConfigFolder+"/config.json", opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"5060/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds: []string{
			opts.ConfigFolder + ":/config",
		},
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_cardigann_" + opts.WebPort})

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
