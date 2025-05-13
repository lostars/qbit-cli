package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/internal/config"
	"strconv"
)

func TorrentAdd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <torrent url>...",
		Short: "Add one or more torrent",
		Long: `You can set default save values in config file to save your time.
Attention: auto management is enabled by default, so make sure your qBittorrent if configured properly.
This method can add torrents from server local file or from URLs.
http://, https://, magnet: and bc://bt/ links are supported.
You can add torrent like: add /t/xx.torrent "magnet:xxx"
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires at least one torrent url")
			}
			return nil
		},
	}

	var (
		category string
		tags     string
		autoTMM  bool
		savePath string
	)
	addCmd.Flags().StringVar(&category, "category", "", "torrent category")
	addCmd.Flags().StringVar(&tags, "tags", "", "torrent tags split by ','")
	addCmd.Flags().BoolVar(&autoTMM, "auto-manage", true, "Whether Automatic Torrent Management should be used, default is true")
	addCmd.Flags().StringVar(&savePath, "save-path", "", "torrent save path")

	addCmd.RunE = func(cmd *cobra.Command, args []string) error {
		params := url.Values{
			"autoTMM": {strconv.FormatBool(autoTMM)},
		}
		if tags != "" {
			params.Add("tags", tags)
		}
		if category != "" {
			params.Add("category", category)
		}
		if savePath != "" && !autoTMM {
			params.Add("savepath", savePath)
		}
		LoadTorrentAddDefault(params)

		if err := api.TorrentAdd(args, params); err != nil {
			return err
		}

		return nil
	}

	return addCmd
}

func LoadTorrentAddDefault(params url.Values) {
	cfg := config.GetConfig()
	// load defaults from config file
	if params.Get("category") == "" {
		params.Set("category", cfg.Torrent.DefaultSaveCategory)
	}
	if params.Get("tags") == "" {
		params.Set("tags", cfg.Torrent.DefaultSaveCategory)
	}
	if params.Get("savepath") == "" {
		params.Set("savepath", cfg.Torrent.DefaultSavePath)
	}
}
