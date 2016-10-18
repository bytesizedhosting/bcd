package engine

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins/deluge"
	"log"
	"net/rpc"
	"testing"
)

var engine RpcEngine
var rpcengine CoreRPC

func init() {
	engine = RpcEngine{}
	engine.server = rpc.NewServer()
	engine.server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	rpcengine = CoreRPC{&engine}
	engine.server.Register(&rpcengine)
	dockerClient, _ := docker.NewClient("unix:///var/run/docker.sock")
	a, err := deluge.New(dockerClient)
	if err != nil {
		log.Fatal("Could not create Plugin")
	}
	engine.Activate(a)
}

func TestGetVersion(t *testing.T) {
	var version string

	rpcengine.GetVersion(1, &version)
	if version == "" {
		t.Error("Did not receive version")
	}
	if version != core.VerString {
		t.Error("Version via RPC did not match version of Core package")
	}
}

func TestGetManifests(t *testing.T) {
	b := &ManifestResponse{}
	rpcengine.GetManifests(nil, b)

	if len(b.Manifests) == 0 {
		t.Error("Did not receive any manifests")
	}

	if b.Manifests[0].Name != "Deluge" {
		t.Error("Manifest has no name set", b.Manifests[0])
	}
}
