package api

import (
	"fmt"
	"net/http"
	"net/url"
)

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)
// all the /torrent/* api here

func TorrentList(params url.Values) ([]Torrent, error) {
	client, err := GetQbitClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Get("/api/v2/torrents/info", params)
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

func TorrentAdd(params url.Values) error {
	client, err := GetQbitClient()
	if err != nil {
		return err
	}
	resp, err := client.Post("/api/v2/torrents/add", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("torrent add fail: %s", resp.Status)
	}
	return nil
}

func TorrentFiles(params url.Values) ([]TorrentFile, error) {
	client, err := GetQbitClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Get("/api/v2/torrents/files", params)
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
	client, err := GetQbitClient()
	if err != nil {
		return err
	}

	params := url.Values{
		"hash":    {hash},
		"oldPath": {old},
		"newPath": {new},
	}
	resp, err := client.Post("/api/v2/torrents/renameFolder", params)
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
	client, err := GetQbitClient()
	if err != nil {
		return err
	}

	params := url.Values{
		"hash":    {hash},
		"oldPath": {old},
		"newPath": {new},
	}
	resp, err := client.Post("/api/v2/torrents/renameFile", params)
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
	c, err := GetQbitClient()
	if err != nil {
		return err
	}

	params := url.Values{
		"hash": {hash},
		"name": {name},
	}
	resp, err := c.Post("/api/v2/torrents/rename", params)
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
