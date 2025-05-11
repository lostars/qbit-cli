package emby

import (
	"net/url"
	"qbit-cli/internal/api"
)

func Items(params url.Values) (*api.EmbyItems, error) {
	embyClient, err := api.GetEmbyClient()
	if err != nil {
		return nil, err
	}

	resp, err := embyClient.Get("/emby/Items", params)
	if err != nil {
		return nil, err
	}
	var result *api.EmbyItems
	if err := api.ParseJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func Item(item string) (*api.EmbyItem, error) {
	embyClient, err := api.GetEmbyClient()
	if err != nil {
		return nil, err
	}

	resp, err := embyClient.Get("/emby/Users/"+embyClient.EmbyUser()+"/Items/"+item, url.Values{})
	if err != nil {
		return nil, err
	}
	var result api.EmbyItem
	if err := api.ParseJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
