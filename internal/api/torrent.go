package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)
// all the /torrent/* api here

func TorrentList(params url.Values) ([]Torrent, error) {
	resp, err := GetQbitClient().Get("/api/v2/torrents/info", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var torrentList []Torrent
	if err := ParseJSON(resp, &torrentList); err != nil {
		return nil, err
	}
	return torrentList, nil
}

func TorrentAdd(urls []string, params url.Values) error {
	var localFiles = make([]*os.File, 0, len(urls))
	var netUrl = make([]string, 0, len(urls))
	for _, path := range urls {
		file, _ := os.Open(path)
		if file != nil {
			localFiles = append(localFiles, file)
		} else {
			netUrl = append(netUrl, path)
		}
	}
	c := GetQbitClient()
	if len(localFiles) > 0 {
		params.Del("urls")
		resp, err := c.PostForm("/api/v2/torrents/add", params, "torrents", localFiles)
		if err != nil {
			fmt.Println(err)
		} else if resp.StatusCode == http.StatusUnsupportedMediaType {
			fmt.Println("file is not valid")
		} else if resp.StatusCode != http.StatusOK {
			fmt.Println(resp.Status)
		}
		resp.Body.Close()
		for _, file := range localFiles {
			file.Close()
		}
	}

	if len(netUrl) > 0 {
		params.Set("urls", strings.Join(netUrl, "\n"))
		resp, err := c.Post("/api/v2/torrents/add", params)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("torrent add fail: %s", resp.Status)
		}
	}

	return nil
}

func TorrentFiles(params url.Values) ([]TorrentFile, error) {
	resp, err := GetQbitClient().Get("/api/v2/torrents/files", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var torrentFiles []TorrentFile
	if err := ParseJSON(resp, &torrentFiles); err != nil {
		return nil, err
	}
	return torrentFiles, nil
}

func TorrentRenameFolder(hash string, old string, new string) error {
	params := url.Values{
		"hash":    {hash},
		"oldPath": {old},
		"newPath": {new},
	}
	resp, err := GetQbitClient().Post("/api/v2/torrents/renameFolder", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "TorrentRenameFolder", nil}
	}
	return nil
}

func TorrentRenameFile(hash string, old string, new string) error {
	params := url.Values{
		"hash":    {hash},
		"oldPath": {old},
		"newPath": {new},
	}
	resp, err := GetQbitClient().Post("/api/v2/torrents/renameFile", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "TorrentRenameFile", nil}
	}
	return nil
}

func RenameTorrent(hash string, name string) error {
	params := url.Values{
		"hash": {hash},
		"name": {name},
	}
	resp, err := GetQbitClient().Post("/api/v2/torrents/rename", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return &QbitClientError{"hash is invalid", "TorrentRename", nil}
	}
	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "RenameTorrent", nil}
	}
	return nil
}

func UpdateTorrent(operation string, params url.Values) error {
	resp, err := GetQbitClient().Post("/api/v2/torrents/"+operation, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "TorrentUpdate: " + operation, nil}
	}
	return nil
}

func TagList() ([]string, error) {
	resp, err := GetQbitClient().Get("/api/v2/torrents/tags", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tagList []string
	if err := ParseJSON(resp, &tagList); err != nil {
		return nil, err
	}
	return tagList, nil
}

// TagUpdate deleteTags createTags
func TagUpdate(operation string, name []string) error {
	params := url.Values{}
	params.Set("tags", strings.Join(name, ","))
	resp, err := GetQbitClient().Post("/api/v2/torrents/"+operation, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func CategoryList() (*[]TorrentCategory, error) {
	resp, err := GetQbitClient().Get("/api/v2/torrents/categories", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var categories map[string]TorrentCategory
	if err := ParseJSON(resp, &categories); err != nil {
		return nil, err
	}
	results := make([]TorrentCategory, 0, len(categories))
	for _, category := range categories {
		results = append(results, category)
	}
	return &results, nil
}

func CategoryAdd(name string, path string) error {
	params := url.Values{}
	params.Set("category", name)
	params.Set("savePath", path)
	resp, err := GetQbitClient().Post("/api/v2/torrents/createCategory", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return errors.New("category already exists")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return errors.New(resp.Status)
	}
	return nil
}

func CategoryDelete(names []string) error {
	params := url.Values{}
	params.Set("categories", strings.Join(names, "\n"))
	resp, err := GetQbitClient().Post("/api/v2/torrents/removeCategories", params)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func CategoryUpdate(name string, path string) error {
	params := url.Values{}
	params.Set("category", name)
	params.Set("savePath", path)
	resp, err := GetQbitClient().Post("/api/v2/torrents/editCategory", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return errors.New("category not exists")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return errors.New(resp.Status)
	}
	return nil
}
