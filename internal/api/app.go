package api

import "net/url"

func QbitAppBuildInfo() (*QbitServerInfo, error) {
	resp, err := GetQbitClient().Get("/api/v2/app/buildInfo", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
	v, err := ParseString(resp)
	if err != nil {
		return "", err
	}
	return v, nil
}
