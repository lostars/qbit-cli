package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// all /rss/* api here

func RssAddFeed(feedUrl string, path string) error {
	params := url.Values{
		"url":  {feedUrl},
		"path": {path},
	}

	c, err := GetQbitClient()
	if err != nil {
		return err
	}

	resp, err := c.Post("/api/v2/rss/addFeed", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "RssAddFeed", nil}
	}

	return nil
}

func RssRuleList() (map[string]*RssRule, error) {
	c, err := GetQbitClient()
	if err != nil {
		return nil, err
	}

	resp, err := c.Get("/api/v2/rss/rules", url.Values{})
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

	c, err := GetQbitClient()
	if err != nil {
		return err
	}

	resp, er := c.Post("/api/v2/rss/setRule", params)
	if er != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "RssSetRule", nil}
	}

	return nil
}

func RssAllItems(withData bool) (map[string]RssSub, error) {
	c, err := GetQbitClient()
	if err != nil {
		return nil, err
	}
	resp, err := c.Get("/api/v2/rss/items", url.Values{"withData": {strconv.FormatBool(withData)}})
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
	c, err := GetQbitClient()
	if err != nil {
		return err
	}
	resp, err := c.Post("/api/v2/rss/removeItem", url.Values{"path": {path}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &QbitClientError{resp.Status, "RssRmSub", nil}
	}
	return nil
}
