package main

import (
	"github.com/src-d/go-compose-installer"
)

func main() {
	// New default Config, containing a suite of standard messages to be
	// printed at every operation.
	cfg := installer.NewDefaultConfig()

	// Name of the project, to be used in the messages.
	cfg.ProjectName = "example"

	// Compose YAML content, a standard docker compose version 2 file.
	cfg.Compose = [][]byte{yml}

	// The defined YAML, contains a template, some variables are not standard
	// variables, so we need to define it at the `Config.TemplateVars` field.
	cfg.TemplateVars = map[string]interface{}{
		"RedisTag": "4.0.6-alpine",
	}

	// Customized message for a success installation.
	cfg.Install.Messages.Success = "" +
		"The example was successfully installed!\n\n" +

		"To test the deployment please navigate to:\n" +
		"http://localhost:5000\n\n" +

		"To uninstall this example, just execute:\n" +
		"./example uninstall\n"

	// New instance of a Installer based on the given Config.
	p, err := installer.New("compose-installer-example", cfg)
	if err != nil {
		panic(err)
	}

	// Execution of the application.
	p.Run()
}

// Standard docker compose yaml file from:
// https://docs.docker.com/compose/gettingstarted/#where-to-go-next
var yml []byte = []byte(`
version: '2'
services:
  web:
    image: srcd/compose-example
    ports:
     - "5000:5000"
  redis:
    image: "redis:{{.RedisTag}}"
`)
