
# go-compose-installer [![GoDoc](https://godoc.org/github.com/src-d/go-compose-installer?status.svg)](https://godoc.org/github.com/src-d/go-compose-installer)

A toolkit to create installers based on docker compose.

*go-compose-installer* allows to deploy complex infrastructure applications using a single binary, with no more dependencies than `docker`.

## Example

This is an example based on the [docker compose tutorial](https://docs.docker.com/compose/gettingstarted/#where-to-go-next), this example installs a basic environment based on a web server backed by a redis service.

```go
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
    image: "redis:alpine"
`)

```

The following code provides an CLI application allowing to anyone with just one
command to install, start, stop and uninstall the environment, without any
external dependency besides docker.

```sh
Usage:
  compose-installer-example [OPTIONS] <command>

Help Options:
  -h, --help  Show this help message

Available commands:
  install    Installs example into your system.
  start      Stars example.
  status     Show the status of example.
  stop       Stops example.
  uninstall  Remove example from your system.
```

## Go Template support

Go templates are supported in the yaml file and also at all the messages. This
is the list of supported variables:

- `.Project` - The project name, from the given `Config.ProjectName`.
- `.Home` - Home folder of the user executing the installer.
- `.OS` - Content of runtime.GOOS.
- `.Arch` - Content ofruntime.GOARCH.
- `.Error` - Only available at `Failure` messages, is the string of the error.

# License

Apache License Version 2.0, see [LICENSE](LICENSE)
