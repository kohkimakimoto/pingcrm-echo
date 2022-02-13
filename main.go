package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kohkimakimoto/pingcrm-echo/app"
	"github.com/kohkimakimoto/pingcrm-echo/resources"
)

func main() {
	log.SetFlags(0)
	handleError := func(err interface{}) {
		log.Fatal("error: ", err)
	}
	defer func() {
		if err := recover(); err != nil {
			handleError(err)
		}
	}()

	var optVersion bool
	flag.BoolVar(&optVersion, "v", false, "")
	flag.BoolVar(&optVersion, "version", false, "")

	var port string
	flag.StringVar(&port, "port", "8080", "")

	var dev bool
	flag.BoolVar(&dev, "dev", false, "")

	var demo bool
	flag.BoolVar(&demo, "demo", false, "")

	var secret string
	flag.StringVar(&secret, "secret", "12345678901234567890123456789012", "")

	var dataDir string
	flag.StringVar(&dataDir, "data-dir", "", "")

	flag.VisitAll(func(f *flag.Flag) {
		name := f.Name
		if len(name) == 1 {
			// skip one character flag
			return
		}
		if name == "dev" {
			// Some flags does not load from env variable.
			return
		}

		// Load flags from environment variables
		if s := os.Getenv(strings.Replace(strings.ToUpper(name), "-", "_", -1)); s != "" {
			_ = f.Value.Set(s)
		}
	})

	flag.Usage = func() {
		fmt.Println(`Usage: ` + app.Name + ` [OPTIONS...]

A demo application to illustrate how Inertia.js works with inertia-echo.

Version: ` + app.Version + `
Commit Hash: ` + app.CommitHash + `

Options:
  -port N           Port number to listen.
  -dev              Run on a dev mode.
                    On the dev mode, tha app need vite dev server to deliver the frontend assets.
  -demo             Run on a demo mode.
                    On the demo mode, the demo user can NOT be updated.
  -secret SECRET    Secret for signature.
                    It is recommended to be more than 32 bytes.
  -data-dir DIR     A directory to store a sqlite database file and any other generated files.
  -h, -help         Show help.
  -v, -version      Print the version.
`)
	}
	flag.Parse()

	if optVersion {
		// show version
		fmt.Println(app.Name + " version " + app.Version)
		return
	}

	c := app.NewConfig()
	c.Addr = ":" + port
	c.IsDev = dev
	c.IsDemo = demo
	c.Secret = []byte(secret)
	c.DataDir = dataDir
	c.PublicFs = publicFs
	c.ViewsFs = resources.ViewsFs
	c.DatabaseFs = databaseFs

	a, err := app.NewApp(c)
	if err != nil {
		handleError(err)
		return
	}
	defer a.Close()
	if err := a.Start(); err != nil {
		handleError(err)
		return
	}
}

//go:embed public
var publicFs embed.FS

//go:embed database
var databaseFs embed.FS
