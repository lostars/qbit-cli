package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"qbit-cli/internal/config"
	"strings"
	"time"
)

type QbitClient struct {
	needAuth bool
	Config   *config.Config
	Client   *http.Client
	Headers  map[string]string
}

var serverInfo *QbitServerInfo

func GetQbitServerInfo() *QbitServerInfo {
	if serverInfo != nil {
		return serverInfo
	}
	info, err := QbitAppBuildInfo()
	if err != nil {
		panic(err)
	}
	appVersion, e := QbitAppVersion()
	if e != nil {
		panic(e)
	}
	apiVersion, er := QbitApiVersion()
	if er != nil {
		panic(er)
	}
	info.AppVersion = appVersion
	info.WebApiVersion = apiVersion
	serverInfo = info
	return serverInfo
}

var client *QbitClient

func GetQbitClient() *QbitClient {
	if client != nil {
		client.login()
		return client
	}

	cfg := config.GetConfig()

	client = &QbitClient{
		needAuth: cfg.Server.Username != "" && cfg.Server.Password != "",
		Config:   cfg,
		Headers:  make(map[string]string),
		Client:   &http.Client{Timeout: time.Second * 10},
	}
	client.login()
	return client
}

func (c *QbitClient) host() string {
	return c.Config.Server.Host
}
func (c *QbitClient) user() string {
	return c.Config.Server.Username
}
func (c *QbitClient) pwd() string {
	return c.Config.Server.Password
}

type JackettClient struct {
	Config *config.Config
	Client *http.Client
}

var jackettClient *JackettClient

func GetJackettClient() *JackettClient {
	if jackettClient != nil {
		return jackettClient
	}
	jackettClient = &JackettClient{
		Config: config.GetConfig(),
		Client: &http.Client{Timeout: time.Second * 10},
	}
	return jackettClient
}

var jackettAuthEndpoint = []string{
	"/api/v2.0/indexers",
}

func (c *JackettClient) Get(endpoint string, params url.Values) (*http.Response, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("apikey", c.Config.Jackett.ApiKey)
	fullUrl := c.Config.Jackett.Host + endpoint
	fullUrl += "?" + params.Encode()
	log.Println(fullUrl)

	req, err := http.NewRequest(http.MethodGet, fullUrl, nil)
	if err != nil {
		return nil, &HTTPClientError{"Get", fullUrl, err}
	}
	for _, e := range jackettAuthEndpoint {
		if e == endpoint {
			req.Header.Set("Cookie", c.Config.Jackett.Cookie)
			break
		}
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type EmbyClient struct {
	Config *config.Config
	Client *http.Client
}

var embyClient *EmbyClient

func GetEmbyClient() *EmbyClient {
	if embyClient != nil {
		return embyClient
	}
	embyClient = &EmbyClient{
		Config: config.GetConfig(),
		Client: &http.Client{Timeout: time.Second * 10},
	}

	return embyClient
}

func (c *EmbyClient) embyHost() string {
	return c.Config.Emby.Host
}

func (c *EmbyClient) embyApiKey() string {
	return c.Config.Emby.ApiKey
}

func (c *EmbyClient) EmbyUser() string {
	return c.Config.Emby.User
}

func (c *EmbyClient) EmbyUserEndpoint(endpoint ...string) string {
	str := []string{"/emby/Users", c.EmbyUser()}
	str = append(str, endpoint...)
	return strings.Join(str, "/")
}

func (c *EmbyClient) Get(endpoint string, params url.Values) (*http.Response, error) {
	fullUrl := c.embyHost() + endpoint
	if params == nil {
		params = url.Values{}
	}
	if len(params) > 0 {
		fullUrl += "?" + params.Encode()
	}
	if config.Debug {
		token := ""
		if len(params) > 0 {
			token = "&X-Emby-Token=" + c.embyApiKey()
		} else {
			token = "?X-Emby-Token=" + c.embyApiKey()
		}
		log.Println(fullUrl + token)
	}

	req, err := http.NewRequest(http.MethodGet, fullUrl, nil)
	if err != nil {
		return nil, &HTTPClientError{"Get", fullUrl, err}
	}
	c.embyAuth(req)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *EmbyClient) embyAuth(req *http.Request) {
	req.Header.Set("X-Emby-Token", c.embyApiKey())
	req.Header.Set("X-Emby-Client", "qbit-cli")
	req.Header.Set("X-Emby-Device-Name", "qbit-cli")
}

func (c *EmbyClient) Post(endpoint string, params url.Values) (*http.Response, error) {
	fullUrl := c.embyHost() + endpoint
	req, err := http.NewRequest(http.MethodPost, fullUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.embyAuth(req)
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, &HTTPClientError{"Post", fullUrl, err}
	}

	return resp, nil
}

func ParseJSON(resp *http.Response, v any) error {
	err := json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

func ParseString(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func ParseRawJSON(resp *http.Response) (string, error) {
	raw := json.RawMessage{}
	err := json.NewDecoder(resp.Body).Decode(&raw)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (c *QbitClient) Get(endpoint string, params url.Values) (*http.Response, error) {
	fullUrl := c.host() + endpoint
	if params != nil && len(params) > 0 {
		fullUrl += "?" + params.Encode()
	}
	log.Println(fullUrl)

	req, err := http.NewRequest(http.MethodGet, fullUrl, nil)
	if err != nil {
		return nil, &HTTPClientError{"Get", fullUrl, err}
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *QbitClient) Post(endpoint string, params url.Values) (*http.Response, error) {
	fullUrl := c.host() + endpoint
	req, err := http.NewRequest(http.MethodPost, fullUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, &HTTPClientError{"Post", fullUrl, err}
	}

	return resp, nil
}

func (c *QbitClient) PostForm(endpoint string, params url.Values, fields string, files []*os.File) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, values := range params {
		for _, value := range values {
			_ = writer.WriteField(key, value)
		}
	}

	for _, file := range files {
		part, err := writer.CreateFormFile(fields, filepath.Base(file.Name()))
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.host()+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *QbitClient) login() {

	if c.Headers["Cookie"] != "" {
		return
	}

	body := url.Values{
		"username": {c.user()},
		"password": {c.pwd()},
	}

	resp, err := c.Client.PostForm(c.host()+"/api/v2/auth/login", body)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic("login failed")
	}

	cookie, _, _ := strings.Cut(resp.Header.Get("Set-Cookie"), ";")
	if cookie == "" {
		panic("login success, but cookie is empty")
	}

	c.Headers["Cookie"] = cookie
}

type HTTPClientError struct {
	message string
	url     string
	err     error
}

type QbitClientError struct {
	message string
	method  string
	err     error
}

func (c *QbitClientError) Error() string {
	errStr := ""
	if c.err != nil {
		errStr = c.err.Error()
	}
	return fmt.Sprintf("%s qbit client error: %s %s", c.method, c.message, errStr)
}

func (e *HTTPClientError) Error() string {
	errStr := ""
	if e.err != nil {
		errStr = e.err.Error()
	}
	return fmt.Sprintf("http client error: %s %s %s", e.url, e.message, errStr)
}
