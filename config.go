package installer

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// NewDefaultConfig returns a Config object will all Messages already configured
// with standard text suitable for any application.
func NewDefaultConfig() *Config {
	cfg := &Config{}
	cfg.Install.Messages = Messages{
		Description:  "Installs {{.Project}} into your system.",
		Announcement: "Installing {{.Project}}...",
		Failure:      "Failed to install {{.Project}}: {{.Error}} .",
		Success:      "{{.Project}} was successfully installed.",
	}

	cfg.Start.Messages = Messages{
		Description:  "Stars {{.Project}}.",
		Announcement: "Starting {{.Project}}...",
		Failure:      "Failed to start {{.Project}}: {{.Error}}.",
		Success:      "{{.Project}} started.",
	}

	cfg.Stop.Messages = Messages{
		Description:  "Stops {{.Project}}.",
		Announcement: "Stopping {{.Project}}...",
		Failure:      "Failed to stop {{.Project}}: {{.Error}}.",
		Success:      "{{.Project}} stopped.",
	}

	cfg.Status.Messages = Messages{
		Description: "Show the status of {{.Project}}.",
	}

	cfg.Uninstall.Messages = Messages{
		Description:  "Remove {{.Project}} from your system.",
		Announcement: "Uninstalling {{.Project}}...",
		Failure:      "Failed to uninstall {{.Project}}: {{.Error}}.",
		Success:      "{{.Project}} was successfully uninstalled.",
	}

	return cfg
}

// Config contains all the messages and exec operations for each command.
type Config struct {
	// ProjectName is the name given to the project being installed.
	ProjectName string
	// Compose is the content of one or more docker compose files in YAML format.
	Compose [][]byte
	// TemplateVars are custom defined variable to be replace on the values
	// supporting templates, such as docker compose files or text messages.
	TemplateVars map[string]interface{}
	// Install operation configuration.
	Install Operation
	// Uninstall operation configuration.
	Uninstall Operation
	// Status operation configuration.
	Status Operation
	// Start operation configuration.
	Start Operation
	// Stop operation configuration.
	Stop Operation
}

type Operation struct {
	Messages Messages
	Execs    []*Exec
}

var DefaultShell = []string{"/bin/sh", "-c"}

func (c *Operation) Run(p *Project, cfg *Config, a Action, noExec bool) error {
	logrus.Info(p.MustRenderTemplate(c.Messages.Announcement, nil))
	if err := a(p, cfg); err != nil {
		logrus.Fatal(p.MustRenderTemplate(c.Messages.Failure, ErrorToMap(err)))
		return err
	}

	if !noExec {
		if err := c.executeExec(p); err != nil {
			logrus.Fatal(p.MustRenderTemplate(c.Messages.Failure, ErrorToMap(err)))
			return err
		}
	}

	logrus.Info(p.MustRenderTemplate(c.Messages.Success, nil))
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

// Messages are the message to be printed after, before, etc, every command.
type Messages struct {
	// Description to be shown in the help of the command.
	Description string
	// Announcement information shown before execute the command.
	Announcement string
	// Failure is the text to be shown just after an error happend.
	Failure string
	// Success in case of a successfully execution.
	Success string
}

type Exec struct {
	Service string
	Cmd     string
}

type Action func(*Project, *Config) error
