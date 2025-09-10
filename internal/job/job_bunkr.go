package job

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qbit-cli/internal/api"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Bunkr struct {
	maxWorker int
	savePath  string
	client    *http.Client
	url       string
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
			return errors.New("album not supported yet")
		} else {
			return errors.New("unknown url")
		}

		return nil
	}

	return cmd
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
	if encryptedFile.Encrypted {

		decrypted, err := decryptBunkrFile(encryptedFile, filename)
		if err != nil {
			return err
		}
		log.Printf("decrypted url: %s\n", decrypted)

		err = downloadFile(r.client, decrypted, fullPath)
		if err != nil {
			return err
		}

	} else {
		err = downloadFile(r.client, r.url, fullPath)
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

func rangeDownload(url string, file *os.File, total, chunkSize int64) {
	if err := file.Truncate(total); err != nil {
		fmt.Println(err)
		return
	}

	tasks := make(chan Task, bunkr.maxWorker)
	var wg sync.WaitGroup

	for i := 0; i < bunkr.maxWorker; i++ {
		wg.Add(1)
		go downloadChunk(i, url, file, tasks, &wg)
	}

	// load metadata
	var metadata map[int64]bool
	metaFile, err := os.Open(file.Name() + metadataSuffix)
	if err != nil {
		log.Println("download metadata not found")
		metadata = make(map[int64]bool, total/chunkSize+1)
	} else {
		err = json.NewDecoder(metaFile).Decode(&metadata)
		if err != nil {
			log.Println("wrong format of metadata")
			metadata = make(map[int64]bool, total/chunkSize+1)
		}
		metaFile.Close()
	}

	lock := &sync.Mutex{}

	chunkIndex := int64(0)
	for start := int64(0); start < total; start += chunkSize {
		end := start + chunkSize - 1
		if end >= total {
			end = total - 1
		}
		tasks <- Task{
			start: start, end: end,
			metadata: &metadata, lock: lock,
			chunkIndex: chunkIndex,
		}
		chunkIndex++
	}
	close(tasks)
	wg.Wait()
}

type Task struct {
	start, end int64
	chunkIndex int64
	metadata   *map[int64]bool
	lock       *sync.Mutex
}

func (t *Task) saveMetadata(metaFile string) {
	t.lock.Lock()

	meta := *t.metadata
	meta[t.chunkIndex] = true
	t.metadata = &meta

	// save to file
	f, _ := os.Create(metaFile)
	defer f.Close()
	_ = json.NewEncoder(f).Encode(meta)

	t.lock.Unlock()
}

func (t *Task) downloadCompleted() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	meta := *t.metadata
	return meta[t.chunkIndex]
}

var metadataSuffix = ".metadata"

func downloadChunk(id int, url string, file *os.File, tasks <-chan Task, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		if task.downloadCompleted() {
			log.Printf("piece: %d downloaded\n", task.chunkIndex)
			continue
		}
		log.Printf("thread %d starting...\n", id)

		for {
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("referer", downloadHost)
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", task.start, task.end))

			client := &http.Client{Timeout: time.Minute * 60}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				time.Sleep(time.Second << 2)
				continue
			}

			if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
				fmt.Println("unexpected status: ", resp.Status)
				resp.Body.Close()
				continue
			}

			writeRespToFile(resp, file, task.start)
			resp.Body.Close()
			task.saveMetadata(file.Name() + metadataSuffix)
			break
		}

	}

}

func writeRespToFile(resp *http.Response, file *os.File, begin int64) {
	buf := make([]byte, 32*1024)
	start := begin
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err := file.WriteAt(buf[:n], start)
			if err != nil {
				continue
			}
			start += int64(n)
		}
		if err == io.EOF {
			break
		}
	}
}

var chunkSize int64 = 10 * 1024 * 1024

func downloadFile(client *http.Client, url string, file string) error {
	// check file if exists
	_, err := os.Stat(file)
	if err == nil {
		return errors.New("file exists: " + file)
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("referer", downloadHost)

	req.Header.Set("Range", "bytes=0-0")
	headResp, err := client.Do(req)
	rangeD := false
	size := int64(0)
	if err == nil {
		defer headResp.Body.Close()
		// Content-Range bytes 0-0/397540573
		contentRange := headResp.Header.Get("Content-Range")
		if contentRange != "" {
			_, err = fmt.Sscanf(contentRange, "bytes 0-0/%d", &size)
			if err != nil {
				log.Println("failed to parse Content-Length, fallback to single thread download")
			} else {
				if size > chunkSize {
					rangeD = true
				} else {
					log.Println("file too small, fallback to single thread download")
				}
			}
		}
	} else {
		log.Println(err)
	}

	tmpFile := file + ".downloading"
	if rangeD {
		out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		rangeDownload(url, out, size, chunkSize)
	} else {
		out, err := os.Create(tmpFile)
		if err != nil {
			return err
		}
		req.Header.Del("Range")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	}

	// rename file
	_ = os.Rename(tmpFile, file)
	// rm metadata file
	os.Remove(tmpFile + metadataSuffix)

	return nil
}

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
