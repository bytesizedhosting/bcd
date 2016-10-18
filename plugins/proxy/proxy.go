package proxy

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"net/url"
)

type ProxySlice []*Proxy

type Proxy struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	ExternalId string `json:"external_id"`
}

type ProxyMap map[string]*url.URL

func (self *ProxyMap) LoadFromConfig(path string) error {
	c := ProxyConfig{}
	err := core.LoadConfig(path, &c)
	if err != nil {
		return err
	}
	for _, p := range c.Proxies {
		self.Add(p.Source, p.Target)
	}
	return nil
}

func (self *ProxyMap) Add(sourceUrl string, targetUrl string) error {
	log.Infof("Adding %s as proxy to %s", sourceUrl, targetUrl)
	u, err := url.Parse(targetUrl)

	if err != nil {
		return err
	}

	pMap := *self
	pMap[sourceUrl] = u
	self = &pMap

	return nil
}
