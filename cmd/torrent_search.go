package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"regexp"
	"strconv"
	"time"
)

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)#search

func TorrentSearch() *cobra.Command {

	var searchCmd = &cobra.Command{
		Use:   "search <keyword> [flags]",
		Short: "Search torrents through qBittorrent plugins",
		Long: `Be attention when you enable auto download,
and ensure that torrent-regex works properly to void unnecessary downloads.
Auto download calls "torrent add ...", which means it also reads default save values of torrent part on config file.
This list display formated json.
`,
		Example: `qbit torrent search <keyword> --category=movie --plugins=bt4g`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a keyword")
			}
			return nil
		},
	}

	var (
		category, plugins                              string
		autoDownload, autoMM                           bool
		torrentRegex, savePath, saveCategory, saveTags string
		interactive                                    bool
	)

	// search flags
	searchCmd.Flags().StringVar(&plugins, "plugins", "enabled", `plugins a|b|c, all and enabled also supported.
make sure you plugin is valid and enabled`)
	searchCmd.Flags().StringVar(&category, "category", "all", "category of plugin(define by plugin)")
	searchCmd.Flags().StringVar(&torrentRegex, "torrent-regex", "", "torrent file name filter")

	// auto download flags
	searchCmd.Flags().BoolVar(&autoDownload, "auto-download", false, "Attention: if true, it will auto download all the torrents that filter by torrent-regex")
	searchCmd.Flags().BoolVar(&autoMM, "auto-manage", true, "whether enable torrent auto management default is true, valid only when auto download enabled")
	// auto download save flags
	searchCmd.Flags().StringVar(&saveCategory, "save-category", "", "torrent save category, valid only when auto download enabled")
	searchCmd.Flags().StringVar(&savePath, "save-path", "", "torrent save path, valid only when auto download enabled")
	searchCmd.Flags().StringVar(&saveTags, "save-tags", "", "torrent save tags, valid only when auto download enabled")
	searchCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")

	searchCmd.RunE = func(cmd *cobra.Command, args []string) error {

		params := url.Values{
			"pattern": {args[0]},
		}

		if plugins != "" {
			params.Set("plugins", plugins)
		}
		if category != "" {
			params.Set("category", category)
		}

		result, err := api.SearchStart(params)
		if err != nil {
			return err
		}

		results, err := api.SearchDetails(1*time.Second, result.ID)
		if err != nil {
			return nil
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

		var urls []string
		var printList = make([]*api.SearchDetail, 0, len(results))
		for _, r := range results {
			if re == nil {
				printList = append(printList, r)
			} else {
				if re.MatchString(r.FileName) {
					printList = append(printList, r)
					urls = append(urls, r.FileURL)
				}
			}
		}

		if interactive {
			interactive = interactive && len(printList) > 0
			if interactive {
				header := []string{"name", "size", "S", "L", "plugin"}
				data := make([][]string, 0, len(printList))
				for _, r := range printList {
					data = append(data, []string{r.FileName, utils.FormatFileSizeAuto(uint64(r.FileSize), 1),
						strconv.FormatInt(int64(r.NBSeeders), 10), strconv.FormatInt(int64(r.NBLeechers), 10), r.EngineName})
				}
				model := utils.InteractiveTableModel{
					Rows:     &data,
					Header:   &header,
					WidthMap: map[int]int{0: 50, 1: 10, 2: 10, 3: 10, 4: 20},
					Delegate: &torrentSearchMsgDelegate{
						autoDownload, autoMM,
						savePath, saveCategory, saveTags,
						printList,
					},
				}
				if _, e := tea.NewProgram(&model, tea.WithAltScreen()).Run(); e != nil {
					return e
				}
			}
		} else {
			fmt.Printf("total search result size: %d\n", len(printList))
			str, err := json.MarshalIndent(printList, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(str))

			if autoDownload && len(printList) > 0 {
				var downloadList = make([]string, 0, len(printList))
				for _, r := range printList {
					downloadList = append(downloadList, r.FileURL)
				}
				AutoDownload(downloadList, savePath, saveCategory, saveTags, autoMM)
			}
		}

		return nil
	}

	return searchCmd
}

func AutoDownload(urls []string, savePath, saveCategory, saveTags string, autoMM bool) {
	addParams := url.Values{}
	addParams.Set("category", saveCategory)
	addParams.Set("tags", saveTags)
	addParams.Set("auto-manage", strconv.FormatBool(autoMM))
	addParams.Set("save-path", savePath)
	LoadTorrentAddDefault(addParams)
	if err := api.TorrentAdd(urls, addParams); err != nil {
		fmt.Println("auto download failed:", err)
	} else {
		fmt.Printf("auto download %d torrent(s) success\n", len(urls))
	}
}

type torrentSearchMsgDelegate struct {
	autoDownload, autoMM             bool
	savePath, saveCategory, saveTags string
	data                             []*api.SearchDetail
}

func (j *torrentSearchMsgDelegate) Operation(msg tea.KeyMsg, cursor int) tea.Cmd {
	switch msg.String() {
	case "enter":
		if j.data == nil || cursor >= len(j.data) {
			return nil
		}
		torrents := j.data[cursor].FileURL
		str := InteractiveDownload([]string{torrents}, j.savePath, j.saveCategory, j.saveTags, j.autoMM)
		return func() tea.Msg {
			return utils.NotifyMsg{Msg: str, Duration: time.Second}
		}
	}
	return nil
}

func (j *torrentSearchMsgDelegate) Desc() string {
	return "[enter] to download"
}

func InteractiveDownload(urls []string, savePath, saveCategory, saveTags string, autoMM bool) string {
	addParams := url.Values{}
	addParams.Set("category", saveCategory)
	addParams.Set("tags", saveTags)
	addParams.Set("auto-manage", strconv.FormatBool(autoMM))
	addParams.Set("save-path", savePath)
	LoadTorrentAddDefault(addParams)
	if err := api.TorrentAdd(urls, addParams); err != nil {
		return fmt.Sprintf("auto download failed: %s", err)
	} else {
		return fmt.Sprintf("auto download %d torrent(s) success\n", len(urls))
	}
}
