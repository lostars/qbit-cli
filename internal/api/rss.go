package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// all /rss/* api here

func RssAddSub(feedUrl string, path string) error {
	params := url.Values{
		"url":  {feedUrl},
		"path": {path},
	}

	resp, err := GetQbitClient().Post("/api/v2/rss/addFeed", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "RssAddSub", nil}
	}

	return nil
}

func RssRuleList() (map[string]*RssRule, error) {
	resp, err := GetQbitClient().Get("/api/v2/rss/rules", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results map[string]*RssRule
	if err := ParseJSON(resp, &results); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return results, nil
}

func RssSetRule(ruleName string, rule *RssRule) error {
	j, err := json.Marshal(rule)
	if err != nil {
		return err
	}

	params := url.Values{
		"name":    {ruleName},
		"ruleDef": {string(j)},
	}

	resp, er := GetQbitClient().Post("/api/v2/rss/setRule", params)
	if er != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "RssSetRule", nil}
	}

	return nil
}

func RssAllItems(withData bool) (map[string]RssSub, error) {
	resp, err := GetQbitClient().Get("/api/v2/rss/items", url.Values{"withData": {strconv.FormatBool(withData)}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results map[string]RssSub
	if err := ParseJSON(resp, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func RssRmSub(path string) error {
	resp, err := GetQbitClient().Post("/api/v2/rss/removeItem", url.Values{"path": {path}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "RssRmSub", nil}
	}
	return nil
}
