package pkgr

import (
	"bytes"

	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/sirupsen/logrus"
)

type Listener struct {
	project    *project.Project
	listenChan chan events.Event
	upCount    int
}

// NewListener create a default listener for the specified project.
func NewListener(p *project.Project) chan<- events.Event {
	l := Listener{
		listenChan: make(chan events.Event),
		project:    p,
	}
	go l.start()
	return l.listenChan
}

func (d *Listener) start() {
	for event := range d.listenChan {
		d.log(&event)
	}
}

func (d *Listener) log(e *events.Event) {
	buffer := bytes.NewBuffer(nil)
	if e.Data != nil {
		for k, v := range e.Data {
			if buffer.Len() > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(k)
			buffer.WriteString("=")
			buffer.WriteString(v)
		}
	}

	if e.EventType == events.ServiceUp {
		d.upCount++
	}

	if e.ServiceName == "" {
		logrus.Debugf("Project [%s]: %s %s", d.project.Name, e.EventType, buffer.Bytes())
		return
	}

	logrus.Debugf("[%d/%d] [%s]: %s %s", d.upCount, d.project.ServiceConfigs.Len(), e.ServiceName, e.EventType, buffer.Bytes())

}
