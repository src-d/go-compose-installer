package pkgr

import (
	"context"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

type Program struct {
	Parser *flags.Parser
}

func NewProgram(name string, cfg *Config) (*Program, error) {
	p, err := NewProject(cfg)
	if err != nil {
		return nil, err
	}

	parser := flags.NewNamedParser(name, flags.Default)
	parser.AddCommand("install",
		fmt.Sprintf(cfg.Install.Messages.Description, cfg.ProjectName), "",
		&InstallCommand{Command: Command{p: p}},
	)
	parser.AddCommand("start",
		fmt.Sprintf(cfg.Start.Messages.Description, cfg.ProjectName), "",
		&StartCommand{Command: Command{p: p}},
	)
	parser.AddCommand("stop",
		fmt.Sprintf(cfg.Stop.Messages.Description, cfg.ProjectName), "",
		&StopCommand{Command: Command{p: p}},
	)
	parser.AddCommand("status",
		fmt.Sprintf(cfg.Status.Messages.Description, cfg.ProjectName), "",
		&StatusCommand{Command: Command{p: p}},
	)
	parser.AddCommand("uninstall",
		fmt.Sprintf(cfg.Uninstall.Messages.Description, cfg.ProjectName), "",
		&UninstallCommand{Command: Command{p: p}},
	)
	return &Program{parser}, nil
}

func (p *Program) Run() {
	if _, err := p.Parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			p.Parser.WriteHelp(os.Stdout)
			//fmt.Printf("\nBuild information\n  commit: %s\n  date: %s\n", version, build)
			os.Exit(1)
		}
	}
}

type Command struct {
	Debug bool `long:"debug" description:"enables the debug mode."`
	p     *Project
}

func (c *Command) Execute([]string) error {
	logrus.SetFormatter(&LogFormatter{})
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	return nil
}

type InstallCommand struct {
	Command
}

func (c *InstallCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Install(context.Background())
}

type StartCommand struct {
	Command
}

func (c *StartCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Start(context.Background())
}

type StopCommand struct {
	Command
}

func (c *StopCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Stop(context.Background())
}

type StatusCommand struct {
	Command
}

func (c *StatusCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Status(context.Background())
}

type UninstallCommand struct {
	Purge bool `long:"purge" description:"remove the docker images and volumes."`
	Command
}

func (c *UninstallCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Uninstall(context.Background(), c.Purge)
}
