package emby

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/internal/config"
	"strconv"
	"strings"
	"time"
)

type JP4K struct {
}

func (j *JP4K) JobName() string {
	return "jp4k"
}

func (j *JP4K) Tags() []string {
	return []string{"Emby", "qBittorrent"}
}

func (j *JP4K) Description() string {
	return `Based on your video metadata that contains 4k tag, but source file is not 4k resolution(width<3000).
It will search through qBittorrent plugin and download the 4k version(based on torrent file name that contains 4k keyword).
You better set plugins flag to run faster.`
}

func init() {
	api.RegisterJob(&JP4K{})
}

func (j *JP4K) RunCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   j.JobName(),
		Short: "Replace your jp video with 4k version.",
		Long:  j.Description(),
	}

	var (
		plugins                          string
		saveCategory, savePath, saveTags string
		autoMM                           bool
	)

	cmd.Flags().StringVar(&plugins, "plugins", "enabled", "which plugin to use, comma separated list of plugin names")
	cmd.Flags().BoolVar(&autoMM, "auto-management", true, "whether enable torrent auto management")

	cmd.Flags().StringVar(&saveCategory, "save-category", "", "torrent save category")
	cmd.Flags().StringVar(&savePath, "save-path", "", "torrent save path")
	cmd.Flags().StringVar(&saveTags, "save-tags", "", "torrent save tags")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		items := fourKItems()
		if items == nil || len(items) <= 0 {
			fmt.Println("no 4k items found")
			return nil
		}

		data := make([]*api.SearchDetail, 0, len(items)*2)
		for _, item := range items {
			// parse code from name
			matches := api.JPCodeRegex.FindStringSubmatch(item.Name)
			if len(matches) < 2 {
				continue
			}
			jpCode := matches[1]
			fmt.Printf("looking 4k version of %s...\n", jpCode)

			params := url.Values{
				"pattern":  {jpCode},
				"category": {"all"},
			}
			if plugins != "" {
				params.Set("plugins", plugins)
			}
			result, err := api.SearchStart(params)
			if err != nil {
				fmt.Printf("search start error: %s\n", err)
				continue
			}
			results, err := api.SearchDetails(1*time.Second, result.ID, params)
			if err != nil {
				fmt.Printf("search details error: %s\n", err)
				continue
			}

			for _, r := range results {
				if api.JP4KRegex.MatchString(r.FileName) {
					data = append(data, &r)
				}
			}
		}

		if len(data) > 0 {
			urls := make([]string, 0, len(data))
			fmt.Println("founded 4k items:")
			for _, item := range data {
				urls = append(urls, item.FileURL)
				fmt.Println(item.FileName)
			}
			err := addTorrents(urls, autoMM, saveCategory, saveTags, savePath)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return cmd
}

func fourKItems() []*api.EmbyItem {
	// search 4k tag first
	tagParams := url.Values{
		"SearchTerm":       {"4k"},
		"IncludeItemTypes": {"genre"},
		"Recursive":        {"true"},
	}
	tags, err := Items(tagParams)
	if err != nil {
		return nil
	}
	tag := tags.Items[0].ID

	params := url.Values{}
	params.Add("Recursive", "true")
	params.Add("GenreIds", tag)
	params.Add("MaxWidth", strconv.Itoa(3000))

	items, err := Items(params)
	if err != nil {
		return nil
	}
	results := make([]*api.EmbyItem, 0, len(items.Items))
	for _, item := range items.Items {
		results = append(results, &item)
	}
	return results
}

func addTorrents(urls []string, autoTMM bool, category, tags, savePath string) error {
	params := url.Values{
		"urls":    {strings.Join(urls, "\n")},
		"autoTMM": {strconv.FormatBool(autoTMM)},
	}

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

	if err := api.TorrentAdd(params); err != nil {
		return err
	}
	return nil
}
