package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"strconv"
	"time"
)

func TorrentList() *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   "List torrents",
		Example: `qbit torrent list --filter=downloading --category=abc`,
	}

	var (
		state, category, hashes, tag string
		limit, offset                uint32
		interactive                  bool
	)

	listCmd.Flags().StringVar(&state, "state", "", `state filter:
all, downloading, seeding, completed, stopped, active, inactive, running, 
stalled, stalled_uploading, stalled_downloading, errored`)
	listCmd.Flags().StringVar(&category, "category", "", "category filter")
	listCmd.Flags().StringVar(&tag, "tag", "", "tag filter")
	listCmd.Flags().StringVar(&hashes, "hashes", "", "hash filter separated by |")
	listCmd.Flags().Uint32Var(&limit, "limit", 0, "results limit")
	listCmd.Flags().Uint32Var(&offset, "offset", 0, "results offset")
	listCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode")

	listCmd.RunE = func(cmd *cobra.Command, args []string) error {
		d := torrentSearch{
			state:    state,
			category: category,
			hashes:   hashes,
			tag:      tag,
			limit:    limit,
			offset:   offset,
		}
		if interactive {
			headers := []string{"name", "hash", "CATE", "state", "PROG", "DOWN", "UP"}
			model := utils.InteractiveTableModel{
				Rows:         d.Rows(),
				Header:       &headers,
				WidthMap:     map[int]int{0: 30},
				DataDelegate: &d,
				Delegate:     &d,
			}
			if _, e := tea.NewProgram(&model, tea.WithAltScreen()).Run(); e != nil {
				return e
			}
			return nil
		}

		torrentList, err := d.fetchData()
		if err != nil {
			return err
		}

		fmt.Printf("total size: %d\n", len(*torrentList))

		headers := []string{"name", "hash", "CATE", "tags", "state", "size", "PROG"}
		var data = make([][]string, len(*torrentList))
		for i, t := range *torrentList {
			data[i] = []string{t.Name, t.Hash, t.Category, t.Tags, t.State, utils.FormatFileSizeAuto(uint64(t.Size), 1), utils.FormatPercent(t.Progress)}
		}
		utils.PrintListWithColWidth(headers, &data, map[int]int{0: 30, 6: 6}, false)
		return nil
	}

	return listCmd
}

type torrentSearch struct {
	state, category, tag, hashes string
	limit, offset                uint32
	rows                         *[][]string
}

func (t *torrentSearch) Frequency() time.Duration {
	return time.Second
}

func (t *torrentSearch) Operation(msg tea.KeyMsg, cursor int) *utils.KeyMsgDelegateModel {
	switch msg.String() {
	case "D":
		params := url.Values{}
		params.Set("deleteFiles", "true")
		return t.notifyReturn(cursor, "delete", params)
	case "d":
		params := url.Values{}
		params.Set("deleteFiles", "false")
		return t.notifyReturn(cursor, "delete", params)
	case "r":
		return t.notifyReturn(cursor, "start", nil)
	case "p":
		return t.notifyReturn(cursor, "stop", nil)
	}
	return nil
}

func (t *torrentSearch) notifyReturn(cursor int, operation string, params url.Values) *utils.KeyMsgDelegateModel {
	data := *t.rows
	hash := data[cursor][1]
	notify := "success"
	if params == nil {
		params = url.Values{}
	}
	params.Set("hashes", hash)
	if err := api.UpdateTorrent(operation, params); err != nil {
		notify = err.Error()
	}
	return &utils.KeyMsgDelegateModel{
		RenderClicked: false,
		NotifyMsg:     utils.NotifyMsg{Msg: notify, Duration: time.Second},
	}
}

func (t *torrentSearch) Desc() string {
	return "[p] pause; [r] resume; [d] delete; [shift+d](D) delete with files"
}

func (t *torrentSearch) fetchData() (*[]api.Torrent, error) {
	var params = url.Values{}
	if t.state != "" {
		params.Set("filter", t.state)
	}
	if t.category != "" {
		// category must be encoded
		params.Set("category", t.category)
	}
	if t.tag != "" {
		// tag must be encoded
		params.Set("tag", t.tag)
	}
	if t.hashes != "" {
		params.Set("hashes", t.hashes)
	}
	if t.limit > 0 {
		params.Set("limit", strconv.FormatUint(uint64(t.limit), 10))
	}
	if t.offset > 0 {
		params.Set("offset", strconv.FormatUint(uint64(t.offset), 10))
	}

	torrentList, err := api.TorrentList(params)
	if err != nil {
		return nil, err
	}
	return &torrentList, nil
}

func (t *torrentSearch) Headers() *[]string {
	return nil
}

func (t *torrentSearch) Rows() *[][]string {
	torrentList, err := t.fetchData()
	if err != nil {
		return nil
	}
	var data = make([][]string, len(*torrentList))
	for i, t := range *torrentList {
		dl := utils.FormatFileSizeAuto(uint64(t.DLSpeed), 1) + "/S"
		up := utils.FormatFileSizeAuto(uint64(t.UPSpeed), 1) + "/S"
		data[i] = []string{t.Name, t.Hash, t.Category, t.State, utils.FormatPercent(t.Progress), dl, up}
	}
	t.rows = &data
	return &data
}
