package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/plugins/proxy"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"regexp"
	"time"
)

func NewMultipleHostReverseProxy(proxyPath string, unknownHost string) *httputil.ReverseProxy {
	cleanReg := regexp.MustCompile(`:\d*`)
	lastUpdate := time.Now()
	proxyMap := &proxy.ProxyMap{}

	/*
		Yes, we are reading a file here and use it as a cheap database
		I expect to do a few requests every few hours so we don't need anything
		better right now. <insert dealwithit.gif>
	*/

	err := proxyMap.LoadFromConfig(proxyPath)
	if err != nil {
		log.Infoln("Could not read config file, starting with empty state.", err)
		log.Infoln(proxyMap)
	}

	info, err := os.Stat(proxyPath)
	if err != nil {
		log.Debugln("Could not get last modification time, setting to now")
	} else {
		lastUpdate = info.ModTime()
	}

	director := func(req *http.Request) {
		log.Debugln("Incoming request from:", req.Host)
		log.Debugln("Getting last modifcation time")

		info, err := os.Stat(proxyPath)
		if err != nil {
			log.Debugln("Could not get last modification time, breaking")
			return
		}

		log.Debugln("Modifcation time is", info.ModTime())
		log.Debugln("Our last update was", lastUpdate)

		if info.ModTime().After(lastUpdate) {
			err := proxyMap.LoadFromConfig(proxyPath)
			if err != nil {
				log.Infoln("Could not read config file, starting with empty state.", err)
			}
			log.Println("Modification time was after our last, updating proxy map:", proxyMap)
			lastUpdate = time.Now()
		}

		domain := cleanReg.ReplaceAllString(req.Host, "")
		m := *proxyMap
		if dom, ok := m[domain]; ok {
			p := path.Join(dom.Path, req.URL.Path)
			scheme := dom.Scheme
			h := dom.Host
			req.URL.Scheme = scheme
			req.URL.Host = h
			req.URL.Path = p
			log.Debugln("Request string", req.URL.String())
			log.Debugln("Raw Query", req.URL.RawQuery)
			log.Debugln("Host", h)
			log.Debugln("Path", p)
		} else {
			log.Debugln("No routes known for the incoming url.")
			if unknownHost == "" {
				log.Debugln("No custom host set, redirecting to standard error domain.")
				req.URL.Scheme = "http"
				req.URL.Host = "download.bytesized-hosting.com"
				req.URL.Path = "/noproxy.html"
			} else {
				log.Debugln("Custom host set, redirecting.")
				req.URL.Scheme = "http"
				req.URL.Host = unknownHost
			}
		}
	}
	return &httputil.ReverseProxy{Director: director}
}
