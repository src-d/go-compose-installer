package pkgr

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
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

	compose, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ProjectName:   c.ProjectName,
			ComposeBytes:  c.Compose,
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

func (p *Project) Install(ctx context.Context) error {
	return p.c.Install.Run(p, p.c, func(*Project, *Config, Wrapper) error {
		return p.Compose.Up(ctx, options.Up{})
	})
}

func (p *Project) Start(ctx context.Context) error {
	return p.c.Start.Run(p, p.c, func(*Project, *Config, Wrapper) error {
		return p.Compose.Start(ctx)
	})
}

func (p *Project) Stop(ctx context.Context) error {
	return p.c.Stop.Run(p, p.c, func(*Project, *Config, Wrapper) error {
		return p.Compose.Stop(ctx, 0)
	})
}

func (p *Project) Uninstall(ctx context.Context, clean bool) error {
	return p.c.Uninstall.Run(p, p.c, func(*Project, *Config, Wrapper) error {
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
