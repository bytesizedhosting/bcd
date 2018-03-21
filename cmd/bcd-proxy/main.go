package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"path"
)

var proxyPath string

var (
	port        = kingpin.Flag("port", "Port to run BCD-Proxy on").Default("80").String()
	userName    = kingpin.Flag("username", "User who's configuration file to read, usually this is the user running BCD.").Default("bytesized").String()
	cachePath   = kingpin.Flag("cache-path", "Location where to store certificates and keyfiles").String()
	email       = kingpin.Flag("email", "Email address to use for Letsencrypt certificates").String()
	unknownHost = kingpin.Flag("unknown-host", "Host to proxy requests to that are unknown to bcd-proxy").String()
	logLevel    = kingpin.Flag("log-level", "Log level").Default("info").String()
)

func init() {
	kingpin.Parse()
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Infof("Could not parse log level %v, using 'info' instead. %v", *logLevel, err)
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.Infoln("Set logging level to", *logLevel)

	user, err := core.GetUser(*userName)
	if err != nil {
		log.Warnf("Could not get config folder for user %s: '%s", userName, err)
		os.Exit(1)
	} else {
		log.Debugln("Received username option", *userName)
	}

	cpath := path.Join(user.HomeDir, ".config", "bcd")
	proxyPath = path.Join(cpath, "proxies.json")
	log.Debugln("Proxies file will be loaded from:", proxyPath)
	if *unknownHost != "" {
		log.Infoln("Proxying unknown routes to", *unknownHost)
	}
}

func main() {
	proxy := NewMultipleHostReverseProxy(proxyPath, *unknownHost)

	s, autocert := NewAutoHttpsServer(*cachePath, *email, proxy)

	go func() {
		log.Fatal(s.ListenAndServeTLS("", ""))
	}()

	log.Fatal(http.ListenAndServe(":"+*port, autocert))
}
