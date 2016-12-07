package znc

import (
	"crypto/sha256"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
	"path"
)

type Znc struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Znc, error) {
	manifest, err := plugins.LoadManifest("znc")

	if err != nil {
		return nil, err
	}

	return &Znc{Base: plugins.Base{DockerClient: client, Name: "znc", Version: 1, Manifest: manifest}, imageName: "bytesized/znc"}, nil
}

func (self *Znc) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&ZncRPC{base: self, BaseRPC: *rpc})
}

type ZncOpts struct {
	plugins.BaseOpts
	EncPassword string `json:"encrypted_password"`
}

func (self *ZncOpts) hashPassword() string {
	sha := sha256.New()
	sha.Write([]byte(self.Password))
	self.EncPassword = fmt.Sprintf("%x", sha.Sum(nil))
	return self.EncPassword
}

func (self *Znc) Install(opts *ZncOpts) error {
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

	opts.hashPassword()

	err = core.EnsurePath(path.Join(opts.ConfigFolder, "/configs/"))
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/znc/data/znc.conf", path.Join(opts.ConfigFolder, "/configs/", "znc.conf"), opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"6868/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_znc_" + opts.WebPort})

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
