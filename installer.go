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

// Installer represent the CLI application based on `go-flags` providing all the
// core commands: `install`, `start`, `stop` and `uninstall`.
type Installer struct {
	Parser *flags.Parser
}

// New returns a new Installer instance based on the given name and *Config.
func New(name string, cfg *Config) (*Installer, error) {
	p, err := NewProject(cfg)
	if err != nil {
		return nil, err
	}

	parser := flags.NewNamedParser(name, flags.Default)
	parser.AddCommand("install",
		p.MustRenderTemplate(cfg.Install.Messages.Description, nil), "",
		&InstallCommand{Command: Command{p: p}},
	)
	parser.AddCommand("start",
		p.MustRenderTemplate(cfg.Start.Messages.Description, nil), "",
		&StartCommand{Command: Command{p: p}},
	)
	parser.AddCommand("stop",
		p.MustRenderTemplate(cfg.Stop.Messages.Description, nil), "",
		&StopCommand{Command: Command{p: p}},
	)
	parser.AddCommand("status",
		p.MustRenderTemplate(cfg.Status.Messages.Description, nil), "",
		&StatusCommand{Command: Command{p: p}},
	)
	parser.AddCommand("uninstall",
		p.MustRenderTemplate(cfg.Uninstall.Messages.Description, nil), "",
		&UninstallCommand{Command: Command{p: p}},
	)

	return &Installer{parser}, nil
}

// Run executes the CLI application, in a standard usage this function should be
// called form a main.main function. It `os.Exit` after be executed.
func (p *Installer) Run() {
	if _, err := p.Parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			p.Parser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

// Command defines a the base CLI command. It implements the non-interface for
// a command, for go-flags, where a `Execute([]string) error` is expected.
// The flags defined here are common to any other command.
type Command struct {
	Debug  bool `long:"debug" description:"enables the debug mode."`
	NoExec bool `long:"no-exec" description:"disable any execution after the action."`

	p *Project
}

// Execute execute the command, to be shadowed by specific implementations.
func (c *Command) Execute([]string) error {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	return nil
}

// InstallCommand defines the "install" command from the installer.
type InstallCommand struct {
	Command
}

func (c *InstallCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Install(context.Background(), &InstallOptions{
		NoExec: c.NoExec,
	})
}

// StartCommand defines the "start" command from the installer.
type StartCommand struct {
	Command
}

func (c *StartCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Start(context.Background(), &StartOptions{
		NoExec: c.NoExec,
	})
}

// StopCommand defines the "stop" command from the installer.
type StopCommand struct {
	Command
}

func (c *StopCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Stop(context.Background(), &StopOptions{
		NoExec: c.NoExec,
	})
}

// StatusCommand defines the "status" command from the installer.
type StatusCommand struct {
	Command
}

func (c *StatusCommand) Execute([]string) error {
	c.Command.Execute(nil)
	return c.p.Status(context.Background())
}

// UninstallCommand defines the "uninstall" command from the installer.
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
