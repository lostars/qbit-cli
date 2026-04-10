package job

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"qbit-cli/internal/api"
	"qbit-cli/internal/config"
	"qbit-cli/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Gofile struct {
	maxWorker           int
	savePath            string
	client              *http.Client
	url, code, password string
	wt, token           string
}

var gofile Gofile

func (r *Gofile) JobName() string {
	return "gofile"
}

func (r *Gofile) Description() string {
	return `Resolve Gofile file share url and download.`
}

func (r *Gofile) Tags() []string {
	return []string{"resolver"}
}

func init() {
	api.RegisterJob(&Gofile{})
}

func GofileTrafficCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic",
		Short: "show your ip traffic in 1 month",
		Args:  cobra.NoArgs,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		g := Gofile{
			client: buildGofileHttpClient(),
		}
		err := g.setAuth()
		if err != nil {
			return err
		}

		traffic, err := g.getTraffic()
		if err != nil {
			return err
		}
		if traffic.Status != "ok" || traffic.Data.IpTraffic == nil {
			return errors.New("traffic return " + traffic.Status)
		}

		var start = time.Now().AddDate(0, -1, 0)
		var used uint64 = 0
		for yearNum, year := range traffic.Data.IpTraffic {
			y, err := strconv.ParseInt(yearNum, 10, 64)
			if err != nil {
				continue
			}
			for monthNum, month := range year {
				m, err := strconv.ParseInt(monthNum, 10, 64)
				if err != nil {
					continue
				}
				for dayNum, day := range month {
					d, err := strconv.ParseInt(dayNum, 10, 64)
					if err != nil {
						continue
					}
					var date = time.Date(int(y), time.Month(m), int(d), 0, 0, 0, 0, time.Local)
					if date.After(start) {
						log.Printf("%s: %s\n", date.Format("2006-01-02"), utils.FormatFileSizeAuto(day, 1))
						used += day
					}
				}
			}
		}

		fmt.Println(traffic.Data.IpInfo.IP)
		fmt.Printf("Total IP Traffic: %s\n", utils.FormatFileSizeAuto(used, 1))
		fmt.Printf("Total IP Traffic In Bytes: %v\n", used)

		return nil
	}

	return cmd
}

func buildGofileHttpClient() *http.Client {
	return &http.Client{Timeout: time.Minute * 1}
}

func (r *Gofile) getTraffic() (*GofileTrafficResp, error) {
	var trafficApi = "https://api.gofile.io/accounts/website"
	req, _ := http.NewRequest(http.MethodGet, trafficApi, nil)
	req.Header.Set("Authorization", "Bearer "+r.token)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer utils.SafeClose(resp.Body)

	var traffic GofileTrafficResp
	err = json.NewDecoder(resp.Body).Decode(&traffic)
	if err != nil {
		return nil, err
	}

	return &traffic, nil
}

type GofileTrafficResp struct {
	Status string `json:"status"`
	Data   struct {
		IpTraffic map[string]map[string]map[string]uint64 `json:"ipTraffic"`
		IpInfo    struct {
			IP   string `json:"_id"`
			CIDR string `json:"cidr"`
		} `json:"ipinfo"`
	} `json:"data"`
}

func (r *Gofile) RunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gofile <url>",
		Short: "Resolve Gofile file share url and download.",
		Long: `Anonymous download support now. 
For free accounts, traffic is counted by IP address. 
If you're using a shared IP address, this may include additional traffic beyond your own usage.`,
		Args: cobra.ExactArgs(1),
	}

	cmd.AddCommand(GofileTrafficCmd())

	cmd.Flags().StringVar(&gofile.savePath, "save-path", "", "file save path, default is current working directory")
	cmd.Flags().StringVarP(&gofile.password, "password", "p", "", "url share password")
	cmd.Flags().IntVar(&gofile.maxWorker, "max-worker", 2, "max worker number, if u get too many requests error, set it smaller or 1")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		l := args[0]
		// valid url like https://gofile.io/d/xxx
		if !strings.Contains(l, "/d/") {
			return errors.New("unknown url")
		}
		u, err := url.Parse(l)
		if err != nil {
			return errors.New("unknown url")
		}
		gofile.code = path.Base(u.Path)
		if gofile.code == "" {
			return errors.New("invalid url")
		}
		gofile.url = l
		gofile.client = buildGofileHttpClient()

		/*
			1. Get token from https://api.gofile.io/accounts
			2. Get file download url from https://api.gofile.io/contents/xxx?contentFilter=&page=1&pageSize=1000&sortField=name&sortDirection=1
			3. Download files with cookie(cookie accountToken=xxx) from phase 1 and X-Website-Token(wt) and X-BL(language)
		*/

		err = gofile.setAuth()
		if err != nil {
			return err
		}

		var contents = gofile.getFiles()
		if contents == nil {
			return errors.New("failed to get files")
		}
		if contents.Status != "ok" {
			return errors.New("contents return " + contents.Status)
		}
		if len(contents.Data.Children) <= 0 {
			return errors.New("contents is empty, maybe need a password")
		}

		downloader := utils.HttpFileDownloader{
			SavePath:  gofile.savePath,
			MaxWorker: gofile.maxWorker,
			Client:    gofile.client,
			Headers: map[string]string{
				"Cookie": "accountToken=" + gofile.token,
			},
		}
		if config.Debug {
			downloader.DebugLogger = log.New(os.Stdout, "[HttpFileDownloader]: ", log.LstdFlags)
		}

		// download file
		var files = make([]GofileFile, 0, len(contents.Data.Children))
		for _, f := range contents.Data.Children {
			files = append(files, f)
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].Size < files[j].Size
		})
		for _, f := range files {
			var fullPath = path.Join(gofile.savePath, f.Name)
			fmt.Printf("saving %s to %s\n", f.Name, fullPath)
			dErr := downloader.Download(f.Link, f.Name)
			if dErr != nil {
				fmt.Printf("file %s download fail: %s\n", f.Link, err)
			} else {
				fmt.Printf("complete saving %s to %s\n", f.Link, fullPath)
			}
		}

		return nil
	}

	return cmd
}

func setReq(req *http.Request) {
	req.Header.Set("Referer", "https://gofile.io/")
	req.Header.Set("Origin", "https://gofile.io")
	req.Header.Set("User-Agent", ua)
}

var ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
var salt = "5d4f7g8sd45fsd"
var language = "en-US"

// from https://gofile.io/dist/js/wt.obf.js
func (r *Gofile) generateWT() string {
	timestamp := time.Now().Unix() / 14400
	rawString := fmt.Sprintf("%s::%s::%s::%d::%s", ua, language, r.token, timestamp, salt)

	hash := sha256.Sum256([]byte(rawString))
	return fmt.Sprintf("%x", hash)
}

func (r *Gofile) setAuth() error {
	// set token
	var tokenUrl = "https://api.gofile.io/accounts"
	tokenReq, _ := http.NewRequest(http.MethodPost, tokenUrl, nil)
	setReq(tokenReq)
	resp, err := r.client.Do(tokenReq)
	if err != nil {
		return err
	}
	defer utils.SafeClose(resp.Body)
	var gofileToken GofileTokenResp
	err = json.NewDecoder(resp.Body).Decode(&gofileToken)
	if err != nil {
		log.Println(err)
		return err
	}
	r.token = gofileToken.Data.Token
	if r.token == "" {
		return fmt.Errorf("token get failed: %s", resp.Status)
	}
	return nil
}

type GofileTokenResp struct {
	Status string `json:"status"`
	Data   struct {
		Id         string `json:"id"`
		RootFolder string `json:"rootFolder"`
		Tier       string `json:"tier"`
		Token      string `json:"token"`
	} `json:"data"`
}

type GofileContentsResp struct {
	Status string `json:"status"`
	Data   struct {
		Id            string                `json:"id"`
		CanAccess     bool                  `json:"canAccess"`
		Type          string                `json:"type"`
		Name          string                `json:"name"`
		Code          string                `json:"code"` // /d/xxx xxx is code
		TotalSize     int64                 `json:"totalSize"`
		ChildrenCount int64                 `json:"childrenCount"`
		Children      map[string]GofileFile `json:"children"`
	} `json:"data"`
	Metadata struct {
		TotalCount  int  `json:"totalCount"`
		TotalPages  int  `json:"totalPages"`
		Page        int  `json:"page"`
		PageSize    int  `json:"pageSize"`
		HasNextPage bool `json:"hasNextPage"`
	} `json:"metadata"`
}
type GofileFile struct {
	Id            string   `json:"id"`
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	Size          int64    `json:"size"`
	Md5           string   `json:"md5"`
	MimeType      string   `json:"mimetype"`
	Servers       []string `json:"servers"`
	SeverSelected string   `json:"severSelected"`
	Link          string   `json:"link"`
}

func (r *Gofile) getFiles() *GofileContentsResp {

	var contentsApi = "https://api.gofile.io/contents/" + r.code
	var params = url.Values{
		"contentFilter": {""},
		"page":          {"1"},
		"pageSize":      {"1000"},
		"sortField":     {"name"},
		"sortDirection": {"1"},
	}
	if r.password != "" {
		sumBytes := sha256.Sum256([]byte(r.password))
		params.Set("password", hex.EncodeToString(sumBytes[:]))
	}
	req, _ := http.NewRequest(http.MethodGet, contentsApi+"?"+params.Encode(), nil)

	setReq(req)
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("X-Website-Token", r.generateWT())
	req.Header.Set("X-BL", language)

	resp, err := r.client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer utils.SafeClose(resp.Body)

	var contents GofileContentsResp
	err = json.NewDecoder(resp.Body).Decode(&contents)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &contents
}
