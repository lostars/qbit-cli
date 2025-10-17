package job

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qbit-cli/internal/api"
	"qbit-cli/internal/config"
	"qbit-cli/pkg/utils"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Bunkr struct {
	maxWorker, albumMaxWorker int
	savePath                  string
	client                    *http.Client
	url                       string
}

var bunkr Bunkr

func (r *Bunkr) JobName() string {
	return "bunkr"
}

func (r *Bunkr) Description() string {
	return `Resolve Bunkr file share url and download.`
}

func (_ *Bunkr) Tags() []string {
	return []string{"resolver"}
}

func init() {
	api.RegisterJob(&Bunkr{})
}

func (r *Bunkr) RunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bunkr <url>",
		Short: "Resolve Bunkr file share url and download.",
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&bunkr.savePath, "save-path", "", "file save path, default is current working directory")
	cmd.Flags().IntVar(&bunkr.maxWorker, "max-worker", 2, "max worker number, if u get too many requests error, set it smaller or 1")
	cmd.Flags().IntVar(&bunkr.albumMaxWorker, "album-max-worker", 2, "max album worker number, if u get 503 error, set it smaller or 1")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		path := args[0]
		if strings.Contains(path, "/f/") {
			// file share
			bunkr.url = path
			bunkr.client = &http.Client{Timeout: time.Minute * 60}
			err := bunkr.download()
			if err != nil {
				return err
			}
		} else if strings.Contains(path, "/a/") {
			// album share
			bunkr.url = path
			bunkr.client = &http.Client{Timeout: time.Minute * 60}
			bunkr.albumDownload()
		} else {
			return errors.New("unknown url")
		}

		return nil
	}

	return cmd
}

func (r *Bunkr) albumDownload() {
	parsedAlbum, err := url.Parse(r.url)
	if err != nil {
		return
	}

	req, _ := http.NewRequest(http.MethodGet, r.url, nil)
	resp, err := r.client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("%s: %s", r.url, resp.Status)
		return
	}

	urls := findHtmlHrefNode(resp.Body)
	var files []string
	host := parsedAlbum.Scheme + "://" + parsedAlbum.Host
	// max page 1000
	var pages = make([]bool, 1002)
	for _, a := range urls {
		if strings.HasPrefix(a, "/f/") {
			files = append(files, host+a)
		} else if strings.HasPrefix(a, "?page") {
			results := regexp.MustCompile(`\?page=(\d+)`).FindStringSubmatch(a)
			if len(results) < 2 {
				log.Println("unknown page: ", a)
				continue
			}

			page, err := strconv.Atoi(results[1])
			if err != nil {
				log.Println("parse page error: ", a)
				continue
			}

			if page >= len(pages) {
				panic("max page exceeded")
			}
			if pages[page] {
				continue
			}

			// parse page
			req, _ := http.NewRequest(http.MethodGet, host+a, nil)
			resp, err := r.client.Do(req)
			if err != nil {
				log.Println(err)
			}
			if resp.StatusCode != http.StatusOK {
				log.Printf("%s: %s", r.url, resp.Status)
				resp.Body.Close()
				continue
			}

			urls := findHtmlHrefNode(resp.Body)
			for _, fileUrl := range urls {
				if strings.HasPrefix(a, "/f/") {
					files = append(files, fileUrl)
				}
			}
			resp.Body.Close()

			pages[page] = true

		}
	}

	fmt.Println("total files: ", len(files))

	var wg sync.WaitGroup
	limit := make(chan struct{}, r.albumMaxWorker)
	for i := range files {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			limit <- struct{}{}

			b := *r
			b.url = files[i]
			err = b.download()
			if err != nil {
				log.Println(err)
			}

			<-limit
		}(i)
	}
	wg.Wait()
}

func findHtmlHrefNode(body io.Reader) []string {
	z := html.NewTokenizer(body)
	var links []string
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		default:
		}
	}
}

func (r *Bunkr) download() error {
	parsedUrl, err := url.Parse(r.url)
	if err != nil {
		return err
	}
	slug := filepath.Base(parsedUrl.Path)
	log.Printf("slug: %s\n", slug)
	if slug == "" {
		return errors.New("invalid url")
	}

	filename, err := getShareFilename(r.client, r.url)
	if err != nil {
		return err
	}
	log.Printf("filename: %s\n", filename)

	// get encrypted url
	apiHost := parsedUrl.Scheme + "://" + parsedUrl.Host + "/api/vs"
	encryptedFile, err := getEncryptFile(r.client, slug, apiHost, r.url)
	if err != nil {
		return err
	}
	fullPath := filepath.Join(r.savePath, filename)
	start := time.Now()

	downloader := utils.HttpFileDownloader{
		SavePath:  r.savePath,
		MaxWorker: r.maxWorker,
		Headers: map[string]string{
			"Referer": downloadHost,
		},
	}
	if config.Debug {
		downloader.DebugLogger = log.New(os.Stdout, "[HttpFileDownloader]: ", log.LstdFlags)
	}

	if encryptedFile.Encrypted {

		decrypted, err := decryptBunkrFile(encryptedFile, filename)
		if err != nil {
			return err
		}
		log.Printf("decrypted url: %s\n", decrypted)

		err = downloader.Download(decrypted, filename)
		if err != nil {
			return err
		}
	} else {
		err = downloader.Download(r.url, filename)
		if err != nil {
			return err
		}
	}
	cost := time.Now().Sub(start)
	fmt.Printf("file saved to: %s, cost: %s\n", fullPath, cost.String())
	return nil
}

func getEncryptFile(client *http.Client, slug string, apiHost string, referer string) (*EncryptedBunkrFile, error) {
	params := map[string]string{
		"slug": slug,
	}
	b, _ := json.Marshal(params)
	req, _ := http.NewRequest(http.MethodPost, apiHost, bytes.NewBuffer(b))
	req.Header.Set("referer", referer)
	req.Header.Set("content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	var encryptedFile EncryptedBunkrFile
	_ = json.NewDecoder(resp.Body).Decode(&encryptedFile)
	return &encryptedFile, nil
}

func decryptBunkrFile(encrypted *EncryptedBunkrFile, filename string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted.Url)
	if err != nil {
		return "", err
	}

	key := "SECRET_KEY_" + strconv.FormatInt(encrypted.Timestamp/3600, 10)

	keyBytes := []byte(key)
	decrypted := make([]byte, len(data))

	for i := 0; i < len(data); i++ {
		decrypted[i] = data[i] ^ keyBytes[i%len(keyBytes)]
	}

	decryptedUrl := string(decrypted)
	separator := ""
	if strings.Contains(decryptedUrl, "?") {
		separator = "&"
	} else {
		separator = "?"
	}
	decryptedUrl += separator + "n=" + url.QueryEscape(filename)

	return decryptedUrl, nil
}

type EncryptedBunkrFile struct {
	Encrypted bool   `json:"encrypted"`
	Timestamp int64  `json:"timestamp"`
	Url       string `json:"url"`
}

var downloadHost = "https://get.bunkrr.su"

func getShareFilename(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`property="og:title" content="(.*)"`)
	fileHtml := string(b)
	matches := re.FindStringSubmatch(fileHtml)
	if matches == nil || len(matches) != 2 {
		log.Printf("unexpected response body: %s", fileHtml)
		return "", errors.New("can't parse download url")
	}

	return matches[1], nil
}
