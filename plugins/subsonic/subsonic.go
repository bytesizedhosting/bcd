package subsonic

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"time"
)

type Subsonic struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Subsonic, error) {
	manifest, err := plugins.LoadManifest("subsonic")

	if err != nil {
		return nil, err
	}

	return &Subsonic{Base: plugins.Base{DockerClient: client, Name: "subsonic", Version: 1, Manifest: manifest}, imageName: "bytesized/subsonic"}, nil
}

func (self *Subsonic) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&SubsonicRPC{base: self, BaseRPC: *rpc})
}

type SubsonicOpts struct {
	plugins.BaseOpts
}

func (self *Subsonic) Install(opts *SubsonicOpts) error {
	var err error

	err = opts.SetDefault("rtorrent")
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":        self.Name,
		"data_folder":   opts.DataFolder,
		"config_folder": opts.ConfigFolder,
		"username":      opts.Username,
	}).Debug("Subsonic options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}
	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		"4040/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		Binds:        plugins.DefaultBindings(opts),
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_subsonic_" + opts.WebPort})

	if err != nil {
		return err
	}

	log.Debugln("Starting docker container")

	err = self.DockerClient.StartContainer(c.ID, nil)
	if err != nil {
		return err
	}

	opts.ContainerId = c.ID

	log.Debugln("Waiting for Subsonic to boot up to create admin user")
	for i := 0; i < 20; i++ {
		subsonicPath := fmt.Sprintf("http://127.0.0.1:%s/rest/createUser.view", opts.WebPort)
		authOpts := url.Values{"username": {opts.Username}, "password": {opts.Password}, "adminRole": {"true"}, "email": {"test@test.com"}, "u": {"admin"}, "p": {"admin"}, "v": {"1.1.0"}, "c": {"BytesizedConnect"}, "f": {"json"}}
		_, err := http.Get(subsonicPath + "?" + authOpts.Encode())
		if err != nil {
			log.Debugf("Subsonic not up yet: '%s'", err)
		} else {
			subsonicPath = fmt.Sprintf("http://127.0.0.1:%s/rest/changePassword.view", opts.WebPort)
			authOpts = url.Values{"username": {"admin"}, "password": {opts.Password}, "u": {"admin"}, "p": {"admin"}, "v": {"1.1.0"}, "c": {"BytesizedConnect"}, "f": {"json"}}
			_, err = http.Get(subsonicPath + "?" + authOpts.Encode())
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	log.Warnln("Subsonic is installed but no credentials could be set")
	return nil
}
