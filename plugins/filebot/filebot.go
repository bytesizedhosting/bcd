package filebot

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins"
	"github.com/fsouza/go-dockerclient"
	"net/rpc"
	"os"
	"path"
)

type Filebot struct {
	plugins.Base
	imageName string
}

func New(client *docker.Client) (*Filebot, error) {
	manifest, err := plugins.LoadManifest("filebot")

	if err != nil {
		return nil, err
	}

	return &Filebot{Base: plugins.Base{DockerClient: client, Name: "filebot", Version: 1, Manifest: manifest}, imageName: "bytesized/filebot"}, nil
}

func (self *Filebot) RegisterRPC(server *rpc.Server) {
	rpc := plugins.NewBaseRPC(self)
	server.Register(&FilebotRPC{base: self, BaseRPC: *rpc})
}

type FilebotOpts struct {
	plugins.BaseOpts
	OutputFolder  string `json:"output_folder"`
	InputFolder   string `json:"input_folder"`
	FilebotAction string `json:"filebot_action"`
	SubtitleLang  string `json:"subtitle_lang"`
}

func (self *Filebot) Install(opts *FilebotOpts) error {
	var err error

	err = opts.SetDefault(self.Name)
	if err != nil {
		return err
	}

	// We are overwriting any given options here as the data_folder should always be the complete host for Filebot to work
	log.Debugln("Overwriting data_folder with homefolder")
	opts.DataFolder = opts.User.HomeDir

	if opts.OutputFolder == "" {
		opts.OutputFolder = "/host/media"
	}

	if opts.InputFolder == "" {
		opts.InputFolder = "/host/data/completed"
	}

	if opts.FilebotAction == "" {
		opts.InputFolder = "symlink"
	}

	err = self.WriteTemplate("plugins/filebot/data/filebot.sh", path.Join(opts.ConfigFolder, "/filebot.sh"), opts)
	if err != nil {
		return err
	}

	err = os.Chmod(path.Join(opts.ConfigFolder, "filebot.sh"), 0744)
	if err != nil {
		return err
	}

	err = self.WriteTemplate("plugins/filebot/data/filebot.conf", opts.ConfigFolder+"/filebot.conf", opts)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"plugin":         self.Name,
		"config_folder":  opts.ConfigFolder,
		"data_folder":    opts.DataFolder,
		"input_folder":   opts.InputFolder,
		"output_folder":  opts.OutputFolder,
		"filebot_action": opts.FilebotAction,
	}).Debug("Plugin options")

	log.Debugln("Pulling docker image", self.imageName)
	err = self.DockerClient.PullImage(docker.PullImageOptions{Repository: self.imageName}, docker.AuthConfiguration{})
	if err != nil {
		return err
	}

	hostConfig := docker.HostConfig{
		Binds: []string{opts.DataFolder + ":/host", opts.ConfigFolder + ":/config"},
	}

	conf := docker.Config{Env: []string{"PUID=" + opts.User.Uid}, Image: self.imageName}

	log.Debugln("Creating docker container")
	c, err := self.DockerClient.CreateContainer(docker.CreateContainerOptions{Config: &conf, HostConfig: &hostConfig, Name: "bytesized_filebot_" + opts.WebPort})

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
