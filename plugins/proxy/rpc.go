package proxy

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"net/rpc"
)

type ProxyConfig struct {
	Proxies []*Proxy `json:"proxies"`
}

type ProxyRPC struct {
	plugins.Base
}

func New() *ProxyRPC {
	return &ProxyRPC{plugins.Base{Name: "Proxy", Version: 1}}
}

func (self *ProxyRPC) RegisterRPC(server *rpc.Server) {
	server.Register(self)
}
func (self *ProxyRPC) List(p *Proxy, res *ProxySlice) error {
	c := ProxyConfig{}
	err := core.LoadHomeConfig("proxies.json", &c)
	if err != nil {
		log.Debugln("Error reading conf:", err)
		return err
	}
	*res = c.Proxies
	return nil
}

// TODO: Change it so you replace names, instead of not changing anything.
func (self *ProxyRPC) Add(p *Proxy, res *ProxySlice) error {
	c := ProxyConfig{}
	err := core.LoadHomeConfig("proxies.json", &c)
	if err != nil {
		log.Debugln("Error reading conf, going to assume we don't have a config file yet and we will create one now. This was the error:", err)
		core.WriteConfig("proxies.json", &c)
	}
	found := false
	for _, proxy := range c.Proxies {
		if p.Source == proxy.Source {
			found = true
		}
	}
	if found == false {
		c.Proxies = append(c.Proxies, p)
		core.WriteConfig("proxies.json", &c)
	} else {
		log.Debugf("Source %s already in config, not adding again.", p.Source)
	}
	*res = c.Proxies
	return nil
}

func (self *ProxyRPC) Remove(p *Proxy, res *ProxySlice) error {
	log.Debugln("Received DELETE request")
	c := ProxyConfig{}
	err := core.LoadHomeConfig("proxies.json", &c)
	if err != nil {
		log.Debugln("Error reading conf:", err)
	}
	proxies := []*Proxy{}
	for _, proxy := range c.Proxies {
		if p.Source != proxy.Source {
			log.Debugln("This is not something we want to delete, adding in")
			proxies = append(proxies, proxy)
		}
	}

	c.Proxies = proxies
	core.WriteConfig("proxies.json", &c)

	*res = c.Proxies
	return nil
}
