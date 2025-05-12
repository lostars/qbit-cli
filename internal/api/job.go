package api

import (
	"fmt"
	"github.com/spf13/cobra"
)

type Job interface {
	JobName() string
	RunCommand() *cobra.Command
}

type Description interface {
	Description() string
}

type Tag interface {
	Tags() []string
}

var (
	jobs = make(map[string]Job, 10)
)

func RegisterJob(job Job) {
	if j := jobs[job.JobName()]; j != nil {
		panic(fmt.Sprintf("job %s already registered!\n", job.JobName()))
	}
	jobs[job.JobName()] = job
}

func ListJobs() []Job {
	var list []Job
	for _, v := range jobs {
		list = append(list, v)
	}
	return list
}
