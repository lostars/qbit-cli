package api

import (
	"fmt"
)

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
	EngineName string `json:"engineName"`
}

type SearchResult struct {
	ID uint32 `json:"id"`
}

type Torrent struct {
	Hash     string  `json:"hash"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	State    string  `json:"state"`
	Progress float64 `json:"progress"`
	Tags     string  `json:"tags"`
	DLSpeed  int64   `json:"dlspeed"`
	UPSpeed  int64   `json:"upspeed"`
	Size     int64   `json:"size"`
}

type TorrentFile struct {
	Name     string  `json:"name"`
	Priority uint8   `json:"priority"`
	Progress float64 `json:"progress"`
	Index    int32   `json:"index"`
	Size     int64   `json:"size"`
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
	SmartFilter               bool     `json:"smartFilter"`
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

type TorrentTracker struct {
	URL           string `json:"url"`
	Status        int    `json:"status"`
	Tier          int    `json:"tier"`
	NumPeers      int    `json:"num_peers"`
	NumSeeds      int    `json:"num_seeds"`
	NumLeeches    int    `json:"num_leeches"`
	NumDownloaded int    `json:"num_downloaded"`
	Msg           string `json:"msg"`
}

type TorrentPeer struct {
	IP           string  `json:"ip"`
	Port         int     `json:"port"`
	Client       string  `json:"client"`
	Connection   string  `json:"connection"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	DLSpeed      int64   `json:"dl_speed"`
	Downloaded   int64   `json:"downloaded"`
	Files        string  `json:"files"`
	Flags        string  `json:"flags"`
	FlagsDesc    string  `json:"flags_desc"`
	PeerIDClient string  `json:"peer_id_client"`
	Progress     float64 `json:"progress"`
	Relevance    float64 `json:"relevance"`
	Shadowbanned bool    `json:"shadowbanned"`
	UpSpeed      int64   `json:"up_speed"`
	Uploaded     int64   `json:"uploaded"`
}

type PeerResult struct {
	Peers      map[string]TorrentPeer `json:"peers"`
	FullUpdate bool                   `json:"full_update"`
	Rid        int64                  `json:"rid"`
	ShowFlags  bool                   `json:"show_flags"`
}
