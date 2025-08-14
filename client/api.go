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

type Endpoint struct {
	ID string `json:"@id"`
}

type EndpointRoot struct {
	Members []*Endpoint `json:"members"`
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

type Client struct {
	target      string
	username    string
	password    string
	accessToken string

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

func New(target string) (*Client, error) {
	section, err := config.GetSection("connection:" + target)
	if err != nil {
		return nil, fmt.Errorf("Connection not found in config: %s", target)
	}

	c := &Client{}
	k, err := section.GetKey("host")
	if err != nil {
		return nil, fmt.Errorf("Connection missing host key: %s", target)
	}
	c.target = k.MustString("")

	k, err = section.GetKey("username")
	if err != nil {
		return nil, fmt.Errorf("Connection missing username key: %s", target)
	}
	c.username = k.MustString("")

	k, err = section.GetKey("password")
	if err != nil {
		return nil, fmt.Errorf("Connection missing password key: %s", target)
	}
	c.password = k.MustString("")

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c.httpClient = &http.Client{Transport: tr}

	return c, nil
}

func (c *Client) Authenticate() error {
	payload := &LoginPayload{
		Username:  c.username,
		Password:  c.password,
		GrantType: "password",
		Scope:     "GUIAccess",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s%s%s", c.target, BasePath, AuthEndpoint), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
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

	c.accessToken = oauthToken.AccessToken

	return nil
}

func (c *Client) FetchEndpoint(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s%s", c.target, endpoint), nil)
	if err != nil {
		return nil, err
	}

	if c.accessToken == "" {
		err = c.Authenticate()
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 401 {
		res.Body.Close()

		// Try to re-auth
		err = c.Authenticate()
		if err != nil {
			return nil, err
		}

		res, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
	} else if res.StatusCode != 200 {
		return nil, fmt.Errorf("Eaton restconf api call failed with HTTP status code: %d endpoint: %s", res.StatusCode, endpoint)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return rawJson, nil
}
