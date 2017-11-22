package pkgr

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	ProjectName string
	Compose     [][]byte

	Install   Operation
	Uninstall Operation
	Start     Operation
	Stop      Operation
}

type Operation struct {
	Messages Messages
	Execs    []*Exec
	Wrapper  Wrapper
}

func (c *Operation) Run(p *Project, cfg *Config, w Wrapper) error {
	logrus.Warn(c.Messages.Announcement)
	if err := w(p, cfg, nil); err != nil {
		logrus.Error(c.Messages.Failure)
		return err
	}

	logrus.Warn(c.Messages.Success)
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

type Wrapper func(*Project, *Config, Wrapper) error
