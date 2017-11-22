package main

import (
	"github.com/src-d/pkgr"
)

func main() {
	c := &pkgr.Config{}
	c.ProjectName = "engine"
	c.Compose = [][]byte{yml}
	c.Install.Messages = pkgr.Messages{
		Description:  "Installs engine into your system.",
		Announcement: "Installing Engine...",
		Failure:      "Failed to install Engine.",
		Success:      "Successfully installed.",
	}

	c.Install.Execs = []*pkgr.Exec{{
		Service: "bblfshd",
		Cmd:     []string{"bblfshctl", "driver", "install", "--all", "--update"},
	}}

	c.Start.Messages = pkgr.Messages{
		Description:  "Stars engine.",
		Announcement: "Starting Engine...",
		Failure:      "Failed to start Engine.",
		Success:      "Engine started.",
	}
	c.Stop.Messages = pkgr.Messages{
		Description:  "Stops engine.",
		Announcement: "Stopping Engine...",
		Failure:      "Failed to stop Engine.",
		Success:      "Engine stopped.",
	}

	c.Status.Messages = pkgr.Messages{
		Description: "Show the status of engine.",
	}

	c.Uninstall.Messages = pkgr.Messages{
		Description:  "Remove engine from your system.",
		Announcement: "Uninstalling Engine...",
		Failure:      "Failed to uninstall Engine.",
		Success:      "Successfully uninstalled.",
	}

	p, err := pkgr.NewProgram("engine-installer", c)
	if err != nil {
		panic(err)
	}

	p.Run()
}

var yml []byte = []byte(`
bblfshd:
  image: bblfsh/bblfshd:v2.2.0
  volumes:
    - {{.Home}}/.engine/bblfshd:/var/lib/bblfshd
  restart: always
  privileged: true
jupyter:
  image: srcd/engine-jupyter:latest
  ports:
     - "8080:8888"
  volumes:
    - {{.Home}}/.engine/dataset:/repositories`)
