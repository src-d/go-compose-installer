package pkgr

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

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

func (c *Operation) Run(p *Project, cfg *Config, a Action) error {
	logrus.Warn(c.Messages.Announcement)
	if err := a(p, cfg); err != nil {
		logrus.Fatal(c.Messages.Failure)
		return err
	}

	if err := c.executeExec(p); err != nil {
		logrus.Fatal(err)
		return err
	}

	logrus.Warn(c.Messages.Success)
	return nil
}

func (c *Operation) executeExec(p *Project) error {
	for _, e := range c.Execs {
		if err := p.Execute(context.Background(), e.Service, e.Cmd...); err != nil {
			return fmt.Errorf("error executing %q in %s:%s", strings.Join(e.Cmd, " "), e.Service, err)
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
	Cmd     []string
}

type Action func(*Project, *Config) error
