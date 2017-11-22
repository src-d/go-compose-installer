package pkgr

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	homedir "github.com/mitchellh/go-homedir"
)

type Project struct {
	Compose project.APIProject
	Docker  *client.Client

	c *Config
}

func NewProject(c *Config) (*Project, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	bytes, err := renderComposeFiles(c.Compose)
	if err != nil {
		return nil, err
	}

	compose, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ProjectName:   c.ProjectName,
			ComposeBytes:  bytes,
			LoggerFactory: &logger.RawLogger{},
		},
	}, nil)

	if err != nil {
		return nil, err
	}

	compose.AddListener(NewListener(compose.(*project.Project)))

	return &Project{
		Compose: compose,
		Docker:  cli,

		c: c,
	}, nil
}

func renderComposeFiles(files [][]byte) ([][]byte, error) {
	var err error
	for i, file := range files {
		files[i], err = renderComposeFile(file)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

func renderComposeFile(file []byte) ([]byte, error) {
	tmpl, err := template.New("test").Parse(string(file))
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	home, _ := homedir.Dir()
	err = tmpl.Execute(buf, map[string]interface{}{
		"Home": home,
	})

	return buf.Bytes(), err
}

func (p *Project) Install(ctx context.Context) error {
	return p.c.Install.Run(p, p.c, func(*Project, *Config) error {
		return p.Compose.Up(ctx, options.Up{})
	})
}

func (p *Project) isInstalled(ctx context.Context) error {
	info, err := p.Compose.Ps(ctx)
	if err != nil {
		return err
	}

	if len(info) != 0 {
		return nil
	}

	return fmt.Errorf("%s is not installed, run install first", p.c.ProjectName)
}

func (p *Project) Start(ctx context.Context) error {
	return p.c.Start.Run(p, p.c, func(*Project, *Config) error {
		if err := p.isInstalled(ctx); err != nil {
			return err
		}

		return p.Compose.Start(ctx)
	})
}

func (p *Project) Stop(ctx context.Context) error {
	return p.c.Stop.Run(p, p.c, func(*Project, *Config) error {
		if err := p.isInstalled(ctx); err != nil {
			return err
		}

		return p.Compose.Stop(ctx, 0)
	})
}

func (p *Project) Uninstall(ctx context.Context, clean bool) error {
	return p.c.Uninstall.Run(p, p.c, func(*Project, *Config) error {
		if err := p.isInstalled(ctx); err != nil {
			return err
		}

		if err := p.Compose.Stop(ctx, 0); err != nil {
			return err
		}

		opts := options.Down{}
		if clean {
			opts.RemoveImages = "all"
			opts.RemoveVolume = true
		}

		return p.Compose.Down(ctx, opts)
	})
}

func (p *Project) Status(ctx context.Context) error {
	info, err := p.Compose.Ps(ctx)
	fmt.Println(info.String([]string{"Name", "Command", "State", "Ports"}, true))
	return err
}

func (p *Project) Execute(ctx context.Context, service string, cmd ...string) error {
	srv, err := p.Compose.CreateService(service) //, options.Run{Detached: true})
	cs, err := srv.Containers(ctx)
	if err != nil {
		return err
	}

	for _, c := range cs {
		if !c.IsRunning(ctx) {
			continue
		}

		if err := p.doExecute(ctx, c.ID(), cmd); err != nil {
			return err
		}
	}

	return err
}

func (p *Project) doExecute(ctx context.Context, containerID string, cmd []string) error {
	cfg := types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	id, err := p.Docker.ContainerExecCreate(ctx, containerID, cfg)
	if err != nil {
		return err
	}

	r, err := p.Docker.ContainerExecAttach(ctx, id.ID, cfg)
	if err != nil {
		return err
	}

	if _, err = io.Copy(os.Stdout, r.Reader); err != nil {
		return err
	}

	resp, err := p.Docker.ContainerExecInspect(ctx, id.ID)
	if err != nil {
		return err
	}

	if resp.ExitCode != 0 {
		return fmt.Errorf("error executing command exit %d", resp.ExitCode)
	}

	return err
}
