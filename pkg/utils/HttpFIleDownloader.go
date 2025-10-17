package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HttpFileDownloader struct {
	Client              *http.Client
	Headers             map[string]string
	SavePath            string
	OverwriteExistFile  bool
	ChunkSize           int64
	MaxWorker, MaxRetry int
	InfoLogger          *log.Logger
	DebugLogger         *log.Logger
}

func defaultClient() *http.Client {
	return &http.Client{Timeout: time.Minute * 1}
}

func (d *HttpFileDownloader) buildRequest(url string, method string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	if d.Headers != nil {
		for k, v := range d.Headers {
			req.Header.Set(k, v)
		}
	}
	return req, nil
}

func (d *HttpFileDownloader) init() {
	if d.Client == nil {
		d.Client = defaultClient()
	}
	if d.InfoLogger == nil {
		d.InfoLogger = log.New(os.Stdout, "", 0)
	}
	if d.DebugLogger == nil {
		d.DebugLogger = log.New(io.Discard, "DEBUG: ", log.LstdFlags)
	}
}

type ServerHeadResult struct {
	AcceptRanges               string
	ContentLength              int64
	ContentType                string
	AccessControlAllowOrigin   string
	AccessControlExposeHeaders string
	AccessControlAllowHeaders  string
	ContentDisposition         string
	LastModified               string
	Date                       string
	URL                        string
}

func (d *HttpFileDownloader) head(url string) (*ServerHeadResult, error) {
	d.init()
	headReq, err := d.buildRequest(url, http.MethodHead)
	if err != nil {
		d.DebugLogger.Printf("head request build error: %v\n", err)
		return nil, err
	}
	headResp, err := d.Client.Do(headReq)
	if err != nil {
		d.DebugLogger.Printf("head do request error: %v\n", err)
		return nil, err
	}
	defer headResp.Body.Close()

	var info = ServerHeadResult{
		URL:                        url,
		AcceptRanges:               headResp.Header.Get("Accept-Ranges"),
		ContentType:                headResp.Header.Get("Content-Type"),
		AccessControlAllowOrigin:   headResp.Header.Get("Access-Control-Allow-Origin"),
		AccessControlExposeHeaders: headResp.Header.Get("Access-Control-Expose-Headers"),
		AccessControlAllowHeaders:  headResp.Header.Get("Access-Control-Allow-Headers"),
		ContentDisposition:         headResp.Header.Get("Content-Disposition"),
		Date:                       headResp.Header.Get("Date"),
		LastModified:               headResp.Header.Get("Last-Modified"),
	}

	contentL := headResp.Header.Get("Content-Length")
	if intVal, err := strconv.ParseInt(contentL, 10, 64); err == nil {
		info.ContentLength = intVal
	}

	return &info, nil
}

func (i *ServerHeadResult) ResumeAvailable() bool {
	if i.ContentLength <= 0 {
		return false
	}
	if i.AcceptRanges == "bytes" {
		return true
	}
	return false
}

func (d *HttpFileDownloader) fileExist(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	return false
}

func (d *HttpFileDownloader) SingleThreadDownload(url, newName string) error {
	d.init()

	if d.skipDownload(url, newName) {
		return nil
	}

	var fullName = path.Join(d.SavePath, newName)
	var tmpFile = fullName + tmpFileSuffix
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	req, err := d.buildRequest(url, http.MethodGet)
	if err != nil {
		return err
	}
	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	// rename file
	_ = os.Rename(tmpFile, fullName)

	return nil
}

var defaultChunkSize = 10 * 1024 * 1024
var defaultMaxWorker = 2
var metadataFileSuffix = ".metadata"
var tmpFileSuffix = ".downloading"
var defaultMaxRetry = 5

func (d *HttpFileDownloader) chunkSize() int64 {
	if d.ChunkSize > 0 {
		return d.ChunkSize
	}
	return int64(defaultChunkSize)
}

func (d *HttpFileDownloader) maxWorker() int {
	if d.MaxWorker > 0 {
		return d.MaxWorker
	}
	return defaultMaxWorker
}

func (d *HttpFileDownloader) maxRetry() int {
	if d.MaxRetry > 0 {
		return d.MaxRetry
	}
	return defaultMaxRetry
}

type task struct {
	start, end int64
	chunkIndex int64
}

func (d *HttpFileDownloader) skipDownload(url, newName string) bool {
	if newName == "" {
		paths := strings.Split(url, "/")
		newName = paths[len(paths)-1]
	}
	var fullName = path.Join(d.SavePath, newName)
	if d.fileExist(fullName) && !d.OverwriteExistFile {
		d.InfoLogger.Printf("[%s] exists, %s skipped\n", fullName, url)
		return true
	}
	return false
}

func (d *HttpFileDownloader) Download(url, newName string) error {
	d.init()

	if d.skipDownload(url, newName) {
		return nil
	}

	headInfo, err := d.head(url)
	if err != nil {
		return err
	}
	if !headInfo.ResumeAvailable() {
		d.DebugLogger.Println("fallback to single thread download")
		return d.SingleThreadDownload(url, newName)
	} else {
		d.DebugLogger.Println("server support range download")
	}

	// range download
	var fullName = path.Join(d.SavePath, newName)
	var tmpFile = fullName + tmpFileSuffix
	var total = headInfo.ContentLength
	file, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := file.Truncate(total); err != nil {
		d.DebugLogger.Printf("truncate file error: %v\n", err)
		return err
	}

	// handle metadata
	var metadataLock sync.Mutex
	var metadataFile = fullName + metadataFileSuffix
	metadata := make(map[int64]bool)
	if metaBytes, err := os.ReadFile(metadataFile); err == nil {
		_ = json.Unmarshal(metaBytes, &metadata)
	}
	d.DebugLogger.Printf("metadata loaded: %v\n", len(metadata))
	saveMetadata := func(chunkIndex int64) error {
		metadataLock.Lock()
		defer metadataLock.Unlock()

		metadata[chunkIndex] = true
		metaBytes, err := json.Marshal(metadata)
		if err != nil {
			d.DebugLogger.Printf("failed to marshal metadata: %v\n", err)
			return err
		}
		return os.WriteFile(metadataFile, metaBytes, 0644)
	}
	isChunkDownloaded := func(chunkIndex int64) bool {
		metadataLock.Lock()
		defer metadataLock.Unlock()
		return metadata[chunkIndex]
	}

	tasks := make(chan task, d.maxWorker())
	var wg sync.WaitGroup
	for i := 0; i < d.maxWorker(); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			for t := range tasks {
				if isChunkDownloaded(t.chunkIndex) {
					d.DebugLogger.Printf("chunk: %d downloaded\n", t.chunkIndex)
					continue
				}
				d.InfoLogger.Printf("worker-%d starting...\n", i)

				var success = false
				for attempt := 0; attempt < d.maxRetry(); attempt++ {

					req, err := d.buildRequest(url, http.MethodGet)
					if err != nil {
						d.DebugLogger.Printf("worker-%d: failed to create request for chunk %d, aborting chunk: %v\n", i, t.chunkIndex, err)
						break
					}
					req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", t.start, t.end))
					resp, err := d.Client.Do(req)
					if err != nil {
						d.DebugLogger.Printf("worker-%d: failed to do request for chunk %d: %v\n", i, t.chunkIndex, err)
						continue
					}

					if resp.StatusCode == http.StatusTooManyRequests {
						resp.Body.Close()
						time.Sleep(time.Second << 2)
						d.DebugLogger.Printf("unexpected status: %v\n", resp.Status)
						continue
					}
					if resp.StatusCode >= 400 && resp.StatusCode < 500 {
						d.DebugLogger.Printf("stop download, unexpected client status: %v\n", resp.Status)
						resp.Body.Close()
						break
					}
					if resp.StatusCode != http.StatusPartialContent {
						d.DebugLogger.Printf("unexpected status: %v\n", resp.Status)
						resp.Body.Close()
						continue
					}

					err = writeParts(resp, file, t.start)
					resp.Body.Close()
					if err != nil {
						d.DebugLogger.Printf("worker-%d: attempt %d, failed to write chunk %d to file: %v\n", i, attempt+1, t.chunkIndex, err)
						continue
					}

					err = saveMetadata(t.chunkIndex)
					if err != nil {
						d.DebugLogger.Printf("worker-%d: attempt %d, failed to save metadata for chunk %d: %v\n", i, attempt+1, t.chunkIndex, err)
						continue
					}

					success = true
					break
				}
				if !success {
					d.DebugLogger.Printf("worker-%d: failed to download chunk %d after %d attempts\n", i, t.chunkIndex, d.maxRetry())
				}

			}

		}(i)
	}

	// create tasks
	go func() {
		chunkIndex := int64(0)
		for start := int64(0); start < total; start += d.chunkSize() {
			end := start + d.chunkSize() - 1
			if end >= total {
				end = total - 1
			}
			tasks <- task{
				start:      start,
				end:        end,
				chunkIndex: chunkIndex,
			}
			chunkIndex++
		}
		close(tasks)
	}()

	wg.Wait()

	// check all chunks
	totalChunks := (total + d.chunkSize() - 1) / d.chunkSize()
	if int64(len(metadata)) < totalChunks {
		return fmt.Errorf("download failed: not all chunks were completed. Expected %d, got %d", totalChunks, len(metadata))
	}

	// rename file
	if err = os.Rename(tmpFile, fullName); err != nil {
		return fmt.Errorf("failed to rename tmp file: %w", err)
	}
	// rm metadata file
	if err = os.Remove(metadataFile); err != nil {
		d.DebugLogger.Printf("warning: failed to remove metadata file: %v\n", err)
	}

	return nil
}

func writeParts(resp *http.Response, file *os.File, begin int64) error {
	start := begin
	_, err := io.Copy(&WriterAt{writer: file, off: start}, resp.Body)
	if err != nil {
		return fmt.Errorf("write to file failed: %w", err)
	}
	return nil
}

type WriterAt struct {
	writer io.WriterAt
	off    int64
}

func (w *WriterAt) Write(p []byte) (n int, err error) {
	n, err = w.writer.WriteAt(p, w.off)
	w.off += int64(n)
	return
}
