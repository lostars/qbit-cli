package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

var client *QbitClient

func GetQbitClient() (*QbitClient, error) {
	if client != nil {
		return client, nil
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	client = &QbitClient{
		needAuth: cfg.Server.Username != "" && cfg.Server.Password != "",
		Config:   cfg,
		Headers:  make(map[string]string),
		Client:   &http.Client{Timeout: time.Second * 10},
	}
	return client, nil
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

func ParseJSON(resp *http.Response, v any) error {
	err := json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

func (c *QbitClient) Get(endpoint string, params url.Values) (*http.Response, error) {
	if err := c.login(); err != nil {
		return nil, err
	}

	fullUrl := c.host() + endpoint
	if params != nil && len(params) > 0 {
		fullUrl += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", fullUrl, nil)
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
	if err := c.login(); err != nil {
		return nil, err
	}

	fullUrl := c.host() + endpoint
	req, err := http.NewRequest("POST", fullUrl, strings.NewReader(params.Encode()))
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

func (c *QbitClient) login() error {

	if c.Headers["Cookie"] != "" {
		return nil
	}

	body := url.Values{
		"username": {c.user()},
		"password": {c.pwd()},
	}

	resp, err := c.Client.PostForm(c.host()+"/api/v2/auth/login", body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{"login failed", "login", nil}
	}

	cookie, _, _ := strings.Cut(resp.Header.Get("Set-Cookie"), ";")
	if cookie == "" {
		return &QbitClientError{"login success, but cookie is empty", "login", nil}
	}

	c.Headers["Cookie"] = cookie

	return nil
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
