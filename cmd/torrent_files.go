package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"strconv"
)

func TorrentFiles() *cobra.Command {
	filesCmd := &cobra.Command{
		Use:     "files <hash>",
		Short:   "List torrent files by torrent hash",
		Example: `qbit torrent files <torrent hash>`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("torrent hash required")
			}
			return nil
		},
	}

	filesCmd.RunE = func(cmd *cobra.Command, args []string) error {
		hash := args[0]
		var params = url.Values{"hash": {hash}}

		torrentFiles, err := api.TorrentFiles(params)
		if err != nil {
			return err
		}

		fmt.Printf("total file size: %d\n", len(torrentFiles))
		fmt.Println("priority = 0 means file is not selected to download")
		headers := []string{"index", "name", "priority", "progress", "size"}
		var data = make([][]string, len(torrentFiles))
		for i, f := range torrentFiles {
			size := utils.FormatFileSizeAuto(uint64(f.Size), 0)
			data[i] = []string{strconv.FormatInt(int64(f.Index), 10), f.Name, strconv.Itoa(int(f.Priority)), utils.FormatPercent(f.Progress), size}
		}
		utils.PrintList(headers, &data)

		return nil
	}

	return filesCmd
}
