package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"regexp"
	"strconv"
)

func JackettSearch() *cobra.Command {
	var searchCmd = &cobra.Command{
		Use:   "search <keyword>",
		Short: "Search torrents through Jackett",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires at least one arg")
			}
			return nil
		},
	}

	var autoDownload, autoMM bool
	var savePath, saveCategory, saveTags string
	var indexer, torrentRegex string
	var jsonFormat bool
	var category []string

	searchCmd.Flags().BoolVar(&autoDownload, "auto-download", false, "auto download")
	searchCmd.Flags().BoolVar(&autoMM, "auto-manage", true, "whether enable torrent auto management default is true, valid only when auto download enabled")
	searchCmd.Flags().StringVar(&savePath, "save-path", "", "save path")
	searchCmd.Flags().StringVar(&saveCategory, "save-category", "", "save category")
	searchCmd.Flags().StringVar(&saveTags, "save-tags", "", "save tags")

	searchCmd.Flags().StringVar(&indexer, "indexer", "all", "indexer")
	searchCmd.Flags().StringVar(&torrentRegex, "torrent-regex", "", "result title filter")
	searchCmd.Flags().BoolVar(&jsonFormat, "json", false, "display results as json format")
	searchCmd.Flags().StringSliceVar(&category, "category", []string{}, "category")

	searchCmd.RunE = func(cmd *cobra.Command, args []string) error {
		result, err := api.JackettSearch(indexer, category, args[0])
		if err != nil {
			return err
		}

		var re *regexp.Regexp
		if torrentRegex != "" {
			r, err := regexp.Compile(torrentRegex)
			if err != nil {
				fmt.Printf("regex: %s compile failed\n", torrentRegex)
			} else {
				re = r
			}
		}
		downloadList := make([]*api.JackettResult, 0, len(*result.Results))
		for _, t := range *result.Results {
			if re != nil {
				if re.MatchString(t.Title) {
					downloadList = append(downloadList, &t)
				}
			} else {
				downloadList = append(downloadList, &t)
			}
		}

		if autoDownload {
			if len(downloadList) > 0 {
				var d = make([]string, len(downloadList))
				for i, t := range downloadList {
					url := t.MagnetUri
					if url == "" {
						url = t.Link
					}
					d[i] = url
				}
				AutoDownload(d, savePath, saveCategory, saveTags, autoMM)
			} else {
				fmt.Println("no results found")
			}
		} else {
			if jsonFormat {
				data, err := json.MarshalIndent(downloadList, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else {
				header := []string{"tracker", "title", "size", "category", "S", "L"}
				var data = make([][]string, 0, len(downloadList))
				for _, r := range downloadList {
					data = append(data, []string{r.TrackerId, r.Title, utils.FormatFileSizeAuto(uint64(r.Size), 1),
						r.CategoryDesc, strconv.FormatInt(int64(r.Seeders), 10), strconv.FormatInt(int64(r.Peers), 10)})
				}
				utils.PrintListWithColWidth(header, &data, map[int]int{1: 50}, false)
			}
		}

		return nil
	}

	return searchCmd
}
