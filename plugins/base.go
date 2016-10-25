package plugins

// https://github.com/kisielk/jsonrpc-example/blob/master/server.go

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/fsouza/go-dockerclient"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"net/rpc"
	"os"
	"os/user"
	"path"
	"text/template"
)

type Manifest struct {
	Version        float32                   `json:"version"`
	ExposedMethods []string                  `json:"exposed_methods"`
	MethodOptions  map[string][]MethodOption `json:"method_options",yaml: "method_options,flow"`
	ShowOptions    []string                  `json:"show_options"`
	Name           string                    `json:"name"`
	RpcName        string                    `json:"rpc_name"`
	WebUrlFormat   string                    `json:"web_url_format"`
	Description    string                    `json:"description"`
}

type MethodOption struct {
	Name          string `json:"name"`
	DefaultValue  string `json:"default_value", yaml:"default_value"`
	Type          string `json:"type"`
	Hint          string `json:"hint"`
	AllowDeletion bool   `json:"allow_deletion"`
}
type BaseOpts struct {
	ContainerId  string     `json:"container_id,omitempty"`
	Password     string     `json:"password,omitempty"`
	WebPort      string     `json:"web_port,omitempty"`
	RunAsUser    string     `json:"run_as_user,omitempty"`
	Username     string     `json:"username,omitempty"`
	ConfigFolder string     `json:"config_folder,omitempty"`
	DataFolder   string     `json:"data_folder,omitempty"`
	MediaFolder  string     `json:"media_folder,omitempty"`
	NoTemplates  string     `json:"no_templates"`
	User         *user.User `json:"user,omitempty"`
}

func (self *BaseOpts) GetBaseOpts() BaseOpts {
	return *self
}
func (opts *BaseOpts) SetDefault(name string) error {
	if opts.RunAsUser == "" {
		log.Debugln("No run_as_user received, using default 'bytesized'")
		opts.RunAsUser = "bytesized"
	}

	user, err := core.GetUser(opts.RunAsUser)
	if err != nil {
		return err
	}

	opts.User = user
	opts.RunAsUser = user.Username

	if opts.Password == "" {
		log.Debugln("No password supplied, generating one.")
		opts.Password = core.GetRandom(14)
	}

	if opts.ConfigFolder == "" {
		log.Debugln("No config_folder supplied, using default.")
		opts.ConfigFolder = path.Join(opts.User.HomeDir, "config", name)
	}

	if opts.DataFolder == "" {
		log.Debugln("No data_folder supplied, using default.")
		opts.DataFolder = path.Join(opts.User.HomeDir, "data")
	}

	if opts.MediaFolder == "" {
		log.Debugln("No media_folder supplied, using default.")
		opts.MediaFolder = path.Join(opts.User.HomeDir, "media")
	}

	if opts.Username == "" {
		log.Debugln("No username supplied, using default.")
		opts.Username = "bytesized"
	}

	if opts.WebPort == "" {
		p, err := core.GetFreePort()
		if err != nil {
			return err
		}
		opts.WebPort = p
	}

	return nil
}

type Plugin interface {
	GetName() string
	GetManifest() *Manifest
	GetVersion() int
	RegisterRPC(*rpc.Server)
}
type appPlugin interface {
	Plugin
	Start(*AppConfig) error
	Stop(*AppConfig) error
	Status(*AppConfig) (*docker.State, error)
	Restart(*AppConfig) error
	Uninstall(*AppConfig) error
}

func DumpManifest(manifest *Manifest) {
	res, err := yaml.Marshal(manifest)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("result:", string(res))
	}
}

func LoadManifest(name string) (*Manifest, error) {
	manifestLoaded := false
	manifest := Manifest{}

	configPath, err := core.ConfigPath()
	if err != nil {
		log.Debugln("Could not retrieve config path, this is a problem...")
	}

	manifestPath := path.Join(configPath, "manifests", name+".yml")

	if _, err := os.Stat(manifestPath); err == nil {
		log.Debugln("Custom manifest found. Loading from path: ", manifestPath)

		data, err := ioutil.ReadFile(manifestPath)
		if err != nil {
			log.Debugln("Could not load custom manifest: ", err)
		}
		err = yaml.Unmarshal(data, &manifest)
		if err != nil {
			log.Debugln("Could not Ummarshal YAML data. Probably broken manifest.", err)
		}
		if err == nil {
			manifestLoaded = true
		}
	}

	if manifestLoaded == false {
		log.Debugf("No custom manifest could be loaded for %s, using build-in one.", name)
		data, err := Asset("plugins/" + name + "/data/manifest.yml")
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(data, &manifest)
		if err != nil {
			return nil, err
		}
	}

	return &manifest, nil
}

func (self *Base) Status(opts *AppConfig) (*docker.State, error) {
	container, err := self.DockerClient.InspectContainer(opts.ContainerId)
	if err != nil {
		return nil, err
	}

	return &container.State, nil
}

func (self *Base) Stop(opts *AppConfig) error {
	err, exists := self.containerExists(opts.ContainerId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Could not find container to stop with id '%s'", opts.ContainerId)
	}

	err = self.DockerClient.StopContainer(opts.ContainerId, 10)
	if err != nil {
		return err
	}
	return nil
}

func (self *Base) Start(opts *AppConfig) error {
	err, exists := self.containerExists(opts.ContainerId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Could not find container to start with id '%s'", opts.ContainerId)
	}

	err = self.DockerClient.StartContainer(opts.ContainerId, nil)
	if err != nil {
		return err
	}
	return nil
}

func (self *Base) Uninstall(opts *AppConfig) error {
	log.Debugln("Removing docker container with id", opts.ContainerId)
	delOpts := docker.RemoveContainerOptions{Force: true, ID: opts.ContainerId}
	err := self.DockerClient.RemoveContainer(delOpts)
	if err != nil {
		return err
	}
	log.Debugln("Docker container removed")
	return nil
}

func (self *Base) Restart(opts *AppConfig) error {
	self.Stop(opts)

	err := self.Start(opts)
	if err != nil {
		return err
	}
	return nil
}

type Base struct {
	prefix       string
	Name         string `json:"name"`
	Version      int    `json:"version"`
	Manifest     *Manifest
	DockerClient *docker.Client
}

type AppConfig struct {
	ContainerId string
}

func (self *Base) GetManifest() *Manifest {
	return self.Manifest
}

func (s *Base) GetName() string {
	return s.Name
}

func (s *Base) GetVersion() int {
	return s.Version
}

type FakeType struct {
	DataFolder string
	BaseOpts
}

type Options interface {
	GetBaseOpts() BaseOpts
}

func DefaultBindings(opts Options) []string {
	baseOpts := opts.GetBaseOpts()
	return []string{baseOpts.DataFolder + ":/data", baseOpts.ConfigFolder + ":/config", baseOpts.MediaFolder + ":/media"}
}

// TODO: Can we extract the baseOpts from the object interface?
func (s *Base) WriteTemplate(templateName string, outputFile string, object Options) error {
	if object.GetBaseOpts().NoTemplates == "true" {
		log.Debugln("We are not (re-)creating templates since no_templates has been set to true, moving on.")
		return nil
	}

	log.WithFields(log.Fields{
		"plugin":       s.Name,
		"templateName": templateName,
		"outputFile":   outputFile,
	}).Debug("Writing template")

	data, err := Asset(templateName)
	if err != nil {
		return err
	}

	tmpl, err := template.New(templateName).Parse(string(data[:]))

	if err != nil {
		return err
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, object)
	if err != nil {
		return err
	}

	return nil
}

func (self *Base) RegisterRPC(server *rpc.Server) {
	server.Register(self)
}

func (s *Base) containerExists(id string) (error, bool) {
	dockerOpts := docker.ListContainersOptions{All: true, Filters: map[string][]string{"id": {id}}}
	containers, err := s.DockerClient.ListContainers(dockerOpts)
	if err != nil {
		return err, false
	}

	if len(containers) == 0 {
		return nil, false
	}

	return nil, true
}
