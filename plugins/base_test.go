package plugins

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

type TemplOpts struct {
	BaseOpts
	DaemonPort string
}

func TestWriteTemplate(t *testing.T) {
	dir, err := ioutil.TempDir("", "bcdtest")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up
	tmpfn := filepath.Join(dir, "tmpfile")

	opts := &TemplOpts{BaseOpts: BaseOpts{}, DaemonPort: "9999"}
	b := Base{}
	log.Println(tmpfn)
	err = b.WriteTemplate("plugins/deluge/data/core.conf", tmpfn, opts)
	if err != nil {
		t.Error("Could not write template:", err)
	}
	data, err := ioutil.ReadFile(tmpfn)
	if err != nil {
		t.Error("Could not read created file:", err)
	}
	if !bytes.Contains(data, []byte("9999")) {
		t.Error("Template does not contain specific port")
		log.Println(string(data[:]))
	}
}

func TestLoadManifest(t *testing.T) {
	_, err := LoadManifest("deluge")
	if err != nil {
		t.Error("Could not load Manifest:", err)

	}

	_, err = LoadManifest("dontexistdeluge")
	if err == nil {
		t.Error("No error thrown on faulty Manifest")

	}
}

type TestOpts struct {
	BaseOpts
	InternalPort string `json:"internal_port,omitempty"`
	DhtPort      string `json:"dht_port,omitempty"`
	DataFolder   string `json:"data_folder,omitempty"`
}
