package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Gofile struct {
	maxWorker int
	savePath  string
	client    *http.Client
	url, code string
	wt, token string
}

var gofile Gofile

func (r *Gofile) JobName() string {
	return "gofile"
}

func (r *Gofile) Description() string {
	return `Resolve Gofile file share url and download.`
}

func (_ *Gofile) Tags() []string {
	return []string{"resolver"}
}

func init() {
	api.RegisterJob(&Gofile{})
}

func (r *Gofile) RunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gofile <url>",
		Short: "Resolve Gofile file share url and download.",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&gofile.savePath, "save-path", "", "file save path, default is current working directory")
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
		gofile.client = &http.Client{Timeout: time.Minute * 60}

		/*

			1. Get token from https://api.gofile.io/accounts
			2. Get wt(a hard code string) from https://gofile.io/dist/js/global.js
			3. Get file list from https://api.gofile.io/contents/xxx?wt=yyy
			4. Download files

		*/

		gofile.setAuth()

		var contents = gofile.getFiles()
		if contents == nil {
			return errors.New("failed to get files")
		}
		if contents.Status != "ok" {
			return errors.New("contents return " + contents.Status)
		}
		if len(contents.Data.Children) <= 0 {
			return errors.New("contents is empty")
		}

		downloader := utils.HttpFileDownloader{
			SavePath:  gofile.savePath,
			MaxWorker: gofile.maxWorker,
			Headers: map[string]string{
				"Authorization": "Bearer " + gofile.token,
			},
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

func (r *Gofile) setAuth() {

	// set wt
	var wtJS = "https://gofile.io/dist/js/global.js"
	req, _ := http.NewRequest(http.MethodGet, wtJS, nil)
	resp, err := r.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	content := string(body)

	match := regexp.MustCompile(`appdata\.wt.*"(.*)"`).FindStringSubmatch(content)
	if len(match) <= 1 {
		log.Println("wt failed to get")
		return
	}
	r.wt = match[1]
	log.Printf("wt: %s\n", r.wt)

	// set token
	var tokenUrl = "https://api.gofile.io/accounts"
	tokenReq, _ := http.NewRequest(http.MethodPost, tokenUrl, nil)
	resp, err = r.client.Do(tokenReq)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	var gofileToken GofileTokenResp
	err = json.NewDecoder(resp.Body).Decode(&gofileToken)
	if err != nil {
		log.Println(err)
		return
	}
	r.token = gofileToken.Data.Token
	if r.token == "" {
		log.Println("token get failed")
	}
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
		"wt":            {r.wt},
		"page":          {"1"},
		"pageSize":      {"1000"},
		"sortField":     {"name"},
		"sortDirection": {"1"},
	}
	req, _ := http.NewRequest(http.MethodGet, contentsApi+"?"+params.Encode(), nil)
	req.Header.Set("Authorization", "Bearer "+r.token)
	resp, err := r.client.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	var contents GofileContentsResp
	err = json.NewDecoder(resp.Body).Decode(&contents)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &contents
}
