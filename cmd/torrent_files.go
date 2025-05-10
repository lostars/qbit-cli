package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
)

func TorrentFiles() *cobra.Command {
	filesCmd := &cobra.Command{
		Use:     "files <hash>",
		Short:   "List torrent files by torrent hash",
		Example: `qbit torrent files <torrent hash>`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("torrent hash required")
			}
			return nil
		},
	}

	filesCmd.RunE = func(cmd *cobra.Command, args []string) error {
		hash := args[0]
		var params = url.Values{"hash": {hash}}

		torrentFiles := api.TorrentFiles(params)
		if torrentFiles == nil {
			return nil
		}

		fmt.Printf("total file size: %d\n", len(torrentFiles))
		for _, f := range torrentFiles {
			fmt.Printf("name: [%s] priority: [%d], progress: [%v]\n", f.Name, f.Priority, f.Progress)
		}

		return nil
	}

	return filesCmd
}
