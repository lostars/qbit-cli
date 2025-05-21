package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
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
	torrentCmd.AddCommand(TorrentFilePriority())
	torrentCmd.AddCommand(TorrentTracker())
	torrentCmd.AddCommand(TorrentPeer())

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

func TorrentFilePriority() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "fp <hash>",
		Short: "Set torrent file priority",
		Long: `Make sure your qBittorrent webapi version is >= 2.8.2
This command use file index which is return by torrent files from webapi 2.8.2`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("torrent hash is required")
			}
			return nil
		},
	}
	var index string
	var priority int

	cmd.Flags().StringVar(&index, "index", "", "index of torrent file, separated by |")
	cmd.Flags().IntVar(&priority, "priority", 0, `priority of torrent file:
0	Do not download
1	Normal priority
6	High priority
7	Maximal priority`)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if index == "" {
			return errors.New("torrent file index is required")
		}
		err := api.SetTorrentFilePriority(args[0], index, priority)
		if err != nil {
			return err
		}
		return nil
	}

	return cmd
}

func TorrentTracker() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "tracker <hash>",
		Short: "List torrent trackers",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("torrent hash is required")
			}
			return nil
		},
	}

	var status int
	cmd.Flags().IntVar(&status, "status", -1, `status of tracker:
0	Tracker is disabled (used for DHT, PeX, and LSD)
1	Tracker has not been contacted yet
2	Tracker has been contacted and is working
3	Tracker is updating
4	Tracker has been contacted, but it is not working (or doesn't send proper replies)
`)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		trackers, err := api.TorrentTrackers(args[0])
		if err != nil {
			return err
		}
		header := []string{"url", "status", "tier", "peers", "seeds", "leeches", "downloaded", "msg"}
		var data = make([][]string, 0, len(*trackers))
		for _, t := range *trackers {
			if status >= 0 {
				if status == t.Status {
					data = append(data, []string{t.URL, strconv.Itoa(t.Status), strconv.Itoa(t.Tier), strconv.Itoa(t.NumPeers),
						strconv.Itoa(t.NumSeeds), strconv.Itoa(t.NumLeeches), strconv.Itoa(t.NumDownloaded), t.Msg})
				}
			} else {
				data = append(data, []string{t.URL, strconv.Itoa(t.Status), strconv.Itoa(t.Tier), strconv.Itoa(t.NumPeers),
					strconv.Itoa(t.NumSeeds), strconv.Itoa(t.NumLeeches), strconv.Itoa(t.NumDownloaded), t.Msg})
			}
		}
		utils.PrintList(header, &data)
		return nil
	}

	return cmd
}

func TorrentPeer() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "peer <hash>",
		Short: "Get torrent peers",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("torrent hash is required")
			}
			return nil
		},
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		peers, err := api.TorrentPeers(args[0])
		if err != nil {
			return err
		}

		header := []string{"area", "host", "conn", "flags", "client", "PROG", "DLS", "UPS", "DL", "UL", "REL", "files"}
		var data = make([][]string, 0, len(*peers))
		for _, t := range *peers {
			data = append(data, []string{t.CountryCode, t.IP + ":" + strconv.Itoa(t.Port), t.Connection, t.Flags, t.Client,
				utils.FormatPercent(t.Progress), utils.FormatFileSizeAuto(uint64(t.DLSpeed), 0) + "/s",
				utils.FormatFileSizeAuto(uint64(t.UpSpeed), 0) + "/S", utils.FormatFileSizeAuto(uint64(t.Downloaded), 0),
				utils.FormatFileSizeAuto(uint64(t.Uploaded), 0), utils.FormatPercent(t.Relevance), t.Files})
		}
		utils.PrintListWithColWidth(header, &data, map[int]int{1: 20, 4: 10}, true)
		return nil
	}

	return cmd
}
