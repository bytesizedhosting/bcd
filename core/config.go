package core

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

const (
	versionMajor = 0
	versionMinor = 17
	versionPatch = 0
)

var (
	VerString string
)

func init() {
	VerString = fmt.Sprintf("%d.%d.%d", versionMajor, versionMinor, versionPatch)
}

type Config struct {
}

type MainConfig struct {
	Config
	ApiKey    string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
	Port      string `json:"port"`
}

func Homedir() (string, error) {
	curUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return curUser.HomeDir, nil
}
func ConfigPath() (string, error) {
	home, err := Homedir()
	if err != nil {
		return "", err
	}
	configPath := home + "/.config/bcd"

	return configPath, nil
}
func LoadHomeConfig(fileName string, conf interface{}) error {
	configPath, err := ConfigPath()
	if err != nil {
		log.Infoln("Error:", err)
		return err
	}
	f := path.Join(configPath, fileName)

	return LoadConfig(f, conf)
}

func LoadConfig(filePath string, conf interface{}) error {
	log.Infoln("Loading config file in:", filePath)
	result, err := ioutil.ReadFile(filePath)
	log.Infoln("Loaded config:", string(result))

	if err != nil {
		return err
	}

	err = json.Unmarshal(result, conf)
	if err != nil {
		return err
	}

	return nil
}
func WriteConfig(fileName string, conf interface{}) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(configPath, 0755)
	if err != nil {
		return err
	}

	c, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	configFile := path.Join(configPath, fileName)

	log.Debug("Writing config file to", configFile)

	err = ioutil.WriteFile(configFile, []byte(c[:]), 0700)
	if err != nil {
		return err
	}

	return nil
}
