package jackett

import (
	"crypto/sha512"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
	"path"
)

type Jackett struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Jackett, error) {
	manifest, err := plugins.LoadManifest("jackett")

	if err != nil {
		return nil, err
	}

	return &Jackett{Base: plugins.Base{DockerClient: client, Name: "jackett", Version: 1, Manifest: manifest}, imageName: "bytesized/jackett"}, nil
}

func (self *Jackett) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&JackettRPC{base: self, BaseRPC: *rpc})
}

type JackettOpts struct {
	plugins.BaseOpts
	ApiKey      string `json:"api_key,omitempty"`
	EncPassword string `json:"encrypted_password,omitempty"`
}

func (opts *JackettOpts) hashPassword() string {
	opts.ApiKey = fmt.Sprintf("%x", core.GetRandom(16))
	log.Debugln("Apikey:", opts.ApiKey)
	passString := opts.Password + opts.ApiKey

	// Jackett uses UTF-16 little endian for the sha encryption, so let's fake that.
	ss := []rune{}
	for _, s := range passString {
		ss = append(ss, s, 0)
	}

	sha := sha512.New()
	sha.Write([]byte(string(ss)))
	opts.EncPassword = fmt.Sprintf("%x", sha.Sum(nil))
	return opts.EncPassword
}

func (self *Jackett) Install(opts *JackettOpts) error {
	var err error

	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}
	opts.hashPassword()

	log.WithFields(log.Fields{
		"plugin":        self.Name,
		"config_folder": opts.ConfigFolder,
		"username":      opts.Username,
		"password":      opts.Password,
		"enc_password":  opts.EncPassword,
	}).Debug("Plugin options")

	err = core.EnsurePath(path.Join(opts.ConfigFolder, "Jackett"))
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/jackett/data/ServerConfig.json", path.Join(opts.ConfigFolder, "Jackett", "ServerConfig.json"), opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"9117/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_jackett_" + opts.WebPort})

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
