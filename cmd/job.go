package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
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
		for _, v := range jobs {
			description := ""
			if d, ok := v.(api.Description); ok {
				description = d.Description()
			}
			var tags []string
			if t, ok := v.(api.Tag); ok {
				tags = t.Tags()
			}
			data = append(data, []string{v.JobName(), fmt.Sprintf("%v", tags), description})
		}

		c := tw.CellConfig{
			Formatting: tw.CellFormatting{
				MaxWidth:  100,
				AutoWrap:  tw.WrapNormal,
				Alignment: tw.AlignNone,
			},
		}
		utils.PrintListWithCellConfig(header, &data, c)

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
