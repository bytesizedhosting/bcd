//go:generate bcd-generate rpc
package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/engines"
	"github.com/bytesizedhosting/bcd/plugins/cardigann"
	"github.com/bytesizedhosting/bcd/plugins/couchpotato"
	"github.com/bytesizedhosting/bcd/plugins/deluge"
	"github.com/bytesizedhosting/bcd/plugins/filebot"
	"github.com/bytesizedhosting/bcd/plugins/jobs"
	"github.com/bytesizedhosting/bcd/plugins/murmur"
	"github.com/bytesizedhosting/bcd/plugins/nzbget"
	"github.com/bytesizedhosting/bcd/plugins/plex"
	"github.com/bytesizedhosting/bcd/plugins/plexpy"
	"github.com/bytesizedhosting/bcd/plugins/plexrequests"
	"github.com/bytesizedhosting/bcd/plugins/proxy"
	"github.com/bytesizedhosting/bcd/plugins/rocketchat"
	"github.com/bytesizedhosting/bcd/plugins/rtorrent"
	"github.com/bytesizedhosting/bcd/plugins/sickrage"
	"github.com/bytesizedhosting/bcd/plugins/sonarr"
	"github.com/bytesizedhosting/bcd/plugins/stats"
	"github.com/bytesizedhosting/bcd/plugins/subsonic"
	"github.com/bytesizedhosting/bcd/plugins/syncthing"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	app      = kingpin.New("bcd", "The Bytesized Connect Daemon")
	port     = app.Flag("port", "Port to run the RPC server on").Default("8112").String()
	endpoint = app.Flag("docker-socket", "Location of the docker socket").Default("unix:///var/run/docker.sock").String()
	logLevel = app.Flag("log-level", "Log level").Default("debug").String()

	register  = app.Command("init", "Initialize this instance of bcd")
	apikey    = register.Arg("apikey", "Apikey supplied by your provider").Required().String()
	apisecret = register.Arg("apisecret", "Apisecret supplied by your provider").Required().String()

	start = app.Command("start", "Start BCD")
)

func startApp(config *core.MainConfig) {
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Infof("Could not parse log level %v, using 'debug' instead. %v", *logLevel, err)
		level = log.DebugLevel
	}
	log.SetLevel(level)
	log.Infoln("Set logging level to", *logLevel)
	dockerClient, _ := docker.NewClient(*endpoint)

	engine := engine.NewRpcEngine(config)

	// Can we DRY this up?
	deluge, err := deluge.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(deluge)
	}

	plex, err := plex.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(plex)
	}

	rocketchat, err := rocketchat.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(rocketchat)
	}

	syncthing, err := syncthing.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(syncthing)
	}

	sickrage, err := sickrage.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(sickrage)
	}

	couchpotato, err := couchpotato.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(couchpotato)
	}

	plexpy, err := plexpy.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(plexpy)
	}

	rtorrent, err := rtorrent.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(rtorrent)
	}

	nzbget, err := nzbget.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(nzbget)
	}

	sonarr, err := sonarr.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(sonarr)
	}

	cardigann, err := cardigann.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(cardigann)
	}

	plexrequests, err := plexrequests.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(plexrequests)
	}

	subsonic, err := subsonic.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(subsonic)
	}

	murmur, err := murmur.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(murmur)
	}

	filebot, err := filebot.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(filebot)
	}

	engine.Activate(stats.New())
	engine.Activate(jobrpc.New())
	engine.Activate(proxy.New())
	engine.Start()
}

func main() {
	log.Println("Starting the Bytesized Connect Daemon", core.VerString)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case register.FullCommand():
		log.Infoln("Initialization run, writing config file.")
		c := core.MainConfig{ApiSecret: *apisecret, ApiKey: *apikey, Port: *port}
		err := core.WriteConfig("config.json", &c)
		if err != nil {
			log.Panicf("Could not write config file: '%s'", err.Error())
		}
		log.Infoln("Initialization run completed, please start the daemon normally.")
	case start.FullCommand():
		c := core.MainConfig{}
		c.Port = *port
		err := core.LoadHomeConfig("config.json", &c)

		if err != nil {
			log.Info("Could not load config file. Please run with '-init APIKEY APISECRET' first.", err.Error())
			os.Exit(1)
		}

		log.Debugf("Using docker socket '%s'", *endpoint)
		startApp(&c)
	}

}
