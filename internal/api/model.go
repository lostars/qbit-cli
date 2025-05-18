package api

import "fmt"

type SearchResults struct {
	Results []SearchDetail `json:"results"`
	Status  string         `json:"status"`
	Total   uint32         `json:"total"`
}

type SearchDetail struct {
	DescLink   string `json:"descrLink"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	FileURL    string `json:"fileUrl"`
	NBLeechers int32  `json:"nbLeechers"`
	NBSeeders  int32  `json:"nbSeeders"`
	SiteUrl    string `json:"siteUrl"`
}

type SearchResult struct {
	ID uint32 `json:"id"`
}

type Torrent struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	State    string  `json:"state"`
	Progress float32 `json:"progress"`
	Tags     string  `json:"tags"`
}

type TorrentFile struct {
	Name     string  `json:"name"`
	Priority uint8   `json:"priority"`
	Progress float32 `json:"progress"`
}

type RssRule struct {
	Enabled                   bool     `json:"enabled"`
	MustContain               string   `json:"mustContain"`
	MustNotContain            string   `json:"mustNotContain"`
	UseRegex                  bool     `json:"useRegex"`
	EpisodeFilter             string   `json:"episodeFilter"`
	PreviouslyMatchedEpisodes []string `json:"previouslyMatchedEpisodes"`
	AffectedFeeds             []string `json:"affectedFeeds"`
	IgnoreDays                int32    `json:"ignoreDays"`
	LastMatch                 string   `json:"lastMatch"`
	AddPaused                 bool     `json:"addPaused"`
	AssignedCategory          string   `json:"assignedCategory"`
	SavePath                  string   `json:"savePath"`
}

type SearchPlugin struct {
	Enabled             bool   `json:"enabled"`
	FullName            string `json:"fullName"`
	Name                string `json:"name"`
	SupportedCategories []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"supportedCategories"`
	Url     string `json:"url"`
	Version string `json:"version"`
}

type RssSub struct {
	UID string `json:"uid"`
	URL string `json:"url"`
}

type TorrentCategory struct {
	Name     string `json:"name"`
	SavePath string `json:"savePath"`
}

type QbitServerInfo struct {
	WebApiVersion     string
	AppVersion        string
	QtVersion         string `json:"qt"`
	LibTorrentVersion string `json:"libtorrent"`
	BoostVersion      string `json:"boost"`
	OpenSSLVersion    string `json:"openssl"`
	Bitness           int32  `json:"bitness"`
	Platform          string `json:"platform"`
	ZlibVersion       string `json:"zlib"`
}

func (info QbitServerInfo) String() string {
	return fmt.Sprintf("platform: %s\nwebapi: %s\napp: %s\nqt: %s\n"+
		"zlib: %s\nlibtorrent: %s\nboost: %s\nopenssl: %s\nbitness: %d",
		info.Platform, info.WebApiVersion, info.AppVersion, info.QtVersion, info.ZlibVersion, info.LibTorrentVersion,
		info.BoostVersion, info.OpenSSLVersion, info.Bitness)
}
