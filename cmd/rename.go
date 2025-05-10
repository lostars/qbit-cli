package cmd

import (
	"github.com/spf13/cobra"
)

func RenameCmd() *cobra.Command {

	renameCmd := &cobra.Command{
		Use:   "rename [commands]",
		Short: "Rename tools for qBittorrent",
	}

	renameCmd.AddCommand(RenameJP())

	return renameCmd
}
