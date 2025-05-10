package cmd

import "github.com/spf13/cobra"

func RssCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "rss [command]",
		Short: "Manage RSS",
	}

	cmd.AddCommand(RssFeed())
	cmd.AddCommand(RssRule())
	cmd.AddCommand(RssSub())

	return cmd
}
