package rtorrent

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
	"os"
	"path"
)

const imageName = "bytesized/rutorrent"

func New(client *docker.Client) (*Rtorrent, error) {
	manifest, err := plugins.LoadManifest("rtorrent")
	if err != nil {
		return nil, err
	}

	return &Rtorrent{plugins.Base{DockerClient: client, Name: "rtorrent", Version: 1, Manifest: manifest}}, nil
}

type Rtorrent struct {
	plugins.Base
}

type RtorrentInstallRes struct {
	Message string        `json:"message"`
	Error   bool          `json:"error"`
	Opts    *RtorrentOpts `json:"options"`
}

type RtorrentOpts struct {
	plugins.BaseOpts
	InternalPort string `json:"internal_port,omitempty"`
	DhtPort      string `json:"dht_port,omitempty"`
}

func (self *Rtorrent) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&RtorrentRPC{base: self, BaseRPC: *rpc})
}

func (self *Rtorrent) Install(opts *RtorrentOpts) error {
	log.Infoln("Starting Rtorrent installation")
	var err error
	var ports []string

	for i := 0; i < 4; i++ {
		p, err := core.GetFreePort()
		if err != nil {
			return err
		}
		ports = append(ports, p)
	}

	if opts.WebPort == "" {
		opts.WebPort = ports[0]
	}
	if opts.InternalPort == "" {
		opts.InternalPort = ports[1]
	}
	if opts.DhtPort == "" {
		opts.DhtPort = ports[2]
	}

	err = opts.SetDefault("rtorrent")
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":        self.Name,
		"data_folder":   opts.DataFolder,
		"config_folder": opts.ConfigFolder,
		"username":      opts.Username,
		"password":      opts.Password,
		"dht_port":      opts.DhtPort,
		"web_port":      opts.WebPort,
		"internal_port": opts.InternalPort,
	}).Debug("Current Rtorrent options")

	err = os.MkdirAll(path.Join(opts.ConfigFolder, "/rtorrent/"), 0755)
	if err != nil {
		return err
	}

	err = core.CreateHttpAuth(path.Join(opts.ConfigFolder, "/.htpasswd"), opts.Username, opts.Password)
	if err != nil {
		return err
	}

	err = core.EnsurePath(path.Join(opts.ConfigFolder, "/nginx/"))
	if err != nil {
		return err
	}
	err = self.WriteTemplate("plugins/rtorrent/data/nginx.conf", path.Join(opts.ConfigFolder, "/nginx/nginx.conf"), opts)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/rtorrent/data/rtorrent.rc", path.Join(opts.ConfigFolder, "/rtorrent/rtorrent.rc"), opts)
	if err != nil {
		return err
	}

	log.Debugln("Pulling docker image", imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	portBindings := map[docker.Port][]docker.PortBinding{
		docker.Port(opts.InternalPort + "/tcp"): []docker.PortBinding{docker.PortBinding{HostPort: opts.InternalPort}},
		docker.Port(opts.DhtPort + "/tcp"):      []docker.PortBinding{docker.PortBinding{HostPort: opts.DhtPort}},
		"80/tcp": []docker.PortBinding{docker.PortBinding{HostPort: opts.WebPort}},
	}
	log.Println(portBindings)

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: imageName}

	hostConfig := docker.HostConfig{
		PortBindings: portBindings,
		NetworkMode:  "host",
		Binds:        plugins.DefaultBindings(opts),
	}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_rtorrent_" + opts.WebPort})

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
