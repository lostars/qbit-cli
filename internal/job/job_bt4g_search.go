package job

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"qbit-cli/internal/api"
	"qbit-cli/internal/cmd"
	"qbit-cli/internal/config"
	"qbit-cli/pkg/utils"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type BT4G struct {
	client *http.Client
}

func (r *BT4G) JobName() string {
	return "bt4g"
}

func (r *BT4G) Description() string {
	return `Search bt4g`
}

func (r *BT4G) Tags() []string {
	return []string{"search"}
}

func init() {
	api.RegisterJob(&BT4G{
		client: &http.Client{
			Timeout: time.Second * 180,
		},
	})
}

var categories = strings.Split("all,movie,audio,doc,app,other", ",")
var sortBy = strings.Split("time,size,seeders,relevance", ",")
var bt4gUrl = "https://bt4gprx.com"

func (r *BT4G) RunCommand() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "bt4g <keyword>",
		Short: "Search bt4g",
		Long:  `Search bt4g through flaresolverr`,
		Args:  cobra.ExactArgs(1),
	}

	category := cmd.FlagsProperty[string]{Flag: "category", Options: categories}
	sort := cmd.FlagsProperty[string]{Flag: "sort", Options: sortBy}

	var autoMM bool
	var savePath, saveTags string
	saveCategory := cmd.FlagsProperty[string]{Flag: "save-category", Register: &cmd.TorrentCategoryFlagRegister{}}
	runCmd.Flags().StringVar(&category.Value, "category", categories[0], "search category")
	runCmd.Flags().StringVar(&sort.Value, "sort", sortBy[0], "search sort by")

	runCmd.Flags().BoolVar(&autoMM, "auto-manage", true, "whether enable torrent auto management default is true, valid only when auto download enabled")
	runCmd.Flags().StringVar(&saveCategory.Value, saveCategory.Flag, "", "torrent save category, valid only when auto download enabled")
	runCmd.Flags().StringVar(&savePath, "save-path", "", "torrent save path, valid only when auto download enabled")
	runCmd.Flags().StringVar(&saveTags, "save-tags", "", "torrent save tags, valid only when auto download enabled")

	runCmd.RunE = func(cmd *cobra.Command, args []string) error {
		result, err := r.sendRequest(fmt.Sprintf("%s/search?q=%s&category=%s&orderby=%s", bt4gUrl, args[0], category.Value, sort.Value))
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}

		printList := parseList(result)
		if len(printList) < 1 {
			fmt.Println("no result")
			return nil
		}

		header := []string{"title", "createTime", "size", "leecher", "seeder"}
		data := make([][]string, 0, len(printList))
		for _, item := range printList {
			data = append(data, []string{item.Title, item.CreateTime, item.Size, item.Leecher, item.Seeder})
		}
		bt4gC := &bt4gIConfig{
			data:        printList,
			currentPage: 1, bt4g: r, pages: printList[0].Pages,
			keyword: args[0], category: category.Value, orderBy: sort.Value,
			autoMM:   autoMM,
			savePath: savePath, saveCategory: saveCategory.Value, saveTags: saveTags,
		}
		model := utils.InteractiveTableModel{
			Rows:         &data,
			Header:       &header,
			WidthMap:     map[int]int{0: 30},
			Delegate:     bt4gC,
			DataDelegate: bt4gC,
		}
		if _, e := tea.NewProgram(&model, tea.WithAltScreen()).Run(); e != nil {
			return e
		}

		return nil
	}

	return runCmd
}

type bt4gIConfig struct {
	autoMM                           bool
	savePath, saveCategory, saveTags string

	bt4g *BT4G

	currentPage, pages int

	keyword, category, orderBy string

	data []*Bt4gSearchResult
}

func (b *bt4gIConfig) Headers() *[]string {
	return nil
}

func (b *bt4gIConfig) Rows() *[][]string {
	result, err := b.bt4g.sendRequest(fmt.Sprintf("%s/search?q=%s&category=%s&orderby=%s&p=%d", bt4gUrl, b.keyword, b.category, b.orderBy, b.currentPage))
	if err != nil {
		return nil
	}

	torrentList := parseList(result)
	var data = make([][]string, len(torrentList))
	for i, item := range torrentList {
		data[i] = []string{item.Title, item.CreateTime, item.Size, item.Leecher, item.Seeder}
	}
	b.data = torrentList
	return &data
}

func (b *bt4gIConfig) Frequency() time.Duration {
	return time.Hour
}

func (b *bt4gIConfig) Desc() string {
	return "[enter] download; [left] previous page; [right] next page"
}

func (b *bt4gIConfig) Operation(msg tea.KeyMsg, cursor int) *utils.KeyMsgDelegateModel {
	switch msg.String() {
	case "enter":
		if b.data == nil || cursor >= len(b.data) {
			return nil
		}
		magnet := b.bt4g.download(b.data[cursor].Url)

		str := ""
		if magnet != "" {
			str = cmd.InteractiveDownload([]string{magnet}, b.savePath, b.saveCategory, b.saveTags, b.autoMM)
		} else {
			str = "download failed from bt4g"
		}

		return &utils.KeyMsgDelegateModel{
			RenderClicked: true,
			NotifyMsg:     utils.NotifyMsg{Msg: str},
		}
	case "left":
		if b.currentPage <= 1 {
			return nil
		}
		b.currentPage--
		return &utils.KeyMsgDelegateModel{
			RenderClicked: false,
			NotifyMsg:     utils.NotifyMsg{UpdateData: true, PendingMsg: "Querying..."},
		}
	case "right":
		if b.currentPage >= b.pages {
			return nil
		}
		b.currentPage++
		return &utils.KeyMsgDelegateModel{
			RenderClicked: false,
			NotifyMsg:     utils.NotifyMsg{UpdateData: true, PendingMsg: "Querying..."},
		}
	}
	return nil
}

func (r *BT4G) download(hash string) string {
	result, err := r.sendRequest(fmt.Sprintf("%s%s", bt4gUrl, hash))
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(result))
	if err != nil {
		return ""
	}
	url, _ := doc.Find("a.btn.btn-info.me-2").Attr("href")

	matches := magnetRegex.FindStringSubmatch(url)
	if len(matches) > 1 {
		magnet := strings.Replace(matches[1], "%3F", "?", 1)
		return magnet + "&tr=" + strings.Join(bt4gTrackers, "&tr=")
	}
	return ""
}

var magnetRegex = regexp.MustCompile(`/(magnet:%3Fxt=.*)`)

type Bt4gSearchResult struct {
	Pages                  int
	Title, Url, CreateTime string
	Size                   string
	Leecher, Seeder        string
}

func parseList(rawHTML string) []*Bt4gSearchResult {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(rawHTML))
	if err != nil {
		return nil
	}

	page := 1
	pageE := doc.Find(".pagination .page-item")
	if pageE != nil {
		page = pageE.Length()
	}

	list := doc.Find("div.list-group-item.result-item")
	if list == nil {
		return nil
	}
	var data = make([]*Bt4gSearchResult, list.Length())
	list.Each(func(i int, s *goquery.Selection) {
		titleE := s.Find(".mb-1 a")
		title, _ := titleE.Attr("title")
		href, _ := titleE.Attr("href")

		createTime := ""
		doc.Find("span.me-2").Each(func(i int, s *goquery.Selection) {
			fullText := s.Text()
			if strings.Contains(fullText, "Creation Time:") {
				createTime = strings.TrimSpace(strings.Split(fullText, "Creation Time:")[1])
			}
		})

		size := s.Find("b.cpill").Text()
		leecher := s.Find("#leechers").Text()
		seeder := s.Find("#seeders").Text()

		data[i] = &Bt4gSearchResult{page, title, href, createTime, size, leecher, seeder}
	})

	return data
}

var htmlTokenRegex = regexp.MustCompile(`var token = "(.*)";`)
var htmlTimestampRegex = regexp.MustCompile(`var timestamp = "(\d+)";`)

func (r *BT4G) sendRequest(url string) (result string, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	r.loadFlaresolverrAuth(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}
	defer utils.SafeClose(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	result = string(bodyBytes)
	if resp.StatusCode == 403 {
		result, err = r.sendFlaresolverrRequest(url)
		return
	}

	tokenMatches := htmlTokenRegex.FindStringSubmatch(result)
	tsMatches := htmlTimestampRegex.FindStringSubmatch(result)
	if len(tokenMatches) > 1 && len(tsMatches) > 1 {
		// resend request
		req.Header.Add("Cookie", fmt.Sprintf("cf_chl_out=%s|%s;", tokenMatches[1], tsMatches[1]))
		resp, err = r.client.Do(req)
		if err != nil {
			return
		}
		defer utils.SafeClose(resp.Body)

		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		result = string(bodyBytes)
	}

	return
}

func (r *BT4G) loadFlaresolverrAuth(req *http.Request) {
	fileData, err := os.ReadFile(filepath.Join(filepath.Dir(config.GetConfig().ConfigPath()), flaresolverrFile))
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	var fs Flaresolverr
	err = json.Unmarshal(fileData, &fs)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	req.Header.Set("User-Agent", fs.Solution.UA)
	req.Header.Set("Origin", bt4gUrl)

	var b strings.Builder
	for _, cMap := range fs.Solution.Cookies {
		b.WriteString(fmt.Sprintf("%s=%s;", cMap["name"], cMap["value"]))
	}
	req.Header.Add("Cookie", b.String())
}

var flaresolverrFile = "flaresolverr.json"

func (r *BT4G) sendFlaresolverrRequest(urlStr string) (string, error) {

	flaresolverr := config.GetConfig().Flaresolverr
	if flaresolverr == "" {
		return "", fmt.Errorf("no flaresolverr config")
	}

	params := map[string]interface{}{
		"cmd":        "request.get",
		"maxTimeout": 180000,
		"url":        urlStr,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(params)

	req, err := http.NewRequest(http.MethodPost, flaresolverr, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", err
	}
	defer utils.SafeClose(resp.Body)

	var fs Flaresolverr
	err = json.NewDecoder(resp.Body).Decode(&fs)
	if err != nil {
		return "", fmt.Errorf("parse json error: %e", err)
	}

	if fs.Status != "ok" || fs.Solution.Status != 200 {
		return "", fmt.Errorf("flaresolverr message: %s", fs.Message)
	}

	// save auth to json file
	jsonFile := filepath.Join(filepath.Dir(config.GetConfig().ConfigPath()), flaresolverrFile)
	file, err := os.Create(jsonFile)
	if err != nil {
		return "", err
	}
	defer utils.SafeClose(file)
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(fs)
	if err != nil {
		return "", err
	}

	return fs.Solution.Response, nil
}

type Flaresolverr struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	StartTimestamp int64  `json:"startTimestamp"`
	EndTimestamp   int64  `json:"endTimestamp"`
	Version        string `json:"version"`
	Solution       struct {
		Url            string                   `json:"url"`
		Status         int                      `json:"status"`
		Response       string                   `json:"response"`
		Headers        map[string]interface{}   `json:"headers"`
		Cookies        []map[string]interface{} `json:"cookies"`
		UA             string                   `json:"userAgent"`
		TurnstileToken string                   `json:"turnstile_token"`
	} `json:"solution"`
}

var bt4gTrackers = []string{
	"udp://tracker.opentrackr.org:1337/announce",
	"udp://open.demonii.com:1337/announce",
	"udp://open.stealth.si:80/announce",
	"udp://wepzone.net:6969/announce",
	"udp://vito-tracker.space:6969/announce",
	"udp://vito-tracker.duckdns.org:6969/announce",
	"udp://udp.tracker.projectk.org:23333/announce",
	"udp://tracker.tryhackx.org:6969/announce",
	"udp://tracker.torrent.eu.org:451/announce",
	"udp://tracker.theoks.net:6969/announce",
	"udp://tracker.t-1.org:6969/announce",
	"udp://tracker.srv00.com:6969/announce",
	"udp://tracker.qu.ax:6969/announce",
	"udp://tracker.plx.im:6969/announce",
	"udp://tracker.opentorrent.top:6969/announce",
	"udp://tracker.gmi.gd:6969/announce",
	"udp://tracker.fnix.net:6969/announce",
	"udp://tracker.flatuslifir.is:6969/announce",
	"udp://tracker.filemail.com:6969/announce",
	"udp://tracker.ducks.party:1984/announce",
	"https://tracker.bt4g.com:443/announce",
}
