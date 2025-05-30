package api

import (
	"net/url"
	"strings"
)

func JackettSearch(indexer string, category []string, query string) (*JackettResults, error) {
	params := url.Values{
		"Query":    []string{query},
		"Category": category,
	}
	endpoint := "/api/v2.0/indexers/_/results"
	endpoint = strings.Replace(endpoint, "_", indexer, 1)
	resp, err := GetJackettClient().Get(endpoint, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result JackettResults
	err = ParseJSON(resp, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func JackettIndexers(enabled bool) (*[]JackettIndexer, error) {
	resp, err := GetJackettClient().Get("/api/v2.0/indexers", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result []JackettIndexer
	err = ParseJSON(resp, &result)
	if err != nil {
		return nil, err
	}

	if !enabled {
		return &result, nil
	}

	var data = make([]JackettIndexer, 0, len(result))
	for _, indexer := range result {
		if indexer.Configured {
			data = append(data, indexer)
		}
	}
	return &data, nil
}
