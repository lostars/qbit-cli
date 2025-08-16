package job

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	c "qbit-cli/cmd"
	"qbit-cli/internal/api"
	"qbit-cli/internal/api/emby"
	"qbit-cli/internal/config"
	"strconv"
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
		savePath, saveTags            string
		autoMM                        bool
		premiereBefore, premiereAfter string
		extraCodes                    []string
	)

	saveCategory := c.FlagsProperty[string]{Flag: "save-category", Register: &c.TorrentCategoryFlagRegister{}}
	plugins := c.FlagsProperty[string]{Flag: "plugins", Register: &c.TorrentPluginsFlagRegister{}}

	cmd.Flags().StringVar(&premiereBefore, "premiere-before", "", "movie premiere before, ISO format")
	cmd.Flags().StringVar(&premiereAfter, "premiere-after", "", "movie premiere after, ISO format")

	cmd.Flags().StringVar(&plugins.Value, plugins.Flag, "enabled", "which plugin to use, comma separated list of plugin names")
	cmd.Flags().BoolVar(&autoMM, "auto-management", true, "whether enable torrent auto management")

	cmd.Flags().StringVar(&saveCategory.Value, saveCategory.Flag, "", "torrent save category")
	cmd.Flags().StringVar(&savePath, "save-path", "", "torrent save path")
	cmd.Flags().StringVar(&saveTags, "save-tags", "", "torrent save tags")

	cmd.Flags().StringSliceVar(&extraCodes, "extra-codes", []string{}, `comma separated list of extra jp code.
this option will find all jp video's 4k version which is filter by extra jp code.`)

	// register completion
	saveCategory.RegisterCompletion(cmd)
	plugins.RegisterCompletion(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		searchParams := url.Values{}
		searchParams.Add("Recursive", "true")
		// video definition filter
		searchParams.Add("MaxWidth", strconv.Itoa(3000))
		if premiereBefore != "" {
			searchParams.Add("MaxPremiereDate", premiereBefore)
		}
		if premiereAfter != "" {
			searchParams.Add("MinPremiereDate", premiereAfter)
		}

		items := fourKItems(searchParams)
		items = append(items, searchExtraCodes(extraCodes)...)
		if items == nil || len(items) <= 0 {
			fmt.Println("no 4k items found")
			return nil
		}

		data := make([]*api.SearchDetail, 0, len(items)*2)
		for _, item := range items {
			// parse code from name
			matches := JPCodeRegex.FindStringSubmatch(item.Name)
			if len(matches) < 2 {
				continue
			}
			jpCode := matches[1]
			fmt.Printf("looking 4k version of %s...\n", jpCode)

			params := url.Values{
				"pattern":  {jpCode},
				"category": {"all"},
			}
			if plugins.Value != "" {
				params.Set("plugins", plugins.Value)
			}
			result, err := api.SearchStart(params)
			if err != nil {
				fmt.Printf("search start error: %s\n", err)
				continue
			}
			results, err := api.SearchDetails(1*time.Second, result.ID)
			if err != nil {
				fmt.Printf("search details error: %s\n", err)
				continue
			}

			for _, r := range results {
				if JP4KRegex.MatchString(r.FileName) {
					data = append(data, r)
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
			err := addTorrents(urls, autoMM, saveCategory.Value, saveTags, savePath)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return cmd
}

func searchExtraCodes(codes []string) []*api.EmbyItem {
	// search extra codes which has no 4K tag
	searchParams := url.Values{
		"SearchTerm": codes,
		"Recursive":  {"true"},
		"Fields":     {"Genres"},
		// 4k video filter
		"MaxWidth": {strconv.Itoa(3000)},
	}
	items, err := emby.Items(searchParams)
	if err != nil {
		return nil
	}
	results := make([]*api.EmbyItem, 0, len(items.Items))
	for _, item := range items.Items {
		for _, g := range item.Genres {
			// escape 4k tag which is handled by #fourKItems
			if g == "4K" || g == "4k" {
				continue
			}
		}
		results = append(results, &item)
	}
	return results
}

func fourKItems(searchParams url.Values) []*api.EmbyItem {
	// search 4k tag first
	tagParams := url.Values{
		"SearchTerm":       {"4k"},
		"IncludeItemTypes": {"genre"},
		"Recursive":        {"true"},
	}
	tags, err := emby.Items(tagParams)
	if err != nil {
		return nil
	}
	if len(tags.Items) < 1 {
		return nil
	}
	tag := tags.Items[0].ID
	searchParams.Add("GenreIds", tag)

	items, err := emby.Items(searchParams)
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
		"autoTMM": {strconv.FormatBool(autoTMM)},
	}

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
	params.Add("tags", tags)
	params.Add("savepath", savePath)
	params.Add("category", category)

	if err := api.TorrentAdd(urls, params); err != nil {
		return err
	}
	return nil
}
