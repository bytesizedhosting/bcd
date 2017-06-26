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
	"github.com/bytesizedhosting/bcd/plugins/headphones"
	"github.com/bytesizedhosting/bcd/plugins/jackett"
	"github.com/bytesizedhosting/bcd/plugins/jobs"
	"github.com/bytesizedhosting/bcd/plugins/murmur"
	"github.com/bytesizedhosting/bcd/plugins/nzbget"
	"github.com/bytesizedhosting/bcd/plugins/plex"
	"github.com/bytesizedhosting/bcd/plugins/plexpy"
	"github.com/bytesizedhosting/bcd/plugins/plexrequests"
	"github.com/bytesizedhosting/bcd/plugins/portainer"
	"github.com/bytesizedhosting/bcd/plugins/proxy"
	"github.com/bytesizedhosting/bcd/plugins/radarr"
	"github.com/bytesizedhosting/bcd/plugins/resilio"
	"github.com/bytesizedhosting/bcd/plugins/rocketchat"
	"github.com/bytesizedhosting/bcd/plugins/rtorrent"
	"github.com/bytesizedhosting/bcd/plugins/sickrage"
	"github.com/bytesizedhosting/bcd/plugins/sonarr"
	"github.com/bytesizedhosting/bcd/plugins/stats"
	"github.com/bytesizedhosting/bcd/plugins/subsonic"
	"github.com/bytesizedhosting/bcd/plugins/syncthing"
	"github.com/bytesizedhosting/bcd/plugins/vnc"
	"github.com/bytesizedhosting/bcd/plugins/znc"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	app      = kingpin.New("bcd", "The Bytesized Connect Daemon")
	port     = app.Flag("port", "Port to run the RPC server on").Default("8112").String()
	logLevel = app.Flag("log-level", "Log level").Default("info").String()

	endpoint = app.Flag("docker-endpoint", "Docker endpoint to use").Default("unix:///var/run/docker.sock").String()

	dockerTLS = app.Flag("docker-tls", "Connect to a TLS enabled Docker daemon").Bool()
	dockerEnv = app.Flag("docker-env", "Connect to Docker using Docker-machine environment variables").Bool()

	ca   = app.Flag("docker-ca-path", "Path to your CA file, required for TLS connection.").String()
	cert = app.Flag("docker-cert-path", "Path to your cert file, required for TLS connection.").String()
	key  = app.Flag("docker-key-path", "Path to your key file, required for TLS connection.").String()

	register  = app.Command("init", "Initialize this instance of bcd")
	apikey    = register.Arg("apikey", "Apikey supplied by your provider").Required().String()
	apisecret = register.Arg("apisecret", "Apisecret supplied by your provider").Required().String()

	start = app.Command("start", "Start BCD")
)

func startApp(config *core.MainConfig) {
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Infof("Could not parse log level %v, using 'info' instead. %v", *logLevel, err)
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.Infoln("Set logging level to", *logLevel)

	var dockerClient *docker.Client

	if *dockerTLS == true {
		log.Infoln("Connecting to Docker daemon via TLS")
		dockerClient, err = docker.NewTLSClient(*endpoint, *cert, *key, *ca)
	} else if *dockerEnv == true {
		log.Infoln("Connecting to Docker daemon via environment variables")
		dockerClient, err = docker.NewClientFromEnv()
	} else {
		dockerClient, err = docker.NewClient(*endpoint)
	}

	if err != nil {
		log.Errorf("Could not connect to Docker daemon: '%s'", err.Error())
		os.Exit(1)
	}

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

	resilio, err := resilio.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(resilio)
	}

	headphones, err := headphones.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(headphones)
	}

	jackett, err := jackett.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(jackett)
	}

	vnc, err := vnc.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(vnc)
	}

	znc, err := znc.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(znc)
	}

	radarr, err := radarr.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(radarr)
	}

	portainer, err := portainer.New(dockerClient)
	if err != nil {
		log.Infoln("Could not enable plugin: ", err)
	} else {
		engine.Activate(portainer)
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
