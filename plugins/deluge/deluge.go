package deluge

import (
	"crypto/sha1"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
	"os"
)

const imageName = "bytesized/deluge"

func New(client *docker.Client) (*Deluge, error) {
	manifest, err := plugins.LoadManifest("deluge")
	if err != nil {
		return nil, err
	}

	return &Deluge{plugins.Base{DockerClient: client, Name: "deluge", Version: 1, Manifest: manifest}}, nil
}

type Deluge struct {
	plugins.Base
}

type DelugeInstallRes struct {
	Message string      `json:"message"`
	Error   bool        `json:"error"`
	Opts    *DelugeOpts `json:"options"`
}

type DelugeOpts struct {
	plugins.BaseOpts
	EncPassword string `json:"encrypted_password,omitempty"`
	Salt        string `json:"salt,omitempty"`
	Password    string `json:"password,omitempty"`
	DaemonPort  string `json:"daemon_port,omitempty"`
}

func (self *Deluge) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&DelugeRPC{base: self, BaseRPC: *rpc})
}

func (self *Deluge) Install(opts *DelugeOpts) error {
	log.Infoln("Starting Deluge installation")
	var err error
	var ports []string

	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	for i := 0; i < 4; i++ {
		p, err := core.GetFreePort()
		if err != nil {
			return err
		}
		ports = append(ports, p)
	}

	opts.Salt = fmt.Sprintf("%x", core.GetRandom(20))

	if opts.DaemonPort == "" {
		opts.DaemonPort = ports[1]
	}

	opts.hashPassword()

	log.WithFields(log.Fields{
		"plugin":       self.Name,
		"datafolder":   opts.DataFolder,
		"configfolder": opts.ConfigFolder,
		"username":     opts.Username,
		"password":     opts.Password,
		"daemonport":   opts.DaemonPort,
		"webport":      opts.WebPort,
	}).Debug("Current Deluge options")

	err = os.MkdirAll(opts.ConfigFolder, 0755)
	if err != nil {
		return err
	}

	// TODO path
	err = self.WriteTemplate("plugins/deluge/data/core.conf", opts.ConfigFolder+"/core.conf", opts)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/deluge/data/web.conf", opts.ConfigFolder+"/web.conf", opts)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/deluge/data/hostlist.conf.1.2", opts.ConfigFolder+"/hostlist.conf.1.2", opts)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/deluge/data/auth", opts.ConfigFolder+"/auth", opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: imageName}

	hostConfig := docker.HostConfig{
		NetworkMode: "host",
		Binds:       plugins.DefaultBindings(opts),
	}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_deluge_" + opts.WebPort})

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

func (self *DelugeOpts) hashPassword() string {
	sha := sha1.New()
	sha.Write([]byte(self.Salt))
	sha.Write([]byte(self.Password))
	self.EncPassword = fmt.Sprintf("%x", sha.Sum(nil))
	return self.EncPassword
}
