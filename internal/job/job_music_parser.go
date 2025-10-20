package job

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"qbit-cli/internal/api"
	c "qbit-cli/internal/cmd"
	"qbit-cli/pkg/utils"
	"strings"
	"time"
)

type MusicParser struct {
	quality    string
	savePath   string
	ffmpeg     string
	parserType string
}

var parserConfig MusicParser

func (parser *MusicParser) coverPathFromAudio(audioPath, coverUrl string) string {
	albumPicPath := strings.TrimSuffix(audioPath, filepath.Ext(audioPath))
	if ext := filepath.Ext(coverUrl); ext != "" {
		albumPicPath += ext
	} else {
		albumPicPath += ".jpg"
	}
	return albumPicPath
}

func (parser *MusicParser) newSavePath(subfolder string) error {
	parser.savePath = filepath.Join(parser.savePath, subfolder)
	if !utils.FileExists(parser.savePath) {
		if err := os.Mkdir(parser.savePath, 0777); err != nil {
			return err
		}
	}
	return nil
}

func (parser *MusicParser) JobName() string {
	return "mp"
}

func (parser *MusicParser) Tags() []string {
	return []string{"Parser", "Music"}
}

func (parser *MusicParser) Description() string {
	return `Music Parser
Parser will automatically save audio metadata include cover and lyrics if you specify ffmpeg command path or got it installed.
You can use --available-quality to show available quality.
Naming rule:
album songs will save to [artist]-[album] folder with [track].[song].[fileExt]
playlist songs will save to [playlist] folder with [track].[song].[fileExt]
single song will save with [track].[song].[fileExt]
NeteaseMusic Parser from https://github.com/Suxiaoqinx/Netease_url
QQMusic Parser from https://github.com/Suxiaoqinx/tencent_url
`
}

func init() {
	api.RegisterJob(&MusicParser{})
}

type musicParserHandler interface {
	parseSong(id string) error
	parseAlbum(id string) error
	parsePlaylist(id string) error
	availableQuality()
}

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

type musicBatchParser interface {
	parseSongs(id ...string) error
}

var parsers = []string{"qq", "netease"}

func (parser *MusicParser) RunCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   parser.JobName(),
		Short: "Music Parser(MP)",
	}

	var (
		songIds, albumIds, playlistIds []string
		showQuality                    bool
	)

	parserDefine := c.FlagsProperty[string]{Flag: "parser", Options: parsers}

	cmd.Flags().StringSliceVar(&songIds, "song-ids", []string{}, "song ids separated by comma")
	cmd.Flags().StringSliceVar(&albumIds, "album-ids", []string{}, "album ids separated by comma")
	cmd.Flags().StringSliceVar(&playlistIds, "playlist-ids", []string{}, "playlist ids separated by comma")

	// common config
	cmd.Flags().StringVar(&parserConfig.ffmpeg, "ffmpeg", "", "ffmpeg command path, provide to add metadata to audio file")
	cmd.Flags().StringVar(&parserConfig.savePath, "save-path", "", "song file save path")
	cmd.Flags().StringVar(&parserConfig.quality, "quality", "", `parser quality, defaults:
netease: exhigh
qq: ogg_640`)
	cmd.Flags().StringVar(&parserConfig.parserType, parserDefine.Flag, "", "Music parser: "+strings.Join(parsers, "|"))

	cmd.Flags().BoolVar(&showQuality, "available-quality", false, "show available quality")

	// netease config
	var maxBitrate bool
	cmd.Flags().BoolVar(&maxBitrate, "max-bitrate", false, `enable max bitrate, only valid with netease parser.
Parser will automatically parse the maximum bitrate downloadable audio, ensure you got a VIP.`)

	// register completion
	parserDefine.RegisterCompletion(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var p musicParserHandler
		switch parserConfig.parserType {
		default:
			return errors.New("unknown parser")
		case "netease":
			p = &NeteaseMusicParser{maxBitrate: maxBitrate}
		case "qq":
			p = &QQMusicParser{}
		}

		if showQuality {
			p.availableQuality()
			return nil
		}

		if batch, ok := p.(musicBatchParser); ok {
			if err := batch.parseSongs(songIds...); err != nil {
				fmt.Println(err)
			}
		} else {
			for _, songId := range songIds {
				if err := p.parseSong(songId); err != nil {
					fmt.Println(err)
				}
			}
		}

		for _, id := range playlistIds {
			if err := p.parsePlaylist(id); err != nil {
				fmt.Println(err)
			}
		}

		// parse albums
		for _, id := range albumIds {
			if err := p.parseAlbum(id); err != nil {
				fmt.Println(err)
			}
		}

		return nil

	}
	return cmd
}

type SongMetadata struct {
	AlbumCover, AlbumName string
	Lyrics                string
	Name                  string
	Artists               []string
	Track                 string
	Disc                  string
	Publish               string
	MetadataBlockPicture  string
	ClearBeforeSaveFFMPEG bool
}

func (m *SongMetadata) save(file string) {
	if !utils.FileExists(file) {
		log.Printf("%s not exists\n", file)
		return
	}
	ext := filepath.Ext(file)
	if ext == "" {
		return
	}
	switch ext {
	default:
		log.Printf("%s is not supported to edit metadata\n", ext)
	case ".ogg":
		// ogg cover edit is not supported
		m.AlbumCover = ""
		m.ClearBeforeSaveFFMPEG = true
		m.saveWithFFMPEG(file)
	case ".mp3", ".flac", "m4a":
		// use ffmpeg
		m.saveWithFFMPEG(file)
	}
}

func (m *SongMetadata) saveWithFFMPEG(file string) {
	var args = make([]string, 0, 50)
	args = append(args, "-y", "-i", file)
	if utils.FileExists(m.AlbumCover) {
		args = append(args, "-i", m.AlbumCover, "-map", "0", "-map", "1", "-metadata:s:v", "comment=Cover (front)", "-disposition:v", "attached_pic")
	}
	if m.ClearBeforeSaveFFMPEG {
		args = append(args, "-map_metadata", "-1")
	}
	if m.Lyrics != "" {
		args = append(args, "-metadata", fmt.Sprintf("lyrics=%s", m.Lyrics))
	}
	if m.Name != "" {
		args = append(args, "-metadata", fmt.Sprintf("Title=%s", m.Name))
	}
	if m.Artists != nil && len(m.Artists) > 0 {
		args = append(args, "-metadata", fmt.Sprintf("Artist=%s", strings.Join(m.Artists, ",")))
	}
	if m.AlbumName != "" {
		args = append(args, "-metadata", fmt.Sprintf("Album=%s", m.AlbumName))
	}
	if m.Track != "" && m.Track != "0" {
		args = append(args, "-metadata", fmt.Sprintf("Track=%s", m.Track))
	}
	if m.Disc != "" && m.Disc != "0" {
		args = append(args, "-metadata", fmt.Sprintf("Disc=%s", m.Disc))
	}
	if m.Publish != "" {
		args = append(args, "-metadata", fmt.Sprintf("Date=%s", m.Publish))
	}
	newFile := file + filepath.Ext(file)
	args = append(args, "-codec", "copy", newFile)

	err := utils.CmdRun(parserConfig.ffmpeg, args, "ffmpeg", []string{"-version"})
	if err != nil {
		log.Println(err)
	} else {
		_ = os.Rename(newFile, file)
	}
}
