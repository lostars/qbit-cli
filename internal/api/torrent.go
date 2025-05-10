package api

import (
	"fmt"
	"net/http"
	"net/url"
)

// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-5.0)
// all the /torrent/* api here

func TorrentList(params url.Values) []Torrent {
	client, err := GetQbitClient()
	if err != nil {
		printApiClientError("TorrentList", err)
		return nil
	}

	resp, err := client.Get("/api/v2/torrents/info", params)
	if err != nil {
		printApiGetError("TorrentList", err)
		return nil
	}
	defer resp.Body.Close()

	var torrentList []Torrent
	if err := client.ParseJSON(resp, &torrentList); err != nil {
		printApiParsJSONError("TorrentList", err)
		return nil
	}
	return torrentList
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

func TorrentFiles(params url.Values) []TorrentFile {
	client, err := GetQbitClient()
	if err != nil {
		printApiClientError("TorrentFiles", err)
		return nil
	}
	resp, err := client.Get("/api/v2/torrents/files", params)
	if err != nil {
		printApiGetError("TorrentFiles", err)
		return nil
	}
	defer resp.Body.Close()

	var torrentFiles []TorrentFile
	if err := client.ParseJSON(resp, &torrentFiles); err != nil {
		printApiParsJSONError("TorrentFiles", err)
		return nil
	}
	return torrentFiles
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
		return fmt.Errorf("torrent:[%s] rename folder fail: %s", hash, resp.Status)
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
		return fmt.Errorf("torrent:[%s] rename file fail: %s", hash, resp.Status)
	}
	return nil
}
