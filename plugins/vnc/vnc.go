package vnc

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
)

type Vnc struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Vnc, error) {
	manifest, err := plugins.LoadManifest("vnc")

	if err != nil {
		return nil, err
	}

	return &Vnc{Base: plugins.Base{DockerClient: client, Name: "vnc", Version: 1, Manifest: manifest}, imageName: "bytesized/vnc"}, nil
}

func (self *Vnc) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&VncRPC{base: self, BaseRPC: *rpc})
}

type VncOpts struct {
	plugins.BaseOpts
	VncPort string `json:"vnc_port"`
}

func (self *Vnc) Install(opts *VncOpts) error {
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
		"6080/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
		"5900/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.VncPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_vnc_" + opts.WebPort})

	if err != nil {
		return err
	}

	log.Debugln("Starting docker container")

	err = self.DockerClient.StartContainer(c.ID, nil)
	if err != nil {
		return err
	}

	exec, err := self.DockerClient.CreateExec(docker.CreateExecOptions{Cmd: []string{"exec", "s6-setuidgid", "bytesized", "/app/set_password", opts.Password}, Container: c.ID})
	if err != nil {
		return err
	}

	err = self.DockerClient.StartExec(exec.ID, docker.StartExecOptions{})
	if err != nil {
		return err
	}

	opts.ContainerId = c.ID

	return nil
}
