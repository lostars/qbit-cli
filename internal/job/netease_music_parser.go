package job

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qbit-cli/internal/api"
	"qbit-cli/internal/config"
	"qbit-cli/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type NeteaseMusicParser struct {
	maxBitrate    bool
	level         string
	savePath      string
	ffmpeg        string
	output        string
	songInfo      *NeteaseSongInfo
	songPrivilege *NeteaseSongPrivilege
}

func (parser *NeteaseMusicParser) JobName() string {
	return "nmp"
}

func (parser *NeteaseMusicParser) Tags() []string {
	return []string{"Parser", "NeteaseMusic"}
}

func (parser *NeteaseMusicParser) Description() string {
	return `A NeteaseMusic Parser implementation from https://github.com/Suxiaoqinx/Netease_url
Automatically save metadata to audio file if you install ffmpeg or set ffmpeg bin file path.
If you got -110 code, ensure you got a VIP.
Cookie is required in config yaml file.
Naming rule:
album songs will save to [artist]-[album] folder with [track].[song].[fileExt]
playlist songs will save to [playlist] folder with [track].[song].[fileExt]
single song will save with [track].[song].[fileExt]
`
}

func init() {
	api.RegisterJob(&NeteaseMusicParser{})
}

var quality = map[string]string{
	"standard": "标准音质",
	"exhigh":   "极高音质",
	"lossless": "无损音质",
	"hires":    "Hires音质",
	"sky":      "沉浸环绕声",
	"jyeffect": "高清环绕声",
	"jymaster": "超清母带",
}
var levelUsage = strings.Replace(fmt.Sprintf("song level: %s", quality), "map[", "[", 1)
var SongOutput = map[string]string{
	"json": "json",
	"file": "file",
}

var cookie = ""

func (parser *NeteaseMusicParser) RunCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   parser.JobName() + " <id>...",
		Short: "NeteaseMusic Parser(NMP)",
		Long:  parser.Description(),
	}

	var (
		o, savePath, level             string
		maxBitrate                     bool
		ffmpeg                         string
		songIds, albumIds, playlistIds []int64
	)

	cmd.Flags().Int64SliceVar(&songIds, "song-ids", []int64{}, "song ids separated by comma")
	cmd.Flags().Int64SliceVar(&albumIds, "album-ids", []int64{}, "album ids separated by comma")
	cmd.Flags().Int64SliceVar(&playlistIds, "playlist-ids", []int64{}, "playlist ids separated by comma")

	cmd.Flags().StringVar(&ffmpeg, "ffmpeg", "", "ffmpeg command path, provide to add metadata to audio file")
	cmd.Flags().BoolVar(&maxBitrate, "max-bitrate", true, `auto get max bitrate song, ensure that you got a VIP`)
	cmd.Flags().StringVar(&o, "output", "file", "output: json|file")
	cmd.Flags().StringVar(&savePath, "save-path", "", "song file save path")
	cmd.Flags().StringVar(&level, "level", "exhigh", levelUsage)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cookie = config.GetConfig().NeteaseMusicCookie
		if cookie == "" {
			return errors.New("cookie is required")
		}
		if quality[level] == "" {
			return errors.New("level is invalid")
		}
		if SongOutput[o] == "" {
			return errors.New("output is invalid")
		}
		p := NeteaseMusicParser{
			maxBitrate: maxBitrate,
			level:      level,
			savePath:   savePath,
			ffmpeg:     ffmpeg,
			output:     o,
		}

		// parse songs
		if len(songIds) > 0 {
			infos, err := getNeteaseSongInfo(songIds...)
			if err != nil {
				return err
			}
			for _, song := range infos.Songs {
				p.songInfo = &song
				p.songPrivilege = &song.Privilege
				_ = p.parseSong()
			}
		}

		// parse playlists
		for _, id := range playlistIds {
			_ = p.parsePlaylist(id)
		}

		// parse albums
		for _, id := range albumIds {
			_ = p.parseAlbum(id)
		}

		return nil
	}

	return cmd
}

func (parser *NeteaseMusicParser) parsePlaylist(id int64) error {
	result, err := playlistDetail(id)
	if err != nil {
		return err
	}
	if result == nil || result.Playlist.TrackIds == nil || len(result.Playlist.TrackIds) == 0 {
		return errors.New("no playlist songs found")
	}
	playlist := result.Playlist.Name
	if !utils.FileExists(playlist) {
		if err := os.Mkdir(playlist, 0777); err != nil {
			return err
		}
	}
	parser.savePath = filepath.Join(parser.savePath, playlist)

	songIds := make([]int64, 0, len(result.Playlist.TrackIds))
	for _, trackId := range result.Playlist.TrackIds {
		songIds = append(songIds, trackId.ID)
	}

	infos, err := getNeteaseSongInfo(songIds...)
	if err != nil {
		return err
	}

	for _, s := range infos.Songs {
		parser.songInfo = &s
		parser.songPrivilege = &s.Privilege
		_ = parser.parseSong()
	}
	return nil
}

func (parser *NeteaseMusicParser) parseAlbum(id int64) error {
	songs, err := albumDetail(id)
	if err != nil {
		return err
	}
	if songs == nil || len(*songs) < 1 {
		return errors.New("no songs found")
	}

	data := *songs
	album := data[0].Album.Name
	if len(data[0].Artists) > 0 {
		album = data[0].Artists[0].Name + "-" + album
	}
	if !utils.FileExists(album) {
		err := os.Mkdir(album, 0777)
		if err != nil {
			return err
		}
	}
	parser.savePath = filepath.Join(parser.savePath, album)
	for _, s := range data {
		parser.songInfo = &s
		parser.songPrivilege = &s.Privilege
		_ = parser.parseSong()
	}
	return nil
}

func (parser *NeteaseMusicParser) parseSong() error {
	songId := parser.songInfo.ID
	filename, albumPic := parser.songInfo.Name, ""
	if len(parser.songInfo.Artists) > 0 {
		albumPic = parser.songInfo.Album.PicUrl
	}

	if parser.maxBitrate && parser.songPrivilege.DownloadMaxBitrateLevel != "" {
		parser.level = parser.songPrivilege.DownloadMaxBitrateLevel
	}
	log.Printf("song level: %s\n", parser.level)
	song, err := getNeteaseSong(songId, parser.level, cookie)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if song.Code != 200 {
		errStr := fmt.Sprintf("%d:%s parse failed: %d", songId, parser.songInfo.Name, song.Code)
		fmt.Println(errStr)
		return errors.New(errStr)
	}

	track := ""
	if parser.songInfo.Track > 0 {
		track += strconv.Itoa(parser.songInfo.Track) + "."
	}
	filename = filepath.Join(parser.savePath, track+filename+"."+song.EncodeType)
	log.Println(filename)

	switch parser.output {
	case "file":
		// download audio file
		_ = utils.DownloadUrlToFile(filename, song.Url)
		// download cover
		cover := strings.TrimSuffix(filename, filepath.Ext(filename))
		if albumPic != "" {
			if ext := filepath.Ext(albumPic); ext != "" {
				cover += ext
			} else {
				cover += ".jpg"
			}
			_ = utils.DownloadUrlToFile(cover, albumPic)
		}
		// save lyric
		lyric := ""
		lr := getSongLyrics(songId)
		if lr != nil {
			lyric = lr.Lyric.Lyric
		}
		saveMetadata(parser.ffmpeg, filename, parser.songInfo, cover, lyric)
		_ = os.Remove(cover)
	case "json":
		d, _ := json.MarshalIndent(song, "", "  ")
		infoJson, _ := json.MarshalIndent(parser.songInfo, "", "  ")
		fmt.Println(string(infoJson))
		fmt.Println(string(d))
	}
	return nil
}

func albumDetail(albumId int64) (*[]NeteaseSongInfo, error) {
	endpoint := albumDetailEndpoint + strconv.FormatInt(albumId, 10)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var album NeteaseAlbumResult
	if err := json.NewDecoder(resp.Body).Decode(&album); err != nil {
		fmt.Println(err)
	}
	return &album.Songs, nil
}

func playlistDetail(playlistId int64) (*NeteasePlaylistResult, error) {
	params := url.Values{"id": {strconv.FormatInt(playlistId, 10)}}
	req, err := http.NewRequest(http.MethodPost, playlistEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result NeteasePlaylistResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &result, nil
}

func saveMetadata(cmd, file string, info *NeteaseSongInfo, coverFile, lyricStr string) {
	if info == nil {
		return
	}
	if !utils.FileExists(file) {
		log.Printf("%s not exists\n", file)
		return
	}

	var args = make([]string, 0, 10)
	args = append(args, "-y", "-i", file)
	if utils.FileExists(coverFile) {
		args = append(args, "-i", coverFile, "-map", "0", "-map", "1", "-metadata:s:v", "comment=Cover (front)")
	}
	if lyricStr != "" {
		args = append(args, "-metadata", fmt.Sprintf("lyrics=%s", lyricStr))
	}
	if info.Name != "" {
		args = append(args, "-metadata", fmt.Sprintf("Title=%s", info.Name))
	}
	if len(info.Artists) > 0 {
		args = append(args, "-metadata", fmt.Sprintf("Artist=%s", info.Artists[0].Name))
	}
	if info.Album.Name != "" {
		args = append(args, "-metadata", fmt.Sprintf("Album=%s", info.Album.Name))
	}
	if info.Track > 0 {
		args = append(args, "-metadata", fmt.Sprintf("Track=%d", info.Track))
	}
	if info.CD != "" {
		args = append(args, "-metadata", fmt.Sprintf("Disc=%s", info.CD))
	}
	if info.PublishTime > 0 {
		publishDate := time.UnixMilli(info.PublishTime)
		args = append(args, "-metadata", fmt.Sprintf("Date=%s", publishDate.Format("2006")))
	}
	newFile := file + filepath.Ext(file)
	args = append(args, "-codec", "copy", newFile)

	err := utils.FFMPEGRun(cmd, args)
	if err != nil {
		log.Println(err)
	} else {
		_ = os.Rename(newFile, file)
	}
}

var aesKey = []byte("e82ckenh8dichen8")

var client = &http.Client{
	Timeout: time.Second * 10,
}

var playlistEndpoint = "https://music.163.com/api/v6/playlist/detail"
var albumDetailEndpoint = "https://music.163.com/api/v1/album/"
var songDetailEndpoint = "https://interface3.music.163.com/api/v3/song/detail"
var songLyricsEndpoint = "https://interface3.music.163.com/api/song/lyric"
var songEndpoint = "https://interface3.music.163.com/eapi/song/enhance/player/url/v1"

func getSongLyrics(id int64) *NeteaseLyricsResult {
	params := url.Values{
		"id":  {strconv.FormatInt(id, 10)},
		"cp":  {"false"},
		"tv":  {"0"},
		"lv":  {"0"},
		"rv":  {"0"},
		"kv":  {"0"},
		"yv":  {"0"},
		"ytv": {"0"},
		"yrv": {"0"},
	}
	req, err := http.NewRequest(http.MethodPost, songLyricsEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var lyric NeteaseLyricsResult
	e := json.NewDecoder(resp.Body).Decode(&lyric)
	if e != nil {
		return nil
	}
	return &lyric
}

func getNeteaseSongInfo(ids ...int64) (*NeteaseSongInfoResult, error) {
	var payload = make([]map[string]interface{}, 0, len(ids))
	for _, id := range ids {
		payload = append(payload, map[string]interface{}{"id": id, "v": 0})
	}
	cJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	params := url.Values{"c": {string(cJSON)}}
	req, err := http.NewRequest(http.MethodPost, songDetailEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info NeteaseSongInfoResult
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func getNeteaseSong(id int64, level string, cookies string) (*NeteaseSong, error) {
	requestId := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10000000) + 20000000)
	headerConfig := map[string]string{
		"os":        "pc",
		"appver":    "",
		"osver":     "",
		"deviceId":  "pyncm!",
		"requestId": requestId,
	}
	payload := map[string]interface{}{
		"ids":        []int64{id},
		"level":      level,
		"encodeType": "flac",
		"header":     headerConfig,
	}
	if level == "sky" {
		payload["immerseType"] = "c51"
	}
	u, _ := url.Parse(songEndpoint)
	digestUrl := strings.Replace(u.Path, "/eapi/", "/api/", 1)
	payloadByte, _ := json.Marshal(payload)
	digestInput := fmt.Sprintf("nobody%suse%smd5forencrypt", digestUrl, string(payloadByte))
	digest := utils.MD5Hex(digestInput)

	paramsStr := fmt.Sprintf("%s-36cd479b6b5-%s-36cd479b6b5-%s", digestUrl, string(payloadByte), digest)
	encData, err := utils.AESEncryptECB([]byte(paramsStr), aesKey)
	if err != nil {
		return nil, err
	}

	params := url.Values{"params": {strings.ToUpper(hex.EncodeToString(encData))}}
	req, err := http.NewRequest(http.MethodPost, songEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result NeteaseSongResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Code != 200 {
		return nil, errors.New(strconv.Itoa(result.Code))
	}
	if len(result.Data) == 0 {
		return nil, errors.New("result data is empty")
	}
	return &result.Data[0], nil
}

type NeteaseSongResult struct {
	Data []NeteaseSong `json:"data"`
	Code int           `json:"code"`
}
type NeteaseSong struct {
	Url        string `json:"url"`
	Size       int64  `json:"size"`
	Bitrate    int64  `json:"br"`
	SampleRate int64  `json:"sr"`
	MD5        string `json:"md5"`
	Code       int    `json:"code"`
	Type       string `json:"type"`
	EncodeType string `json:"encodeType"`
	Payed      int    `json:"payed"`
	Level      string `json:"level"`
}

type NeteaseSongInfoResult struct {
	Songs      []NeteaseSongInfo      `json:"songs"`
	Code       int                    `json:"code"`
	Privileges []NeteaseSongPrivilege `json:"privileges"`
}

type NeteaseSongPrivilege struct {
	MaxBitrateLevel         string `json:"maxBrLevel"`
	PlayMaxBitrateLevel     string `json:"playMaxBrLevel"`
	DownloadMaxBitrateLevel string `json:"downloadMaxBrLevel"`
}

type NeteaseSongInfo struct {
	Name    string `json:"name"`
	ID      int64  `json:"id"`
	Artists []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"ar"`
	Album struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		PicUrl string `json:"picUrl"`
	} `json:"al"`
	CD          string               `json:"cd"`
	Track       int                  `json:"no"`
	PublishTime int64                `json:"publishTime"`
	Single      int                  `json:"single"` // 1=single
	Privilege   NeteaseSongPrivilege `json:"privilege"`
}

type NeteaseLyricsResult struct {
	Code  int `json:"code"`
	Lyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"lrc"`
}

type NeteaseAlbumResult struct {
	ResourceState bool              `json:"resourceState"`
	Songs         []NeteaseSongInfo `json:"songs"`
	Code          int               `json:"code"`
}

type NeteasePlaylistResult struct {
	Code     int `json:"code"`
	Playlist struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		TrackIds []struct {
			ID int64 `json:"id"`
		} `json:"trackIds"`
	} `json:"playlist"`
}
