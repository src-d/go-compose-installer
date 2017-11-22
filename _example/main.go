package main

import (
	"github.com/src-d/pkgr"
)

func main() {
	cfg := pkgr.NewDefaultConfig()
	cfg.ProjectName = "engine"
	cfg.Compose = [][]byte{yml}
	cfg.Install.Execs = []*pkgr.Exec{{
		Service: "bblfshd",
		Cmd:     []string{"bblfshctl", "driver", "install", "--all", "--update"},
	}}

	p, err := pkgr.NewProgram("engine-installer", cfg)
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
