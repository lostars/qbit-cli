package api

import (
	"qbit-cli/pkg/utils"
	"time"
)

type EmbyItems struct {
	Items            []EmbyItem `json:"Items"`
	TotalRecordCount int        `json:"TotalRecordCount"` // do not use, may return 0 even items is not empty
}

type EmbyItem struct {
	ID             string            `json:"Id"`
	ParentId       string            `json:"ParentId"`
	SeriesId       string            `json:"SeriesId"`
	SeriesName     string            `json:"SeriesName"`
	ChildCount     int               `json:"ChildCount"`
	Status         string            `json:"Status"`
	SeasonId       string            `json:"SeasonId"`
	SeasonName     string            `json:"SeasonName"`
	Album          string            `json:"Album"`
	AlbumId        string            `json:"AlbumId"`
	AlbumArtist    string            `json:"AlbumArtist"`
	IndexNumber    int               `json:"IndexNumber"`
	Name           string            `json:"Name"`
	Type           string            `json:"Type"`
	Path           string            `json:"Path"`
	Genres         []string          `json:"Genres"`
	ProductionYear int               `json:"ProductionYear"`
	PremiereDate   time.Time         `json:"PremiereDate"`
	CreatedDate    time.Time         `json:"DateCreated"`
	MediaSources   []EmbyMediaSource `json:"MediaSources"`
	CollectionType string            `json:"CollectionType"`
	ImageTags      struct {
		Primary string `json:"Primary"`
		Logo    string `json:"Logo"`
	} `json:"ImageTags"`
	Overview string `json:"Overview"`
}

func (item *EmbyItem) IsMovie() bool {
	return item.Type == "Movie"
}
func (item *EmbyItem) IsSeries() bool {
	return item.Type == "Series"
}
func (item *EmbyItem) IsSeason() bool {
	return item.Type == "Season"
}
func (item *EmbyItem) IsMusicAlbum() bool {
	return item.Type == "MusicAlbum"
}
func (item *EmbyItem) IsMusicCollection() bool {
	return item.CollectionType == "music"
}
func (item *EmbyItem) IsMovieCollection() bool {
	return item.CollectionType == "movies"
}
func (item *EmbyItem) IsTVShowCollection() bool {
	return item.CollectionType == "tvshows"
}

type EmbyMediaSource struct {
	ID           string         `json:"Id"`
	Path         string         `json:"Path"`
	Container    string         `json:"Container"`
	Size         uint64         `json:"Size"`
	MediaStreams []MediaStreams `json:"MediaStreams"`
}

type MediaStreams struct {
	Type         string `json:"Type"`
	Height       int    `json:"Height"`
	Width        int    `json:"Width"`
	Codec        string `json:"Codec"`
	Index        int    `json:"Index"`
	DisplayTitle string `json:"DisplayTitle"`
}

func (item *EmbyItem) MediaVideoStream() *[]MediaStreams {
	if item.MediaSources == nil || len(item.MediaSources) == 0 {
		return nil
	}
	var streams []MediaStreams
	for _, source := range item.MediaSources {
		if source.MediaStreams == nil || len(source.MediaStreams) == 0 {
			return nil
		}
		for _, stream := range source.MediaStreams {
			if stream.Type == "Video" {
				streams = append(streams, stream)
			}
		}
	}
	return &streams
}

func (item *EmbyItem) MediaVideoSourceCount() int {
	if item.MediaSources == nil {
		return 0
	}
	return len(item.MediaSources)
}

func (item *EmbyItem) MediaVideoSourceSizeView() []string {
	if item.MediaSources == nil || len(item.MediaSources) == 0 {
		return nil
	}
	var result []string
	for _, source := range item.MediaSources {
		result = append(result, utils.FormatFileSizeAuto(source.Size, 1))
	}
	return result
}
