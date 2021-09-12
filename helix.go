package retwitch

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type HelixAPI struct {
	http.Client

	useridCache map[string]string
}

type HelixCheermote struct {
	CheerID     string
	CheerPrefix string
	CheerValue  int
	CheerColor  string
	NextTierID  string
	ImageURL    string
}

type HelixChatBadge struct {
	BadgeID         string
	BadgeVersion    string
	ImageURL        string
	ImageURLForSize map[string]string
}

func (h *HelixAPI) GetUserID(login string) (id string, err error) {
	if cachedid, iscached := h.useridCache[login]; iscached {
		return cachedid, nil
	}

	q := url.Values{"login": {login}}
	url := "https://api.twitch.tv/helix/users?" + q.Encode()

	resp, err := h.getEnsureOK(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	type ResponseUser struct {
		ID string `json:"id"`
	}
	type ResponseContainer struct {
		Data []ResponseUser `json:"data"`
	}

	var body ResponseContainer
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&body)
	if err != nil {
		return
	}

	id = body.Data[0].ID

	h.useridCache[login] = id
	return
}

func (h *HelixAPI) GetCheermotes(bcid string) (cheermotePrefixes []string, cheermoteInfo map[string]HelixCheermote, err error) {
	query := ""
	if bcid != "" {
		query = "?broadcaster_id=" + bcid
	}

	resp, err := h.getEnsureOK("https://api.twitch.tv/helix/bits/cheermotes" + query)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	type ResponseTier struct {
		MinBits int                                     `json:"min_bits"`
		Color   string                                  `json:"color"`
		Images  map[string]map[string]map[string]string `json:"images"`
	}
	type ResponseCheermote struct {
		Prefix string         `json:"prefix"`
		Tiers  []ResponseTier `json:"tiers"`
	}
	type ResponseContainer struct {
		Data []ResponseCheermote `json:"data"`
	}

	var body ResponseContainer
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&body)
	if err != nil {
		return
	}

	cheermotePrefixes = make([]string, 0, len(body.Data))
	cheermoteInfo = map[string]HelixCheermote{}
	for _, mote := range body.Data {
		var lastTier HelixCheermote
		cheermotePrefixes = append(cheermotePrefixes, mote.Prefix)
		for _, tier := range mote.Tiers {
			currentTier := HelixCheermote{
				CheerID:     mote.Prefix + strconv.Itoa(tier.MinBits),
				CheerPrefix: mote.Prefix,
				CheerValue:  tier.MinBits,
				CheerColor:  tier.Color,
				ImageURL:    tier.Images["dark"]["animated"]["1"],
			}

			cheermoteInfo[currentTier.CheerID] = currentTier

			if lastTier.CheerID != "" {
				lastTier.NextTierID = currentTier.CheerID
				cheermoteInfo[lastTier.CheerID] = lastTier
			}

			lastTier = currentTier
		}
	}

	return
}

func (h *HelixAPI) GetCheermotesFor(username string) (cheermotePrefixes []string, cheermotesInfo map[string]HelixCheermote, err error) {
	bcid, err := h.GetUserID(username)
	if err != nil {
		return
	}

	return h.GetCheermotes(bcid)
}

func (h *HelixAPI) GetGlobalChatBadges() (badges map[string]HelixChatBadge, err error) {
	resp, err := h.getEnsureOK("https://api.twitch.tv/helix/chat/badges/global")
	if err != nil {
		return
	}

	defer resp.Body.Close()
	return findHelixChatBadges(resp.Body)
}

func (h *HelixAPI) GetChannelChatBadges(bcid string) (badges map[string]HelixChatBadge, err error) {
	resp, err := h.getEnsureOK("https://api.twitch.tv/helix/chat/badges?broadcaster_id=" + bcid)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	return findHelixChatBadges(resp.Body)
}

func (h *HelixAPI) GetChannelChatBadgesFor(username string) (badges map[string]HelixChatBadge, err error) {
	bcid, err := h.GetUserID(username)
	if err != nil {
		return
	}

	return h.GetChannelChatBadges(bcid)
}

func findHelixChatBadges(rbody io.ReadCloser) (badges map[string]HelixChatBadge, err error) {
	type ResponseVersion map[string]string
	type ResponseBadge struct {
		SetID    string            `json:"set_id"`
		Versions []ResponseVersion `json:"versions"`
	}
	type ResponseContainer struct {
		Data []ResponseBadge `json:"data"`
	}

	var body ResponseContainer
	dec := json.NewDecoder(rbody)
	err = dec.Decode(&body)
	if err != nil {
		return
	}

	badges = map[string]HelixChatBadge{}
	for _, badge := range body.Data {
		for _, version := range badge.Versions {
			helixBadge := HelixChatBadge{
				BadgeID:         badge.SetID,
				BadgeVersion:    version["id"],
				ImageURL:        version["image_url_1x"],
				ImageURLForSize: map[string]string{},
			}

			for key, value := range version {
				size := strings.TrimPrefix(key, "image_url_")
				if size == key {
					continue
				}

				helixBadge.ImageURLForSize[size] = value
			}

			badges[badge.SetID+"/"+version["id"]] = helixBadge
		}
	}

	return
}

func (h *HelixAPI) getEnsureOK(url string) (resp *http.Response, err error) {
	resp, err = h.Get(url)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = httpStatusError{resp}
	}

	return
}

type helixrt struct {
	auth *twitchauth
	rt   http.RoundTripper
}

func getHelixAPI(auth *twitchauth) *HelixAPI {
	return &HelixAPI{
		Client: http.Client{
			Transport: &helixrt{auth: auth},
		},
		useridCache: map[string]string{},
	}
}

func (hrt *helixrt) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := hrt.auth.update(); err != nil {
		return nil, err
	}

	req = req.Clone(req.Context())
	req.Header.Add("Authorization", hrt.auth.AccessToken)
	req.Header.Add("Client-Id", hrt.auth.ClientID)

	if hrt.rt == nil {
		hrt.rt = http.DefaultTransport
	}

	// TODO: More gracefully handle token expiration

	return hrt.rt.RoundTrip(req)
}
