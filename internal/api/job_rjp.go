package api

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type RenameJP struct{}

func (r *RenameJP) JobName() string {
	return "rjp"
}

func (r *RenameJP) Description() string {
	return `You may use [flags] to filter you JP torrents.
Only support torrent that contains one single jp video code.
Only rename files those are selected to download.
Torrent file struct supported as follows(file extension not mattered):
a.mp4
folder/a.mp4
`
}

func (_ *RenameJP) Tags() []string {
	return []string{"qBittorrent"}
}

func init() {
	RegisterJob(&RenameJP{})
}

func (r *RenameJP) RunCommand() *cobra.Command {

	jp := &cobra.Command{
		Use:   r.JobName(),
		Short: "Auto rename jp video filename and directory",
		Long:  r.Description(),
	}

	var (
		state, category, hashes, tag string
		renameTorrent                bool
	)

	jp.Flags().StringVar(&state, "filter", "", `state filter:
all, downloading, seeding, completed, stopped, active, inactive, 
running, stalled, stalled_uploading, stalled_downloading, errored`)
	jp.Flags().StringVar(&category, "category", "", "category filter")
	jp.Flags().StringVar(&tag, "tag", "", "tag filter")
	jp.Flags().StringVar(&hashes, "hashes", "", "hash filter separated by |'")
	jp.Flags().BoolVar(&renameTorrent, "rename-torrent", false, "whether to rename torrent files")

	jp.RunE = func(cmd *cobra.Command, args []string) error {
		params := url.Values{}
		if state != "" {
			params.Set("filter", state)
		}
		if tag != "" {
			params.Set("tag", tag)
		}
		if hashes != "" {
			params.Set("hashes", hashes)
		}
		if category != "" {
			params.Set("category", category)
		}

		torrentList, err := TorrentList(params)
		if err != nil {
			return err
		}

		fmt.Printf("total size: %d\n", len(torrentList))
		for _, t := range torrentList {
			// get torrent files
			fileList, err := TorrentFiles(url.Values{"hash": {t.Hash}})
			if fileList == nil {
				fmt.Println(err.Error())
				continue
			}

			for _, file := range fileList {
				// priority = 0 means file is not selected to download
				if file.Priority == 0 {
					continue
				}

				files := strings.Split(file.Name, "/")
				l := len(files)
				if l == 1 {
					jpCode := parseJPName(files[0], "")
					if jpCode == "" {
						continue
					}
					rename(renameTorrent, t, jpCode)
					if newPath := jpCode + filepath.Ext(files[0]); newPath != files[0] {
						if err := TorrentRenameFile(t.Hash, file.Name, newPath); err != nil {
							fmt.Printf("hash:[%s] new path: %s rename file failed: %v\n", t.Hash, newPath, err)
						}
						rename(renameTorrent, t, jpCode)
					}
				} else if l == 2 {
					newFolder := parseJPCode(files[1], files[0])
					if newFolder == "" {
						continue
					}
					rename(renameTorrent, t, newFolder)
					// rename only when name changed
					sleep := false
					if newFolder != files[0] {
						sleep = true
						if err := TorrentRenameFolder(t.Hash, files[0], newFolder); err != nil {
							fmt.Printf("[%s] %s -> %s renameFolder failed\n", t.Hash, files[0], newFolder)
						}
					}
					newPath := newFolder + "/" + parseJPName(files[1], files[0]) + filepath.Ext(files[1])
					oldPath := newFolder + "/" + files[1]
					if newPath != oldPath {
						if sleep {
							time.Sleep(500 * time.Millisecond)
						}
						if err := TorrentRenameFile(t.Hash, oldPath, newPath); err != nil {
							fmt.Printf("[%s] %s -> %s renameFile failed: %s\n", t.Hash, oldPath, newPath, err)
						}
					}
				}
			}
		}

		return nil
	}

	return jp
}

func rename(rename bool, t Torrent, name string) {
	if !rename || t.Name == name {
		return
	}
	err := RenameTorrent(t.Hash, name)
	if err != nil {
		fmt.Printf("%s rename failed: %v\n", t.Hash, err)
	}
}

func parseJPCode(fileName string, folder string) string {
	matches := JPCodeRegex.FindStringSubmatch(fileName)
	jpCode := ""
	if len(matches) <= 1 {
		matches = JPCodeRegex.FindStringSubmatch(folder)
		if len(matches) <= 1 {
			return ""
		}
	}
	jpCode = matches[1]

	matches = JP4KRegex.FindStringSubmatch(fileName)
	if len(matches) > 2 {
		jpCode += "-" + matches[2]
	}
	return jpCode
}

func parseJPName(fileName string, folder string) string {
	jpCode := parseJPCode(fileName, folder)
	if jpCode == "" {
		return ""
	}

	if matches := JPPartsRegex.FindStringSubmatch(fileName); len(matches) > 2 {
		jpCode += "-cd" + matches[2]
	}

	if JPCNRegex.MatchString(fileName) {
		jpCode += "-C"
	}

	return jpCode
}

var JPCodeRegex = regexp.MustCompile(`([a-zA-Z]{2,5}-[0-9]{3,5}|FC2-PPV-\d{5,})`)
var JP4KRegex = regexp.MustCompile(`([-\[])(4[kK])`)
var JPPartsRegex = regexp.MustCompile(`\d+([-_]|-cd)([1-5])^[kK]`)
var JPCNRegex = regexp.MustCompile(`\d+(-[cC]|ch)`)
