package engine

import (
	"github.com/bytesizedhosting/bcd/core"
)

type CoreRPC struct {
	engine *RpcEngine
}

func (self *CoreRPC) GetManifests(nothing *PluginResponse, res *ManifestResponse) error {
	for _, plugin := range self.engine.plugins {
		p := *plugin
		manifest := p.GetManifest()
		if manifest != nil {
			res.Manifests = append(res.Manifests, manifest)
		}
	}
	return nil
}

func (self *CoreRPC) GetVersion(_ int, res *string) error {
	*res = core.VerString
	return nil
}

func (self *CoreRPC) GetPlugins(nothing *PluginResponse, res *PluginResponse) error {
	for _, plugin := range self.engine.plugins {
		p := *plugin
		res.Plugins = append(res.Plugins, &SimplePluginRes{Name: p.GetName(), Version: p.GetVersion()})
	}

	return nil
}
