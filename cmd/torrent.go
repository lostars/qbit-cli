package cmd

import "github.com/spf13/cobra"

func TorrentCmd() *cobra.Command {
	torrentCmd := &cobra.Command{
		Use:   "torrent [commands]",
		Short: "Torrent sub commands",
	}

	torrentCmd.AddCommand(TorrentAdd())
	torrentCmd.AddCommand(TorrentList())
	torrentCmd.AddCommand(TorrentFiles())

	return torrentCmd
}
