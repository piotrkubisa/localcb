package ci

import (
	"context"

	"github.com/docker/docker/client"
)

type Pipeline struct {
	Context context.Context
	Client  *client.Client

	Stages []*Stage
}

func NewPipeline() (*Pipeline, error) {
	ctx := context.Background()

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	p := &Pipeline{
		Context: ctx,
		Client:  cli,
	}
	return p, nil
}

func (p *Pipeline) AddStage(stage *Stage) {
	p.Stages = append(p.Stages, stage)
}

type Stage struct {
	Name     string
	Commands []Command
}

func NewStage(name string, cmds []Command) *Stage {
	return &Stage{
		Name:     name,
		Commands: cmds,
	}
}

type Command struct {
	Exec string
}

func NewCommand(cmd string) Command {
	return Command{cmd}
}
