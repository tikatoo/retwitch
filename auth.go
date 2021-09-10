package retwitch

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
)

type twitchauth struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	AccessExpiry time.Time
}

func getDefaultAuth() (a *twitchauth, err error) {
	if defaultTwitchClientID == "" || defaultTwitchClientSecret == "" {
		return nil, errNoDefaultAuth
	}

	a = &twitchauth{
		ClientID:     defaultTwitchClientID,
		ClientSecret: defaultTwitchClientSecret,
	}
	err = a.update()
	return
}

func (a *twitchauth) update() (err error) {
	if a.AccessToken != "" && time.Until(a.AccessExpiry) > time.Duration(10)*time.Second {
		return nil
	}

	q := url.Values{
		"client_id":     {a.ClientID},
		"client_secret": {a.ClientSecret},
		"grant_type":    {"client_credentials"},
	}

	url := "https://id.twitch.tv/oauth2/token?" + q.Encode()
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	type CredsResponse struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	var cr CredsResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&cr)
	if err != nil {
		return
	}

	if cr.TokenType != "bearer" {
		err = errTwitchAuthTokenType
		return
	}

	a.AccessToken = "Bearer " + cr.AccessToken
	a.AccessExpiry = time.Now().Add(time.Duration(cr.ExpiresIn) * time.Second)
	return
}

var errNoDefaultAuth = errors.New("can't find default client settings")
var errTwitchAuthTokenType = errors.New("don't understand twitch auth token type")
var (
	defaultTwitchClientID     = os.Getenv("TWITCH_CLIENT_ID")
	defaultTwitchClientSecret = os.Getenv("TWITCH_CLIENT_SECRET")
)
