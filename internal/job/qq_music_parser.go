package job

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qbit-cli/internal/config"
	"qbit-cli/pkg/utils"
	"strconv"
	"strings"
)

type QQMusicParser struct {
	UIN      string
	songInfo *songData
}

type songData struct {
	Mid, Track, Disc, Name       string
	AlbumName, AlbumMid, Publish string
	PlaylistName                 string
	ID                           int64
	Artists                      []string
}

func (i *QQMusicSongInfo) toSongData() *songData {
	d := songData{
		Mid:       i.Mid,
		Name:      i.Name,
		AlbumName: i.Album.Name,
		AlbumMid:  i.Album.Mid,
		ID:        i.ID,
		Artists:   i.artists(),
	}
	if i.Track > 0 {
		d.Track = strconv.Itoa(i.Track)
	}
	if i.Disc > 0 {
		d.Disc = strconv.Itoa(i.Disc)
	}
	if i.Album.Publish != "" {
		d.Publish = i.Album.Publish[:4]
	}
	return &d
}

func (t *QQMusicSongTrack) toSongData() *songData {
	d := songData{
		Mid:       t.SongMid,
		Name:      t.SongName,
		AlbumName: t.AlbumName,
		AlbumMid:  t.AlbumMid,
		ID:        t.SongId,
	}
	if len(t.Singers) > 0 {
		singers := make([]string, 0, len(t.Singers))
		for _, singer := range t.Singers {
			singers = append(singers, singer.Name)
		}
		d.Artists = singers
	}
	if t.TrackId > 0 {
		d.Track = strconv.Itoa(t.TrackId)
	}
	if t.Disc > 0 {
		d.Disc = strconv.Itoa(t.Disc)
	}
	return &d
}

func (t *QQMusicPlaylist) toSongData() *[]songData {
	var tracks []songData
	for _, track := range t.Tracks {
		d := track.toSongData()
		d.PlaylistName = t.Name
		tracks = append(tracks, *d)
	}
	return &tracks
}

func (album *QQMusicAlbum) toSongData() *[]songData {
	var tracks []songData
	for _, t := range album.Tracks {
		d := t.toSongData()
		if album.Publish != "" {
			d.Publish = album.Publish[:4]
		}
		tracks = append(tracks, *d)
	}
	return &tracks
}

func (parser *QQMusicParser) parseAlbum(id string) error {
	album, err := getAlbum(id)
	if err != nil {
		return err
	}
	// save to singer-album folder
	tracks := album.toSongData()
	err = parserConfig.newSavePath(album.AlbumSinger + "-" + album.Name)
	if err != nil {
		return err
	}

	for _, t := range *tracks {
		parser.songInfo = &t
		err = parser.parseSong(t.Mid)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (parser *QQMusicParser) parsePlaylist(id string) error {
	playlist, err := getPlaylist(id)
	if err != nil {
		return err
	}
	// save to playlist sub folder
	tracks := playlist.toSongData()
	err = parserConfig.newSavePath(playlist.Name)
	if err != nil {
		return err
	}

	for _, t := range *tracks {
		parser.songInfo = &t
		err = parser.parseSong(t.Mid)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (parser *QQMusicParser) parseSong(id string) error {
	var q QQMusicQuality
	// check quality
	if parserConfig.quality != "" {
		if qua, ok := QQMusicQualityType[parserConfig.quality]; !ok {
			return errors.New("invalid quality")
		} else {
			q = qua
		}
	} else {
		parserConfig.quality = "ogg_640"
	}

	if parser.songInfo == nil {
		// get from song id
		i, err := songInfo(id)
		if err != nil {
			return err
		}
		parser.songInfo = i.toSongData()
	}
	songMid := parser.songInfo.Mid
	payload := map[string]interface{}{
		"req_1": map[string]interface{}{
			"module": "vkey.GetVkeyServer",
			"method": "CgiGetVkey",
			"param": map[string]interface{}{
				"filename":  []string{fmt.Sprintf("%s%s%s%s", q.S, songMid, songMid, q.Ext)},
				"guid":      "10000",
				"songmid":   []string{songMid},
				"songtype":  []int{0},
				"uin":       parser.UIN,
				"loginflag": 1,
				"platform":  "20",
			},
		},
		"loginUin": parser.UIN,
		"comm": map[string]interface{}{
			"uin":    parser.UIN,
			"format": "json",
			"ct":     24,
			"cv":     0,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, qqSongParseEndpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", qqCookie)
	req.Header.Set("User-Agent", userAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result QQMusicParseResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Code != 0 || result.Req1.Data.Sip == nil || len(result.Req1.Data.Sip) == 0 {
		return errors.New(fmt.Sprintf("parse error %d", result.Code))
	}
	// even if u get a VIP, some songs still lack some quality.
	if result.Req1.Data.MidUrlInfo[0].PUrl == "" {
		return errors.New(fmt.Sprintf("songMid: %s %s, parse error: need a VIP or change another quality",
			songMid, parserConfig.quality))
	}

	// download audio file
	track := ""
	if parser.songInfo.Track != "" {
		track = parser.songInfo.Track + "."
	}
	audioFilePath := filepath.Join(parserConfig.savePath, track+parser.songInfo.Name+q.Ext)
	log.Println(audioFilePath)
	audioUrl := result.Req1.Data.Sip[0] + result.Req1.Data.MidUrlInfo[0].PUrl
	err = utils.DownloadUrlToFile(audioFilePath, audioUrl)
	if err != nil {
		return err
	}

	// download cover
	albumPicPath := ""
	if parser.songInfo.AlbumMid != "" {
		albumPicUrl := fmt.Sprintf(qqSongAlbumCoverFormat, parser.songInfo.AlbumMid)
		albumPicPath = parserConfig.coverPathFromAudio(audioFilePath, albumPicUrl)
		err = utils.DownloadUrlToFile(albumPicPath, albumPicUrl)
		if err != nil {
			return err
		}
		defer os.Remove(albumPicPath)
	}
	// get lyrics
	lrc, err := getLyrics(parser.songInfo.ID)
	if err != nil {
		return err
	}
	lyrics, err := base64.StdEncoding.DecodeString(lrc.Data.Lyric)
	if err != nil {
		return err
	}
	l := string(lyrics)
	if strings.Index(l, "此歌曲为没有填词的纯音乐") >= 0 {
		l = ""
	}
	// save metadata
	metadata := SongMetadata{
		AlbumCover: albumPicPath,
		AlbumName:  parser.songInfo.AlbumName,
		Lyrics:     l,
		Name:       parser.songInfo.Name,
		Artists:    parser.songInfo.Artists,
		Track:      parser.songInfo.Track,
		Disc:       parser.songInfo.Disc,
	}
	if parser.songInfo.Publish != "" {
		metadata.Publish = parser.songInfo.Publish
	}
	metadata.save(audioFilePath)

	return nil
}

var qqCookie = config.GetConfig().QQMusicCookie
var qqSongInfoEndpoint = "https://c.y.qq.com/v8/fcg-bin/fcg_play_single_song.fcg"
var qqSongAlbumCoverFormat = "https://y.qq.com/music/photo_new/T002R800x800M000%s.jpg"
var qqSongParseEndpoint = "https://u.y.qq.com/cgi-bin/musicu.fcg"
var qqAlbumEndpoint = "https://c.y.qq.com/v8/fcg-bin/fcg_v8_album_info_cp.fcg"
var qqPlaylistEndpoint = "https://c.y.qq.com/qzone/fcg-bin/fcg_ucc_getcdinfo_byids_cp.fcg"

var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"

func getPlaylist(id string) (*QQMusicPlaylist, error) {
	params := url.Values{
		"disstid": []string{id},
		"format":  []string{"json"},
		"type":    []string{"1"},
		"utf8":    []string{"1"},
	}
	req, err := http.NewRequest(http.MethodGet, qqPlaylistEndpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("referer", "https://y.qq.com/")
	req.Header.Set("Cookie", qqCookie)
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QQMusicPlaylistResult
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	// use sub code to check not code
	if result.SubCode != 0 {
		return nil, fmt.Errorf("playlist result error code: %d %s", result.SubCode, result.Msg)
	}
	if len(result.Playlists) < 1 {
		return nil, errors.New("no playlist found")
	}

	return &result.Playlists[0], nil
}

func getAlbum(id string) (*QQMusicAlbum, error) {
	params := url.Values{
		"albummid": []string{id},
		"format":   []string{"json"},
	}
	req, err := http.NewRequest(http.MethodGet, qqAlbumEndpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", qqCookie)
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QQMusicAlbumResult
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.SubCode != 0 {
		return nil, fmt.Errorf("album result error code: %d %s", result.SubCode, result.Msg)
	}
	return &result.Album, nil
}

func getLyrics(songId int64) (*QQMusicSongLyric, error) {
	payload := map[string]interface{}{
		"music.musichallSong.PlayLyricInfo.GetPlayLyricInfo": map[string]interface{}{
			"module": "music.musichallSong.PlayLyricInfo",
			"method": "GetPlayLyricInfo",
			"param": map[string]interface{}{
				"trans_t":    0,
				"roma_t":     0,
				"crypt":      0, // 1 define to encrypt
				"lrc_t":      0,
				"interval":   208,
				"trans":      1,
				"ct":         6,
				"singerName": "",
				"type":       0,
				"qrc_t":      0,
				"cv":         80600,
				"roma":       1,
				"songID":     songId,
				"qrc":        0, // 1 define base64 or compress Hex
				"albumName":  "",
				"songName":   "",
			},
		},
		"comm": map[string]interface{}{
			"wid":                         "",
			"tmeAppID":                    "qqmusic",
			"authst":                      "",
			"uid":                         "",
			"gray":                        "0",
			"OpenUDID":                    "",
			"ct":                          "6",
			"patch":                       "2",
			"psrf_qqopenid":               "",
			"sid":                         "",
			"psrf_access_token_expiresAt": "",
			"cv":                          "80600",
			"gzip":                        "0",
			"qq":                          "",
			"nettype":                     "2",
			"psrf_qqunionid":              "",
			"psrf_qqaccess_token":         "",
			"tmeLoginType":                "2",
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, qqSongParseEndpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Cookie", qqCookie)
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QQMusicParseResult
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 || result.Lyric.Code != 0 {
		return nil, errors.New(fmt.Sprintf("%d get lyrics error: %d, %d", songId, result.Code, result.Lyric.Code))
	}

	return &result.Lyric, nil
}

func songInfo(id string) (*QQMusicSongInfo, error) {
	params := url.Values{
		"platform": {"yqq"},
		"format":   {"json"},
	}
	_, e := strconv.ParseInt(id, 10, 64)
	if e != nil {
		params.Set("songmid", id)
	} else {
		params.Set("songid", id)
	}

	req, err := http.NewRequest(http.MethodPost, qqSongInfoEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", qqCookie)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QQMusicSongInfoResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(strconv.Itoa(result.Code))
	}
	if len(result.Data) < 1 {
		return nil, errors.New("no song found")
	}

	return &result.Data[0], nil
}

func (parser *QQMusicParser) availableQuality() {
	for k, v := range QQMusicQualityType {
		fmt.Printf("%s: %s\n", k, v.Bitrate)
	}
}

var QQMusicQualityType = map[string]QQMusicQuality{
	"128":      {"M500", ".mp3", "128kbps"},
	"320":      {"M800", ".mp3", "320kbps"},
	"flac":     {"F000", ".flac", "FLAC"},
	"master":   {"AI00", ".flac", "Master"},
	"atmos_2":  {"Q000", ".flac", "Atmos 2"},
	"atmos_51": {"Q001", ".flac", "Atmos 5.1"},
	"ogg_640":  {"O801", ".ogg", "640kbps"},
	"ogg_320":  {"O800", ".ogg", "320kbps"},
	"ogg_192":  {"O600", ".ogg", "192kbps"},
	"ogg_96":   {"O400", ".ogg", "96kbps"},
	"aac_192":  {"C600", ".m4a", "192kbps"},
	"aac_96":   {"C400", ".m4a", "96kbps"},
	"aac_48":   {"C200", ".m4a", "48kbps"},
}

type QQMusicQuality struct {
	S, Ext, Bitrate string
}

type QQMusicSongInfo struct {
	Name  string `json:"name"`
	ID    int64  `json:"id"`
	Mid   string `json:"mid"`
	Track int    `json:"index_album"`
	Disc  int    `json:"index_cd"`
	Album struct {
		ID      int64  `json:"id"`
		Mid     string `json:"mid"`
		Name    string `json:"name"`
		Publish string `json:"time_public"`
	} `json:"album"`
	Singer []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Mid  string `json:"mid"`
	} `json:"singer"`
}

func (i *QQMusicSongInfo) artists() []string {
	var artists = make([]string, 0, len(i.Singer))
	for _, s := range i.Singer {
		artists = append(artists, s.Name)
	}
	return artists
}

type QQMusicSongInfoResult struct {
	Code int               `json:"code"`
	Data []QQMusicSongInfo `json:"data"`
}

type QQMusicParseResult struct {
	Code int `json:"code"`
	Req1 struct {
		Code int `json:"code"`
		Data struct {
			MidUrlInfo []struct {
				SongMid  string `json:"songmid"`
				Filename string `json:"filename"`
				PUrl     string `json:"purl"`
				Result   int    `json:"result"`
			} `json:"midurlinfo"`
			Sip []string `json:"sip"`
		} `json:"data"`
	} `json:"req_1"`
	Lyric QQMusicSongLyric `json:"music.musichallSong.PlayLyricInfo.GetPlayLyricInfo"`
}

type QQMusicSongLyric struct {
	Code int `json:"code"`
	Data struct {
		SongId int64  `json:"songID"`
		Lyric  string `json:"lyric"`
		Trans  string `json:"trans"`
	} `json:"data"`
}

type QQMusicAlbumResult struct {
	Code    int          `json:"code"`
	Msg     string       `json:"message"`
	SubCode int          `json:"subcode"`
	Album   QQMusicAlbum `json:"data"`
}

type QQMusicAlbum struct {
	Publish     string             `json:"aDate"`
	Genre       string             `json:"genre"`
	ID          int64              `json:"id"`
	Lang        string             `json:"lan"`
	Name        string             `json:"name"`
	Mid         string             `json:"mid"`
	AlbumSinger string             `json:"singername"`
	Tracks      []QQMusicSongTrack `json:"list"`
}

type QQMusicSongTrack struct {
	AlbumName string `json:"albumname"`
	AlbumMid  string `json:"albummid"`
	AlbumId   int64  `json:"albumid"`
	SongId    int64  `json:"songid"`
	SongMid   string `json:"songmid"`
	SongName  string `json:"songname"`
	TrackId   int    `json:"belongCD"`
	Disc      int    `json:"cdIdx"`
	Singers   []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Mid  string `json:"mid"`
	} `json:"singer"`
}

type QQMusicPlaylist struct {
	Name   string             `json:"dissname"`
	ID     string             `json:"disstid"`
	Tracks []QQMusicSongTrack `json:"songlist"`
}

type QQMusicPlaylistResult struct {
	Code      int               `json:"code"`
	Msg       string            `json:"msg"`
	SubCode   int               `json:"subcode"`
	Playlists []QQMusicPlaylist `json:"cdlist"`
}
