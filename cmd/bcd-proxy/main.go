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
	port     = kingpin.Flag("port", "Port to run BCD-Proxy on").Default("80").String()
	userName = kingpin.Flag("username", "User configuration to read").Default("bytesized").String()
)

func init() {
	log.SetLevel(log.DebugLevel)
	kingpin.Parse()

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
}

func main() {
	proxy := NewMultipleHostReverseProxy(proxyPath)

	log.Fatal(http.ListenAndServe(":"+*port, proxy))
}
