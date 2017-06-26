package nzbget

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
	"os"
)

type Nzbget struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Nzbget, error) {
	manifest, err := plugins.LoadManifest("nzbget")

	if err != nil {
		return nil, err
	}

	return &Nzbget{Base: plugins.Base{DockerClient: client, Name: "nzbget", Version: 1, Manifest: manifest}, imageName: "bytesized/nzbget"}, nil
}

func (self *Nzbget) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&NzbgetRPC{base: self, BaseRPC: *rpc})
}

type NzbgetOpts struct {
	plugins.BaseOpts
}

func (self *Nzbget) Install(opts *NzbgetOpts) error {
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
	}).Debug("Nzbget options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/nzbget/data/nzbget.conf", opts.ConfigFolder+"/nzbget.conf", opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"6789/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_nzbget_" + opts.WebPort})

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
