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

func RefreshItem(item string, params url.Values) error {
	if params == nil {
		params = url.Values{}
	}
	resp, err := api.GetEmbyClient().Post("/emby/Items/"+item+"/Refresh", params)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New(resp.Status)
	}
}

func RefreshItemByItemId(itemID string) error {
	params := url.Values{
		"Recursive":           {"true"},
		"MetadataRefreshMode": {"FullRefresh"},
		"ImageRefreshMode":    {"FullRefresh"},
		"ReplaceAllMetadata":  {"true"},
		"ReplaceAllImages":    {"true"},
	}
	return RefreshItem(itemID, params)
}

func ResetItemMetadata(item string) error {
	params := url.Values{
		"ItemIds": []string{item},
	}
	resp, err := api.GetEmbyClient().Post("/emby/items/metadata/reset", params)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New(resp.Status)
	}
}
