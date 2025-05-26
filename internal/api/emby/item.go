package emby

import (
	"errors"
	"net/http"
	"net/url"
	"qbit-cli/internal/api"
)

func Items(params url.Values) (*api.EmbyItems, error) {
	embyClient := api.GetEmbyClient()
	resp, err := embyClient.Get(embyClient.EmbyUserEndpoint("Items"), params)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return &api.EmbyItems{}, nil
		} else {
			return nil, errors.New(resp.Status)
		}
	}
	var result *api.EmbyItems
	if err := api.ParseJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func Item(item string) (*api.EmbyItem, error) {
	embyClient := api.GetEmbyClient()
	resp, err := embyClient.Get(embyClient.EmbyUserEndpoint("Items", item), url.Values{})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	var result api.EmbyItem
	if err := api.ParseJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
