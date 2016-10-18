package engine

import (
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/bytesizedhosting/bcd/core"
	"github.com/bytesizedhosting/bcd/plugins"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

type SimplePluginRes struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type PluginResponse struct {
	Plugins []*SimplePluginRes `json:"plugins"`
}
type ManifestResponse struct {
	Manifests []*plugins.Manifest `json:"manifests"`
}

type RpcEngine struct {
	server  *rpc.Server
	plugins []*plugins.Plugin
	port    string
	config  *core.MainConfig
}

func (self *RpcEngine) Server() *rpc.Server {
	return self.server
}

func NewRpcEngine(config *core.MainConfig) *RpcEngine {
	engine := RpcEngine{}
	engine.server = rpc.NewServer()
	engine.server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	engine.port = config.Port
	engine.config = config
	engine.server.Register(&CoreRPC{&engine})
	return &engine
}

func (self *RpcEngine) Start() {
	log.WithFields(log.Fields{
		"port": self.port,
	}).Info("Starting RPC server")

	l, e := net.Listen("tcp", ":"+self.port)
	if e != nil {
		log.Fatal("Could not bind on port:", e)
	}

	http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, _ := r.BasicAuth()

		log.Debugf("received username %s and password %s", username, password)
		if username != self.config.ApiKey || password != self.config.ApiSecret {
			log.Debug("Wrong authentication, not processing")
			log.Debugf("Expected %s and %s", self.config.ApiKey, self.config.ApiSecret)
			w.WriteHeader(401)
			return
		}

		if r.URL.Path == "/rpc" {
			log.Debug("Received call to RPC interface")
			body, err := ioutil.ReadAll(r.Body)
			if err == nil {
				log.Debugf("HTTP Body: %s", body)
			}
			// Place the data back in the buffer so we can process it normally
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			serverCodec := jsonrpc.NewServerCodec(&HttpConn{in: r.Body, out: w})
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(200)
			err = self.server.ServeRequest(serverCodec)
			if err != nil {
				log.Printf("Error while serving JSON request: %v", err)
				return
			}
		}

	}))
}

func (self *RpcEngine) Activate(p plugins.Plugin) {
	log.WithFields(log.Fields{
		"plugin":  p.GetName(),
		"version": p.GetVersion(),
	}).Info("RPC Engine is activating plugin.")

	// Register the plugins RPC exposed methods
	p.RegisterRPC(self.server)

	// Add enabled plugins
	self.plugins = append(self.plugins, &p)
}
