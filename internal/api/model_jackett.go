package api

import "time"

type JackettIndexer struct {
	ElapsedTime int    `json:"ElapsedTime"`
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Results     int    `json:"Results"`
	Status      int    `json:"Status"`
}

type JackettResults struct {
	Indexers *[]JackettIndexer `json:"Indexers"`
	Results  *[]JackettResult  `json:"Results"`
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
