package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const tokenRedirectURI = "urn:ietf:wg:oauth:2.0:oob"
const tokenEndpoint = "https://accounts.google.com/o/oauth2/token"

func getAccessToken(clientID, clientSecret, refreshToken string) (string, error) {
	form := url.Values{}
	form.Add("grant_type", "refresh_token")
	form.Add("client_id", clientID)
	form.Add("client_secret", clientSecret)
	form.Add("refresh_token", refreshToken)
	form.Add("redirect_uri", tokenRedirectURI)

	res, err := http.Post(
		tokenEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()))

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error refreshing access token: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got error response: %d %q", res.StatusCode, body)
	}

	var t struct {
		AccessToken string `json:"access_token"`
	}
	err = json.Unmarshal(body, &t)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshall token response: %v", err)
	}

	return t.AccessToken, nil
}
