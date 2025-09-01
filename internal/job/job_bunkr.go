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
	"time"
)

type Bunkr struct{}

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

	var savePath string

	cmd.Flags().StringVar(&savePath, "save-path", "", "file save path, default is current working directory")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		u := args[0]
		parsedUrl, err := url.Parse(u)
		if err != nil {
			return err
		}

		slug := filepath.Base(parsedUrl.Path)
		log.Printf("slug: %s\n", slug)
		if slug == "" {
			return errors.New("invalid url")
		}
		client := &http.Client{Timeout: time.Minute * 60}

		filename, err := getShareFilename(client, u)
		if err != nil {
			return err
		}
		log.Printf("filename: %s\n", filename)

		// get encrypted url
		apiHost := parsedUrl.Scheme + "://" + parsedUrl.Host + "/api/vs"
		encryptedFile, err := getEncryptFile(client, slug, apiHost, u)
		log.Println(encryptedFile)
		fullPath := filepath.Join(savePath, filename)
		if encryptedFile.Encrypted {

			decrypted, err := decryptBunkrFile(encryptedFile, filename)
			if err != nil {
				return err
			}
			log.Printf("decrypted url: %s\n", decrypted)

			err = downloadFile(client, decrypted, fullPath)
			if err != nil {
				return err
			}

		} else {
			err = downloadFile(client, u, fullPath)
			if err != nil {
				return err
			}
		}
		fmt.Printf("file saved to: %s\n", fullPath)

		return nil
	}

	return cmd
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

func downloadFile(client *http.Client, url string, file string) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("referer", downloadHost)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
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
