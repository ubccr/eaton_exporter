package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gopkg.in/ini.v1"
)

var (
	config *ini.File
)

type EatonEndpoint interface {
	GetPath() string
}

type LoginPayload struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	GrantType string `json:"grant_type"`
	Scope     string `json:"scope"`
}

type OAuthToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

type Handler struct {
	endpoints []EatonEndpoint
	target    string
	username  string
	password  string

	httpClient *http.Client
}

func LoadConfig(path string) error {
	var err error
	config, err = ini.Load(path)
	if err != nil {
		return err
	}

	return nil
}

func GetHandle(target string) (*Handler, error) {
	section, err := config.GetSection("connection:" + target)
	if err != nil {
		return nil, fmt.Errorf("Connection not found in config: %s", target)
	}

	h := &Handler{}
	k, err := section.GetKey("host")
	if err != nil {
		return nil, fmt.Errorf("Connection missing host key: %s", target)
	}
	h.target = k.MustString("")

	k, err = section.GetKey("username")
	if err != nil {
		return nil, fmt.Errorf("Connection missing username key: %s", target)
	}
	h.username = k.MustString("")

	k, err = section.GetKey("password")
	if err != nil {
		return nil, fmt.Errorf("Connection missing password key: %s", target)
	}
	h.password = k.MustString("")

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	h.httpClient = &http.Client{Transport: tr}

	return h, nil
}

func (h *Handler) AddEndpoint(endpoint EatonEndpoint) {
	h.endpoints = append(h.endpoints, endpoint)
}

func (h *Handler) authenticate() error {
	payload := &LoginPayload{
		Username:  h.username,
		Password:  h.password,
		GrantType: "password",
		Scope:     "GUIAccess",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s%s%s", h.target, BasePath, AuthEndpoint), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	res, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Failed to authenticate to Eaton rest api with HTTP status code: %d", res.StatusCode)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var oauthToken OAuthToken

	err = json.Unmarshal(rawJson, &oauthToken)
	if err != nil {
		return err
	}

	if oauthToken.AccessToken == "" {
		return fmt.Errorf("Failed to obtain access_token")
	}

	AccessToken = oauthToken.AccessToken

	return nil
}

func (h *Handler) fetchEndpoint(endpoint EatonEndpoint) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s%s%s", h.target, BasePath, endpoint.GetPath()), nil)
	if err != nil {
		return err
	}

	if AccessToken == "" {
		err = h.authenticate()
		if err != nil {
			return err
		}
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AccessToken))
	res, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		res.Body.Close()

		// Try to re-auth
		err = h.authenticate()
		if err != nil {
			return err
		}

		res, err = h.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
	} else if res.StatusCode != 200 {
		return fmt.Errorf("Eaton restconf api call failed with HTTP status code: %d", res.StatusCode)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawJson, &endpoint)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) Fetch() error {
	for _, e := range h.endpoints {
		err := h.fetchEndpoint(e)
		if err != nil {
			return err
		}
	}

	return nil
}
