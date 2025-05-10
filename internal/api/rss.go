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
		return fmt.Errorf("add feed status: %s", resp.Status)
	}

	return nil
}

func RssRuleList() map[string]*RssRule {
	c, err := GetQbitClient()
	if err != nil {
		printApiClientError("RssRuleList", err)
		return nil
	}

	resp, err := c.Get("/api/v2/rss/rules", url.Values{})
	if err != nil {
		printApiGetError("RssRuleList", err)
		return nil
	}
	defer resp.Body.Close()

	var results map[string]*RssRule
	if err := c.ParseJSON(resp, &results); err != nil {
		printApiParsJSONError("RssRuleList", err)
		return nil
	}
	return results
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

	_, er := c.Post("/api/v2/rss/setRule", params)
	if er != nil {
		return err
	}

	return nil
}

func RssAllItems(withData bool) map[string]RssSub {
	c, err := GetQbitClient()
	if err != nil {
		printApiClientError("RssAllItems", err)
		return nil
	}
	resp, err := c.Get("/api/v2/rss/items", url.Values{"withData": {strconv.FormatBool(withData)}})
	if err != nil {
		printApiGetError("RssAllItems", err)
		return nil
	}
	defer resp.Body.Close()

	var results map[string]RssSub
	if err := c.ParseJSON(resp, &results); err != nil {
		printApiParsJSONError("RssAllItems", err)
		return nil
	}
	return results
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
		return fmt.Errorf("remove rss sub status: %s", resp.Status)
	}
	return nil
}
