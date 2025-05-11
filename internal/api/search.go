package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// all the /search/* api here

func SearchStart(params url.Values) (SearchResult, error) {
	result := SearchResult{}
	client, err := GetQbitClient()
	if err != nil {
		return result, err
	}

	resp, err := client.Post("/api/v2/search/start", params)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return result, &QbitClientError{resp.Status, "SearchStart", nil}
	}

	if err := ParseJSON(resp, &result); err != nil {
		return result, err
	}
	return result, nil
}

// SearchDetails get all search results, slow(may take seconds)
// Attention: you must use the same auth information to start search and get results.
// Or you will get a 404 from /api/v2/search/results
func SearchDetails(d time.Duration, resultID uint32, params url.Values) ([]SearchDetail, error) {
	client, err := GetQbitClient()
	if err != nil {
		return nil, err
	}
	status := "Running"
	var details []SearchDetail
	for status == "Running" {
		time.Sleep(d)

		var offset uint32 = 0
		params = make(url.Values)
		params.Set("id", strconv.FormatUint(uint64(resultID), 10))
		params.Set("limit", strconv.Itoa(500))
		params.Set("offset", strconv.FormatUint(uint64(offset), 10))

		resp, err := client.Get("/api/v2/search/results", params)
		if err != nil {
			fmt.Printf(err.Error())
			break
		}

		var result SearchResults
		err = ParseJSON(resp, &result)
		resp.Body.Close()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		status = result.Status
		offset = result.Total
		if result.Total > 0 && len(result.Results) > 0 {
			details = append(details, result.Results...)
		}
	}

	return details, nil
}

func SearchPlugins() (*[]SearchPlugin, error) {
	c, err := GetQbitClient()
	if err != nil {
		return nil, err
	}

	resp, err := c.Get("/api/v2/search/plugins", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []SearchPlugin
	if err := ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
