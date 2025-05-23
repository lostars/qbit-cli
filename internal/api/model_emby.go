package api

import (
	"qbit-cli/pkg/utils"
	"time"
)

type EmbyItems struct {
	Items            []EmbyItem `json:"Items"`
	TotalRecordCount int        `json:"TotalRecordCount"`
}

type EmbyItem struct {
	ID             string            `json:"Id"`
	Name           string            `json:"Name"`
	Type           string            `json:"Type"`
	ProductionYear int               `json:"ProductionYear"`
	PremiereDate   time.Time         `json:"PremiereDate"`
	MediaSources   []EmbyMediaSource `json:"MediaSources"`
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
