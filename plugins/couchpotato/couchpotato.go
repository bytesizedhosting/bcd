package couchpotato

import (
	"crypto/md5"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
	"os"
)

type Couchpotato struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Couchpotato, error) {
	manifest, err := plugins.LoadManifest("couchpotato")

	if err != nil {
		return nil, err
	}

	return &Couchpotato{Base: plugins.Base{DockerClient: client, Name: "couchpotato", Version: 1, Manifest: manifest}, imageName: "bytesized/couchpotato"}, nil
}

func (self *Couchpotato) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&CouchpotatoRPC{base: self, BaseRPC: *rpc})
}

type CouchpotatoOpts struct {
	plugins.BaseOpts
	EncPassword string `json:"encrypted_password,omitempty"`
}

func (self *CouchpotatoOpts) hashPassword() string {
	sha := md5.New()
	sha.Write([]byte(self.Password))
	self.EncPassword = fmt.Sprintf("%x", sha.Sum(nil))
	return self.EncPassword
}

func (self *Couchpotato) Install(opts *CouchpotatoOpts) error {
	var err error

	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	opts.hashPassword()

	log.WithFields(log.Fields{
		"plugin":       self.Name,
		"datafolder":   opts.DataFolder,
		"configfolder": opts.ConfigFolder,
		"media_folder": opts.MediaFolder,
		"username":     opts.Username,
	}).Debug("Couchpotato options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/couchpotato/data/settings.conf", opts.ConfigFolder+"/settings.conf", opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"5050/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_couchpotato_" + opts.WebPort})

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
