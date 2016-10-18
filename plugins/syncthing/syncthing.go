package syncthing

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"golang.org/x/crypto/bcrypt"
	"net/rpc"
	"os"
)

type Syncthing struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Syncthing, error) {
	manifest, err := plugins.LoadManifest("syncthing")
	if err != nil {
		return nil, err
	}

	return &Syncthing{Base: plugins.Base{DockerClient: client, Name: "Syncthing", Version: 1, Manifest: manifest}, imageName: "bytesized/syncthing"}, nil
}

func (self *Syncthing) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&SyncthingRPC{base: self, BaseRPC: *rpc})
}

type SyncthingOpts struct {
	plugins.BaseOpts
	EncPassword string `json:"encrypted_password,omitempty"`
}

func (self *SyncthingOpts) hashPassword() string {
	log.Debugln("Encrypting password", self.Password)
	encPassword, err := bcrypt.GenerateFromPassword([]byte(self.Password), 0)
	if err != nil {
		log.Panicln("Could not encrypt password, HALP")
	}
	self.EncPassword = string(encPassword)
	return self.EncPassword
}

func (self *Syncthing) Install(opts *SyncthingOpts) error {
	var err error
	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	var ports []string

	for i := 0; i < 3; i++ {
		p, err := core.GetFreePort()
		if err != nil {
			return err
		}
		ports = append(ports, p)
	}

	opts.hashPassword()

	log.WithFields(log.Fields{
		"plugin":             self.Name,
		"data_folder":        opts.DataFolder,
		"config_folder":      opts.ConfigFolder,
		"username":           opts.Username,
		"password":           opts.Password,
		"encrypted_password": opts.EncPassword,
	}).Debug("Syncthing options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/syncthing/data/config.xml", opts.ConfigFolder+"/config.xml", opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"8384/tcp":  []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
		"22000/tcp": []docker.PortBinding{docker.PortBinding{HostPort: ports[1]}},
		"21025/udp": []docker.PortBinding{docker.PortBinding{HostPort: ports[2]}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Infoln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_syncthing_" + opts.WebPort})

	if err != nil {
		return err
	}

	log.Infoln("Starting docker container")

	err = self.DockerClient.StartContainer(c.ID, nil)
	if err != nil {
		return err
	}

	opts.ContainerId = c.ID

	return nil
}
