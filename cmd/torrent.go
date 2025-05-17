package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"strconv"
	"strings"
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
	torrentCmd.AddCommand(TorrentUpdate())
	torrentCmd.AddCommand(DeleteTorrents())
	torrentCmd.AddCommand(TagCmd())
	torrentCmd.AddCommand(TorrentCategoryCmd())

	return torrentCmd
}

func DeleteTorrents() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete <hashes>...",
		Short: "Delete torrents",
	}

	var deleteFiles, all bool
	cmd.Flags().BoolVar(&deleteFiles, "delete-files", false, "delete files")
	cmd.Flags().BoolVar(&all, "all", false, "delete all torrents")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		hashes := strings.Join(args, "|")
		if all {
			hashes = "all"
		} else {
			if len(args) < 1 {
				return errors.New("torrent update requires at least a hash")
			}
		}

		params := url.Values{}
		params.Set("hashes", hashes)
		params.Set("deleteFiles", strconv.FormatBool(all))
		err := api.UpdateTorrent("delete", params)
		if err != nil {
			return err
		}

		return nil
	}

	return cmd
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
