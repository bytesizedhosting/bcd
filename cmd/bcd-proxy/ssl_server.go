package main

import (
	"crypto/tls"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/context"
	"net/http"
	"path"
	"strings"
)

func noByshioHP(ctx context.Context, host string) error {
	log.Debugln("Checking certificate policy against:", host)
	if strings.Contains(host, "bysh.io") == true {
		log.Warnln("Domain contains bysh.io, we are not allowed to create SSL here.")
		return errors.New("Can't create https routes for bysh.io stock domain, please enable your own domain.")
	}
	return nil
}

func defaultCachePath() string {
	home, err := core.Homedir()
	if err != nil {
		log.Panic("Could not find out homedir, exiting")
	}
	return path.Join(home, ".config", "bcd-proxy")
}

func NewAutoHttpsServer(cachePath string, email string, handler http.Handler) (*http.Server, http.Handler) {
	if cachePath == "" {
		cachePath = defaultCachePath()
	}
	log.Infoln("Setting up auto-manager with cache at", cachePath)

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: noByshioHP,
		Cache:      autocert.DirCache(cachePath),
		Email:      email,
	}

	autocert := m.HTTPHandler(handler)

	s := &http.Server{
		Addr:      ":https",
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		Handler:   handler,
	}

	return s, autocert
}
