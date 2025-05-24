package api

import "time"

type JackettIndexer struct {
	ID         string `json:"id"`
	Configured bool   `json:"configured"`
	Language   string `json:"language"`
	SiteLink   string `json:"site_link"`
	Type       string `json:"type"`
	Caps       *[]struct {
		ID   string `json:"ID"`
		Name string `json:"Name"`
	} `json:"caps"`
}

type JackettResults struct {
	Indexers *[]struct {
		ElapsedTime int    `json:"ElapsedTime"`
		ID          string `json:"ID"`
		Name        string `json:"Name"`
		Results     int    `json:"Results"`
		Status      int    `json:"Status"`
	} `json:"Indexers"`
	Results *[]JackettResult `json:"Results"`
}

type JackettResult struct {
	Title        string    `json:"Title"`
	Category     []int     `json:"Category"`
	CategoryDesc string    `json:"CategoryDesc"`
	Link         string    `json:"Link"`      // direct download url
	MagnetUri    string    `json:"MagnetUri"` // maybe empty
	Peers        int       `json:"Peers"`
	Seeders      int       `json:"Seeders"`
	TrackerId    string    `json:"TrackerId"`
	TrackerType  string    `json:"TrackerType"`
	Size         int64     `json:"Size"`
	PublishDate  time.Time `json:"PublishDate"`
	InfoHash     string    `json:"InfoHash"`
	Guid         string    `json:"Guid"`
	Details      string    `json:"Details"`
}
