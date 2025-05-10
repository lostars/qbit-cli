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
		Use:   "add <torrent url> ... [flags]",
		Short: "Add torrent, you can add one or more torrents seperated by blank space",
		Long: `You can set default save values in config file to save your time.
Attention: auto management is enabled by default, so make sure your qBittorrent if configured properly.
`,
		Example: `qbit torrent add 'magnet:xxxx' 'xx' --category=abc --tags=a,b,c --auto-management --save-path=/ab/c`,
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
	addCmd.Flags().BoolVar(&autoTMM, "auto-management", true, "Whether Automatic Torrent Management should be used, default is true")
	addCmd.Flags().StringVar(&savePath, "save-path", "", "torrent save path")

	addCmd.RunE = func(cmd *cobra.Command, args []string) error {

		cfg, _ := config.GetConfig()
		// load defaults from config file
		if category == "" && cfg != nil {
			category = cfg.Torrent.DefaultSaveCategory
		}
		if tags == "" && cfg != nil {
			tags = cfg.Torrent.DefaultSaveTags
		}
		if savePath == "" && cfg != nil {
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
