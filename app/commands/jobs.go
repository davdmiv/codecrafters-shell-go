package commands

import (
	"fmt"

	"github.com/codecrafters-io/shell-starter-go/app/jobs"
)

// Jobs is the shell builtin `jobs`.
type Jobs struct {
	manager *jobs.Manager
}

func NewJobs(manager *jobs.Manager) *Jobs {
	return &Jobs{manager: manager}
}

func (j *Jobs) Name() string {
	return "jobs"
}

func (j *Jobs) Execute(ctx *Context, args []string) error {
	listings := j.manager.List()
	for _, item := range listings {
		if _, err := fmt.Fprintln(ctx.Stdout, jobs.FormatListing(item)); err != nil {
			return err
		}
	}
	// Bash removes finished jobs after they have been reported by `jobs`.
	j.manager.RemoveFinished()
	return nil
}
