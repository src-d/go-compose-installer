package installer

import (
	"context"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&LogFormatter{})
}

type Installer struct {
	Parser *flags.Parser
}

func New(name string, cfg *Config) (*Installer, error) {
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

	return &Installer{parser}, nil
}

func (p *Installer) Run() {
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
	Debug  bool `long:"debug" description:"enables the debug mode."`
	NoExec bool `long:"no-exec" description:"disable any execution after the action."`

	p *Project
}

func (c *Command) Execute([]string) error {
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
	return c.p.Install(context.Background(), &InstallOptions{
		NoExec: c.NoExec,
	})
}

type StartCommand struct {
	Command
}

func (c *StartCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Start(context.Background(), &StartOptions{
		NoExec: c.NoExec,
	})
}

type StopCommand struct {
	Command
}

func (c *StopCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Stop(context.Background(), &StopOptions{
		NoExec: c.NoExec,
	})
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
	Force bool `long:"force" description:"force the uninstall process."`
	Command
}

func (c *UninstallCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Uninstall(context.Background(), &UninstallOptions{
		NoExec: c.NoExec,
		Purge:  c.Purge,
		Force:  c.Force,
	})
}
