package main

import (
	"github.com/src-d/go-compose-installer"
)

func main() {
	cfg := installer.NewDefaultConfig()
	cfg.ProjectName = "engine"
	cfg.Compose = [][]byte{yml}
	cfg.Install.Execs = []*installer.Exec{{
		Service: "bblfshd",
		Cmd:     []string{"bblfshctl", "driver", "install", "--all", "--update"},
	}}

	p, err := installer.New("engine-installer", cfg)
	if err != nil {
		panic(err)
	}

	p.Run()
}

var yml []byte = []byte(`
version: '2'
services:
  bblfshd:
    image: bblfsh/bblfshd:v2.2.0
    volumes:
      - {{.Home}}/.engine/bblfshd:/var/lib/bblfshd
    restart: always
    privileged: true
  jupyter:
    image: jfontan/engine-jupyter:latest
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - {{.Home}}/.engine/dataset:/repositories
`)
