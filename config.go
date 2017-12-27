package installer

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

func NewDefaultConfig() *Config {
	cfg := &Config{}
	cfg.Install.Messages = Messages{
		Description:  "Installs %s into your system.",
		Announcement: "Installing %s...",
		Failure:      "Failed to install %s: %s.",
		Success:      "%s was successfully installed.",
	}

	cfg.Start.Messages = Messages{
		Description:  "Stars %s.",
		Announcement: "Starting %s...",
		Failure:      "Failed to start %s: %s.",
		Success:      "%s started.",
	}

	cfg.Stop.Messages = Messages{
		Description:  "Stops %s.",
		Announcement: "Stopping %s...",
		Failure:      "Failed to stop %s: %s.",
		Success:      "%s stopped.",
	}

	cfg.Status.Messages = Messages{
		Description: "Show the status of %s.",
	}

	cfg.Uninstall.Messages = Messages{
		Description:  "Remove %s from your system.",
		Announcement: "Uninstalling %s...",
		Failure:      "Failed to uninstall %s: %s.",
		Success:      "%s was successfully uninstalled.",
	}

	return cfg
}

type Config struct {
	ProjectName string
	Compose     [][]byte

	Install   Operation
	Uninstall Operation
	Status    Operation
	Start     Operation
	Stop      Operation
}

type Operation struct {
	Messages Messages
	Execs    []*Exec
}

var DefaultShell = []string{"/bin/sh", "-c"}

func (c *Operation) Run(p *Project, cfg *Config, a Action, noExec bool) error {
	logrus.Infof(c.Messages.Announcement, cfg.ProjectName)
	if err := a(p, cfg); err != nil {
		logrus.Fatalf(c.Messages.Failure, cfg.ProjectName, err)
		return err
	}

	if !noExec {
		if err := c.executeExec(p); err != nil {
			logrus.Fatalf(c.Messages.Failure, cfg.ProjectName, err)
			return err
		}
	}

	logrus.Infof(c.Messages.Success, cfg.ProjectName)
	return nil
}

func (c *Operation) executeExec(p *Project) error {
	for _, e := range c.Execs {
		if err := p.Execute(context.Background(), e.Service, "sh", "-c", e.Cmd); err != nil {
			return fmt.Errorf("error executing %q in %s:%s", e.Cmd, e.Service, err)
		}

	}

	return nil
}

type Messages struct {
	Description  string
	Announcement string
	Failure      string
	Success      string
}

type Exec struct {
	Service string
	Cmd     string
}

type Action func(*Project, *Config) error
