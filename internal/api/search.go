package api

import (
	"errors"
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
		return result, errors.New("user has reached the limit of max Running searches (currently set to 5)")
	}

	if err := client.ParseJSON(resp, &result); err != nil {
		return result, err
	}
	return result, nil
}

// SearchDetails get all search results, slow(may take seconds)
// Attention: you must use the same auth information to start search and get results.
// Or you will get a 404 from /api/v2/search/results
func SearchDetails(d time.Duration, resultID uint32, params url.Values) []SearchDetail {
	client, err := GetQbitClient()
	if err != nil {
		printApiClientError("SearchDetails", err)
		return nil
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
			break
		}

		var result SearchResults
		err = client.ParseJSON(resp, &result)
		resp.Body.Close()
		if err != nil {
			continue
		}

		status = result.Status
		offset = result.Total
		if result.Total > 0 && len(result.Results) > 0 {
			details = append(details, result.Results...)
		}
	}

	return details
}

func SearchPlugins() *[]SearchPlugin {
	c, err := GetQbitClient()
	if err != nil {
		printApiClientError("SearchPlugins", err)
		return nil
	}

	resp, err := c.Get("/api/v2/search/plugins", nil)
	if err != nil {
		printApiGetError("SearchPlugins", err)
		return nil
	}
	defer resp.Body.Close()

	var result []SearchPlugin
	if err := c.ParseJSON(resp, &result); err != nil {
		printApiParsJSONError("SearchPlugins", err)
		return nil
	}

	return &result
}
