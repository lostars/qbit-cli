package cmd

import (
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"sort"
	"strings"
)

func JobCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "job [command]",
		Short: "Job management",
	}

	cmd.AddCommand(RunJob())
	cmd.AddCommand(JobList())

	return cmd
}

func JobList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "Job list",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		jobs := api.ListJobs()
		header := []string{"name", "tags", "description"}
		data := make([][]string, len(jobs))
		sort.Slice(jobs, func(i, j int) bool {
			return jobs[i].JobName() < jobs[j].JobName()
		})
		for _, v := range jobs {
			description := ""
			if d, ok := v.(api.Description); ok {
				description = d.Description()
			}
			var tags []string
			if t, ok := v.(api.Tag); ok {
				tags = t.Tags()
			}
			data = append(data, []string{v.JobName(), strings.Join(tags, ","), description})
		}

		utils.PrintListWithColWidth(header, &data, map[int]int{2: 50}, true)

		return nil
	}

	return cmd
}

func RunJob() *cobra.Command {
	var run = &cobra.Command{
		Use:   "run [command]",
		Short: "Run job",
	}

	jobs := api.ListJobs()
	for _, job := range jobs {
		run.AddCommand(job.RunCommand())
	}

	return run
}
