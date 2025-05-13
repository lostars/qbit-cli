package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"strconv"
	"strings"
)

func TorrentUpdate() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "update <hash>",
		Short: "A bulk of torrent operations, support multiple or all torrents.",
	}

	var (
		all                                                      bool
		stop, start, recheck, reannounce                         bool
		increasePriority, decreasePriority                       bool
		maximalPriority, minimalPriority                         bool
		downloadLimit, uploadLimit                               int64
		category, tags, torrentLocation, removeTags              string
		autoManage, forceStart, superSeeding, sequentialDownload bool
		firstOrLastPieceFirst                                    bool
	)

	cmd.Flags().BoolVar(&all, "all", false, "update all torrents")

	cmd.Flags().BoolVar(&stop, "stop", false, "stop torrent")
	cmd.Flags().BoolVar(&start, "start", false, "start torrent")
	cmd.Flags().BoolVar(&recheck, "recheck", false, "recheck torrent")
	cmd.Flags().BoolVar(&reannounce, "reannounce", false, "reannounce torrent")

	cmd.Flags().BoolVar(&increasePriority, "increase-priority", false, "increase torrent priority")
	cmd.Flags().BoolVar(&decreasePriority, "decrease-priority", false, "decrease torrent priority")

	cmd.Flags().BoolVar(&maximalPriority, "maximal-priority", false, "maximal torrent priority")
	cmd.Flags().BoolVar(&minimalPriority, "minimal-priority", false, "minimal torrent priority")

	cmd.Flags().Int64Var(&downloadLimit, "download-limit", 0, "download limit(bytes)")
	cmd.Flags().Int64Var(&uploadLimit, "upload-limit", 0, "upload limit(bytes)")

	cmd.Flags().StringVar(&category, "category", "", "torrent category")
	cmd.Flags().StringVar(&tags, "tags", "", "torrent tags, separated by comma")
	cmd.Flags().StringVar(&removeTags, "remove-tags", "", "torrent tags to remove")
	cmd.Flags().StringVar(&torrentLocation, "torrent-location", "", "torrent location")

	cmd.Flags().BoolVar(&superSeeding, "super-seeding", false, "super seeding")
	cmd.Flags().BoolVar(&sequentialDownload, "sequential-download", false, "sequential download")
	cmd.Flags().BoolVar(&forceStart, "force-start", false, "force start torrent")
	cmd.Flags().BoolVar(&autoManage, "auto-manage", false, "auto manage torrent")

	cmd.Flags().BoolVar(&firstOrLastPieceFirst, "first-last-first", false, "first or last piece first")

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

		if stop {
			update("stop", params)
		}
		if start {
			update("start", params)
		}
		if recheck {
			update("recheck", params)
		}
		if reannounce {
			update("reannounce", params)
		}

		if increasePriority {
			update("increasePrio", params)
		}
		if decreasePriority {
			update("decreasePrio", params)
		}

		if maximalPriority {
			update("topPrio", params)
		}
		if minimalPriority {
			update("bottomPrio", params)
		}

		if downloadLimit > 0 {
			update("setDownloadLimit", params)
		}
		if uploadLimit > 0 {
			update("setUploadLimit", params)
		}

		if category != "" {
			params.Set("category", category)
			update("setCategory", params)
		}
		if tags != "" {
			params.Set("tags", tags)
			update("addTags", params)
		}
		if removeTags != "" {
			params.Set("tags", removeTags)
			update("removeTags", params)
		}
		if torrentLocation != "" {
			params.Set("location", torrentLocation)
			update("setLocation", params)
		}

		if autoManage {
			params.Set("enable", strconv.FormatBool(autoManage))
			update("setAutoManagement", params)
		}
		if sequentialDownload {
			update("toggleSequentialDownload", params)
		}
		if forceStart {
			update("setForceStart", params)
		}
		if firstOrLastPieceFirst {
			update("toggleFirstLastPiecePrio", params)
		}
		if superSeeding {
			update("toggleSuperSeeding", params)
		}

		return nil
	}

	return cmd
}

func update(operation string, params url.Values) {
	err := api.TorrentUpdate(operation, params)
	if err != nil {
		fmt.Printf("%s %s failed: %v\n", params.Get("hashes"), operation, err)
	} else {
		fmt.Println("done.")
	}
}
