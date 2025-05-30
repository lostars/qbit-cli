package job

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qbit-cli/internal/config"
	"qbit-cli/pkg/utils"
	"strconv"
	"strings"
	"time"
)

type NeteaseMusicParser struct {
	maxBitrate    bool
	songInfo      *NeteaseSongInfo
	songPrivilege *NeteaseSongPrivilege
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

func (parser *NeteaseMusicParser) availableQuality() {
	for k, v := range quality {
		fmt.Printf("%s: %s\n", k, v)
	}
}

var neteaseCookie = config.GetConfig().NeteaseMusicCookie

func (parser *NeteaseMusicParser) parsePlaylist(id string) error {
	playlistId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("invalid playlist id")
	}
	result, err := playlistDetail(playlistId)
	if err != nil {
		return err
	}
	if result == nil || result.Playlist.TrackIds == nil || len(result.Playlist.TrackIds) == 0 {
		return errors.New("no playlist songs found")
	}
	playlist := result.Playlist.Name
	err = parserConfig.newSavePath(playlist)
	if err != nil {
		return err
	}

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
		_ = parser.parseSong(strconv.FormatInt(s.ID, 10))
	}
	return nil
}

func (parser *NeteaseMusicParser) parseAlbum(id string) error {
	albumId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("invalid album id")
	}
	songs, err := albumDetail(albumId)
	if err != nil {
		return err
	}
	if songs == nil || len(*songs) < 1 {
		return errors.New(fmt.Sprintf("album %s no songs found", id))
	}

	data := *songs
	album := data[0].Album.Name
	if len(data[0].Artists) > 0 {
		album = data[0].Artists[0].Name + "-" + album
	}
	err = parserConfig.newSavePath(album)
	if err != nil {
		return err
	}
	for _, s := range data {
		parser.songInfo = &s
		parser.songPrivilege = &s.Privilege
		_ = parser.parseSong(strconv.FormatInt(s.ID, 10))
	}
	return nil
}

func (parser *NeteaseMusicParser) parseSong(id string) error {
	if parser.songInfo == nil {
		log.Printf("netease song load from id: %s\n", id)
		songId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return errors.New("invalid song id")
		}
		info, err := getNeteaseSongInfo(songId)
		if err != nil {
			return err
		}
		parser.songInfo = &info.Songs[0]
		if info.Privileges != nil && len(info.Privileges) > 0 {
			parser.songPrivilege = &info.Privileges[0]
		} else {
			log.Printf("%s no privileges found\n", id)
		}
	}

	// check quality
	if parserConfig.quality != "" {
		if q := quality[parserConfig.quality]; q == "" {
			return errors.New("invalid quality")
		}
	} else {
		parserConfig.quality = "exhigh"
	}

	songId := parser.songInfo.ID
	filename, albumPic := parser.songInfo.Name, ""
	if len(parser.songInfo.Artists) > 0 {
		albumPic = parser.songInfo.Album.PicUrl
	}

	if parser.maxBitrate && parser.songPrivilege.DownloadMaxBitrateLevel != "" {
		parserConfig.quality = parser.songPrivilege.DownloadMaxBitrateLevel
	}
	log.Printf("song quality: %s\n", parserConfig.quality)
	song, err := getNeteaseSong(songId, parserConfig.quality)
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
	filename = filepath.Join(parserConfig.savePath, track+filename+"."+song.EncodeType)

	// download audio file
	err = utils.DownloadUrlToFile(filename, song.Url)
	if err != nil {
		return err
	}
	// download cover
	cover := parserConfig.coverPathFromAudio(filename, albumPic)
	if albumPic != "" {
		_ = utils.DownloadUrlToFile(cover, albumPic)
		defer os.Remove(cover)
	}
	// get lyrics
	lyric := ""
	lr := getSongLyrics(songId)
	if lr != nil {
		lyric = lr.Lyric.Lyric
	}
	// save metadata
	metadata := SongMetadata{
		AlbumCover: cover,
		AlbumName:  parser.songInfo.Album.Name,
		Lyrics:     lyric,
		Name:       parser.songInfo.Name,
		Artists:    parser.songInfo.artists(),
		Track:      strconv.Itoa(parser.songInfo.Track),
		Disc:       parser.songInfo.CD,
		Publish:    time.UnixMilli(parser.songInfo.PublishTime).Format("2006"),
	}
	metadata.save(filename)

	return nil
}

func (parser *NeteaseMusicParser) parseSongs(id ...string) error {
	if len(id) == 0 {
		return nil
	}
	var ids = make([]int64, 0, len(id))
	for _, v := range id {
		if songId, err := strconv.ParseInt(v, 10, 64); err == nil {
			ids = append(ids, songId)
		}
	}
	infos, err := getNeteaseSongInfo(ids...)
	if err != nil {
		return err
	}
	for _, song := range infos.Songs {
		parser.songInfo = &song
		parser.songPrivilege = &song.Privilege
		_ = parser.parseSong("")
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
	req.Header.Set("Cookie", neteaseCookie)
	resp, err := httpClient.Do(req)
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
	req.Header.Set("Cookie", neteaseCookie)
	resp, err := httpClient.Do(req)
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

var aesKey = []byte("e82ckenh8dichen8")

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
	req.Header.Set("Cookie", neteaseCookie)
	resp, err := httpClient.Do(req)
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
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info NeteaseSongInfoResult
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}
	if info.Songs == nil || len(info.Songs) < 1 {
		log.Println("ids: ", ids)
		return nil, errors.New("no songs found")
	}

	return &info, nil
}

func getNeteaseSong(id int64, level string) (*NeteaseSong, error) {
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
	req.Header.Set("Cookie", neteaseCookie)
	resp, err := httpClient.Do(req)
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

func (i *NeteaseSongInfo) artists() []string {
	if i.Artists == nil {
		return nil
	}
	var artists []string
	for _, artist := range i.Artists {
		artists = append(artists, artist.Name)
	}
	return artists
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
