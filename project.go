package installer

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

const (
	// EnvComposeFile overrides the built-in composer file.
	EnvComposeFile = "INSTALLER_COMPOSE"
)

// Project performs all the operations at docker compose level.
type Project struct {
	Compose project.APIProject
	Docker  *client.Client

	c *Config
}

// NewProject returns a new Project based on the given Config.
func NewProject(c *Config) (*Project, error) {
	p := &Project{c: c}
	return p, p.initialize()
}

func (p *Project) initialize() error {
	var err error
	p.Docker, err = client.NewEnvClient()
	if err != nil {
		return err
	}

	bytes, err := p.renderComposeFiles(p.c.Compose)
	if err != nil {
		return err
	}

	p.Compose, err = docker.NewProject(&ctx.Context{
		Context: project.Context{
			ProjectName:   p.c.ProjectName,
			ComposeBytes:  bytes,
			LoggerFactory: &logger.RawLogger{},
		},
	}, nil)

	if err != nil {
		return err
	}

	p.Compose.AddListener(NewListener(p.Compose.(*project.Project)))
	return nil
}

func (p *Project) renderComposeFiles(files [][]byte) ([][]byte, error) {
	files = p.overrideComposeFiles(files)
	for i, file := range files {
		c, err := p.RenderTemplate(string(file), nil)
		if err != nil {
			return nil, err
		}

		files[i] = []byte(c)
	}

	return files, nil
}

func (p *Project) overrideComposeFiles(files [][]byte) [][]byte {
	file := os.Getenv(EnvComposeFile)
	if file == "" {
		return files
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		logrus.Fatal(err)
	}

	return [][]byte{content}
}

func (p *Project) Install(ctx context.Context, opts *InstallOptions) error {
	return p.c.Install.Run(p, p.c, func(*Project, *Config) error {
		return p.Compose.Up(ctx, options.Up{})
	}, opts.NoExec)
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

func (p *Project) Start(ctx context.Context, opts *StartOptions) error {
	return p.c.Start.Run(p, p.c, func(*Project, *Config) error {
		if err := p.isInstalled(ctx); err != nil {
			return err
		}

		return p.Compose.Start(ctx)
	}, opts.NoExec)
}

func (p *Project) Stop(ctx context.Context, opts *StopOptions) error {
	return p.c.Stop.Run(p, p.c, func(*Project, *Config) error {
		if err := p.isInstalled(ctx); err != nil {
			return err
		}

		return p.Compose.Stop(ctx, 0)
	}, opts.NoExec)
}

func (p *Project) Uninstall(ctx context.Context, opts *UninstallOptions) error {
	return p.c.Uninstall.Run(p, p.c, func(*Project, *Config) error {
		if !opts.Force {
			if err := p.isInstalled(ctx); err != nil {
				return err
			}
		}

		if err := p.Compose.Stop(ctx, 0); err != nil {
			return err
		}

		d := options.Down{}
		if opts.Purge {
			d.RemoveImages = "all"
			d.RemoveVolume = true
		}

		return p.Compose.Down(ctx, d)
	}, opts.NoExec)
}

func (p *Project) Status(ctx context.Context) error {
	info, err := p.Compose.Ps(ctx)
	fmt.Println(info.String([]string{"Name", "Command", "State", "Ports"}, true))
	return err
}

func (p *Project) Execute(ctx context.Context, service string, cmd ...string) error {
	srv, err := p.Compose.CreateService(service) //, options.Run{Detached: true})
	if err != nil {
		return err
	}

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

func (p *Project) MustRenderTemplate(tmpl string, vars map[string]interface{}) string {
	v, err := p.RenderTemplate(tmpl, vars)
	if err != nil {
		panic(err)
	}

	return v
}

func (p *Project) RenderTemplate(tmpl string, vars map[string]interface{}) (string, error) {
	t, err := template.New("compose").Parse(tmpl)
	if err != nil {
		return "", err
	}

	if vars == nil {
		vars = map[string]interface{}{}
	}

	vars["Project"] = p.c.ProjectName
	vars["Home"], _ = homedir.Dir()
	vars["OS"] = runtime.GOOS
	vars["Arch"] = runtime.GOARCH

	for k, v := range p.c.TemplateVars {
		vars[k] = v
	}

	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, vars)

	return buf.String(), err
}

func ErrorToMap(err error) map[string]interface{} {
	return map[string]interface{}{"Error": err.Error()}
}
