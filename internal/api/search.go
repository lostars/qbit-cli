package api

import (
	"fmt"
	"net/http"
	"net/url"
	"qbit-cli/pkg/utils"
	"strconv"
	"strings"
	"time"
)

// all the /search/* api here

func SearchStart(params url.Values) (SearchResult, error) {
	result := SearchResult{}
	resp, err := GetQbitClient().Post("/api/v2/search/start", params)
	if err != nil {
		return result, err
	}
	defer utils.SafeClose(resp.Body)

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
func SearchDetails(d time.Duration, resultID uint32) ([]*SearchDetail, error) {
	client := GetQbitClient()
	status := "Running"
	// duplicate removal
	m := make(map[string]SearchDetail)
	for status == "Running" {
		time.Sleep(d)

		params := url.Values{}
		params.Set("id", strconv.FormatUint(uint64(resultID), 10))
		params.Set("limit", strconv.Itoa(500))
		// 通过状态来判断搜索是否结束 不使用offset分页获取结果
		params.Set("offset", "0")

		resp, err := client.Get("/api/v2/search/results", params)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		var result SearchResults
		err = ParseJSON(resp, &result)
		utils.SafeClose(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		status = result.Status
		if result.Total > 0 && len(result.Results) > 0 {
			for _, r := range result.Results {
				m[r.FileName] = r
			}
		}
	}

	details := make([]*SearchDetail, 0, len(m))
	for _, result := range m {
		details = append(details, &result)
	}

	return details, nil
}

func SearchPlugins() (*[]SearchPlugin, error) {
	resp, err := GetQbitClient().Get("/api/v2/search/plugins", nil)
	if err != nil {
		return nil, err
	}
	defer utils.SafeClose(resp.Body)

	var result []SearchPlugin
	if err := ParseJSON(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func UpdatePlugin() error {
	resp, err := GetQbitClient().Get("/api/v2/search/updatePlugins", nil)
	if err != nil {
		return err
	}
	defer utils.SafeClose(resp.Body)
	if resp.StatusCode == http.StatusOK {
		return &QbitClientError{resp.Status, "UpdatePlugin", nil}
	}
	return nil
}

func InstallPlugin(sources []string) error {
	params := url.Values{}
	params.Set("sources", strings.Join(sources, "|"))
	resp, err := GetQbitClient().Get("/api/v2/search/installPlugin", nil)
	if err != nil {
		return err
	}
	defer utils.SafeClose(resp.Body)
	return nil
}

func UninstallPlugin(hashes []string) error {
	params := url.Values{}
	params.Set("names", strings.Join(hashes, "|"))
	resp, err := GetQbitClient().Get("/api/v2/search/uninstallPlugin", nil)
	if err != nil {
		return err
	}
	defer utils.SafeClose(resp.Body)
	return nil
}

func EnablePlugin(name []string, enable bool) error {
	params := url.Values{}
	params.Set("names", strings.Join(name, "|"))
	params.Set("enable", strconv.FormatBool(enable))
	resp, err := GetQbitClient().Post("/api/v2/search/enablePlugin", params)
	if err != nil {
		return err
	}
	defer utils.SafeClose(resp.Body)
	return nil
}
