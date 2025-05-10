package cmd

import "github.com/spf13/cobra"

func TorrentCmd() *cobra.Command {
	torrentCmd := &cobra.Command{
		Use:   "torrent [commands]",
		Short: "Manage torrents",
	}

	torrentCmd.AddCommand(TorrentAdd())
	torrentCmd.AddCommand(TorrentList())
	torrentCmd.AddCommand(TorrentFiles())
	torrentCmd.AddCommand(TorrentSearch())

	return torrentCmd
}
