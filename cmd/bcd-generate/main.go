package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	app        = kingpin.New("bcd-generate", "Generation tools for BCD")
	rpc        = app.Command("rpc", "Regenerate all RPC wrappers")
	newPlugin  = app.Command("plugin", "Create a new plugin")
	pluginName = newPlugin.Arg("name", "The name for plugin").Required().String()
)
var Blacklist = map[string]bool{"jobs": true, "stats": true, "proxy": true}

type RpcTemplate struct {
	Name      string
	LowerName string
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	return fileInfo.IsDir(), err
}
func split(str string) []string {
	return strings.Split(regexp.MustCompile(`-|_|([a-z])([A-Z])`).ReplaceAllString(strings.Trim(str, `-|_| `), `$1 $2`), ` `)
}

func Titleize(str string) string {
	pieces := split(str)

	for i := 0; i < len(pieces); i++ {
		pieces[i] = fmt.Sprintf(`%v%v`, strings.ToUpper(string(pieces[i][0])), strings.ToLower(pieces[i][1:]))
	}

	return strings.Join(pieces, ` `)
}
func WriteTemplate(tmplName string, outputFile string, pkgName string) error {
	pkg := RpcTemplate{Name: Titleize(pkgName), LowerName: pkgName}
	log.Println("Generating template for package:", pkg.Name)
	log.Println("Saving template-file to folder:", outputFile)
	tmplFile, err := ioutil.ReadFile(path.Join("cmd/bcd-generate/templates", tmplName))
	tmpl, err := template.New("RPCTemplate").Parse(string(tmplFile[:]))
	if err != nil {
		return err
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	err = tmpl.Execute(file, pkg)
	if err != nil {
		return err
	}
	log.Println("File succesfully generated")
	return nil
}

func PluginPath() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	log.Println("Locating plugins, starting here.", dir)
	pluginFolder := path.Join(dir, "plugins")

	return pluginFolder, nil
}

func main() {
	log.Println("BCD Generate")
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case newPlugin.FullCommand():
		pkgName := strings.ToLower(*pluginName)
		log.Println("Creating new plugin folder for:", pkgName)
		pluginPath, err := PluginPath()
		pluginPath = path.Join(pluginPath, pkgName)
		if err != nil {
			log.Panic(err)
		}
		log.Println("Creating plugin here:", pluginPath)
		err = os.MkdirAll(pluginPath, 0755)
		if err != nil {
			log.Panic(err)
		}
		t := path.Join(pluginPath, pkgName+".go")
		err = WriteTemplate("plugin.go.templ", t, pkgName)
		if err != nil {
			log.Panic(err)
		}
		t = path.Join(pluginPath, "rpc_proxy.go")
		err = WriteTemplate("rpc.go.templ", t, pkgName)
		if err != nil {
			log.Panic(err)
		}
		log.Println("Plugin created")

	case rpc.FullCommand():
		target, err := PluginPath()
		log.Println("Attempting to find plugins in:", target)
		if err != nil {
			log.Panic(err)
		}
		files, err := filepath.Glob(path.Join(target, "*"))
		if err != nil {
			log.Panic(err)
		}
		for _, f := range files {
			d, _ := IsDirectory(f)
			if d {
				m := regexp.MustCompile(`(\w*)$`)
				mpath := regexp.MustCompile(`(.*)\/(\w*)$`)
				packageName := m.FindString(f)
				packagePath := mpath.FindString(f)
				if Blacklist[packageName] {
					log.Printf("Package '%s' is in black list, skipping", packageName)
					continue
				}
				t := path.Join(packagePath, "rpc_proxy.go")
				err := WriteTemplate("rpc.go.templ", t, packageName)
				if err != nil {
					log.Warnln("Could not create template from data in:", f, err)
				}
			}

		}
	}
}
