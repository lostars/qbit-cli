package api

import (
	"net/url"
	"qbit-cli/pkg/utils"
)

func QbitAppBuildInfo() (*QbitServerInfo, error) {
	resp, err := GetQbitClient().Get("/api/v2/app/buildInfo", url.Values{})
	if err != nil {
		return nil, err
	}
	defer utils.SafeClose(resp.Body)
	var info QbitServerInfo
	err = ParseJSON(resp, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func QbitApiVersion() (string, error) {
	resp, err := GetQbitClient().Get("/api/v2/app/webapiVersion", url.Values{})
	if err != nil {
		return "", err
	}
	defer utils.SafeClose(resp.Body)
	v, err := ParseString(resp)
	if err != nil {
		return "", err
	}
	return v, nil
}

func QbitAppVersion() (string, error) {
	resp, err := GetQbitClient().Get("/api/v2/app/version", url.Values{})
	if err != nil {
		return "", err
	}
	defer utils.SafeClose(resp.Body)
	v, err := ParseString(resp)
	if err != nil {
		return "", err
	}
	return v, nil
}

func QbitAppPreference() (string, error) {
	resp, err := GetQbitClient().Get("/api/v2/app/preferences", url.Values{})
	if err != nil {
		return "", err
	}
	defer utils.SafeClose(resp.Body)
	json, err := ParseRawJSON(resp)
	if err != nil {
		return "", err
	}
	return json, nil
}

func QbitSetAppPreference(json string) error {
	params := url.Values{}
	params.Set("json", json)
	resp, err := GetQbitClient().Post("/api/v2/app/setPreferences", params)
	if err != nil {
		return err
	}
	defer utils.SafeClose(resp.Body)
	return nil
}
