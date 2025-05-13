package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/internal/config"
	"strconv"
	"strings"
)

func TorrentAdd() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add <torrent url>...",
		Short: "Add one or more torrent",
		Long: `You can set default save values in config file to save your time.
Attention: auto management is enabled by default, so make sure your qBittorrent if configured properly.
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

		cfg := config.GetConfig()
		// load defaults from config file
		if category == "" {
			category = cfg.Torrent.DefaultSaveCategory
		}
		if tags == "" {
			tags = cfg.Torrent.DefaultSaveTags
		}
		if savePath == "" {
			savePath = cfg.Torrent.DefaultSavePath
		}

		params := url.Values{
			"urls":    {strings.Join(args, "\n")},
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

		if err := api.TorrentAdd(params); err != nil {
			return err
		}

		return nil
	}

	return addCmd
}
