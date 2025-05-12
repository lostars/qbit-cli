package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
)

func TorrentCmd() *cobra.Command {
	torrentCmd := &cobra.Command{
		Use:   "torrent [commands]",
		Short: "Manage torrents",
	}

	torrentCmd.AddCommand(TorrentAdd())
	torrentCmd.AddCommand(TorrentList())
	torrentCmd.AddCommand(TorrentFiles())
	torrentCmd.AddCommand(TorrentSearch())
	torrentCmd.AddCommand(RenameTorrentCmd())

	return torrentCmd
}

func RenameTorrentCmd() *cobra.Command {
	var torrentCmd = &cobra.Command{
		Use:   "rename <hash> <name>",
		Short: "Rename a torrent",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("need to provide hash and name")
			}
			return nil
		},
	}

	torrentCmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := api.RenameTorrent(args[0], args[1])
		if err != nil {
			return err
		}
		return nil
	}

	return torrentCmd
}
