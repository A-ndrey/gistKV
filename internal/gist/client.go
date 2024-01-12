package gist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.ocm/A-ndrey/gistKV/internal/node"
	"net/http"
)

const (
	gistsURL        = "https://api.github.com/gists"
	accept          = "application/vnd.github.v3+json"
	gistFileName    = "gistKV_state.json"
	warnDescription = "Manual editing of this file is NOT RECOMMENDED"
)

var (
	NotFound = errors.New("gist not found")
)

type file struct {
	RawURL    string `json:"raw_url,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	Content   string `json:"content,omitempty"`
}

type Client interface {
	Read() (*node.Node, error)
	Write(root *node.Node) error
}

type client struct {
	token string
	url   string
}

func NewClient(token string) (Client, error) {
	url, err := find(token)
	if errors.Is(err, NotFound) {
		url, err = create(token)
	}

	if err != nil {
		return nil, fmt.Errorf("can't create client, %w", err)
	}

	return &client{token: token, url: url}, nil
}

func (c client) Read() (*node.Node, error) {
	resp, err := http.Get(c.url)
	if err != nil {
		return nil, fmt.Errorf("can't get gist from url=%s, %w", c.url, err)
	}

	var gistInfo struct {
		Files map[string]file `json:"files"`
	}

	err = json.NewDecoder(resp.Body).Decode(&gistInfo)
	if err != nil {
		return nil, fmt.Errorf("can't decode gist info from url=%s, %w", c.url, err)
	}

	root := &node.Node{}

	if file, ok := gistInfo.Files[gistFileName]; !ok {
		return nil, fmt.Errorf("has no gist file=%s", gistFileName)
	} else if file.Truncated {
		err = readRaw(file.RawURL, root)
	} else {
		err = json.Unmarshal([]byte(file.Content), root)
	}

	if err != nil {
		return nil, fmt.Errorf("can't read gist, %w", err)
	}

	if !root.IsRoot() {
		return nil, fmt.Errorf("gist haven't root node")
	}

	return root, nil
}

func (c client) Write(root *node.Node) error {
	if !root.IsRoot() {
		return fmt.Errorf("can't write not root node")
	}

	content, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}

	var body struct {
		Files map[string]file `json:"files"`
	}

	body.Files = map[string]file{gistFileName: {Content: string(content)}}

	bodyBytes, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	addHeaders(req, c.token)

	client := http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func readRaw(url string, data interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("can't get gist from url=%s, %w", url, err)
	}

	err = json.NewDecoder(resp.Body).Decode(data)
	if err != nil {
		return fmt.Errorf("can't decode gist from url=%s, %w", url, err)
	}

	return nil
}

func find(token string) (string, error) {
	req, err := http.NewRequest("GET", gistsURL, nil)
	if err != nil {
		return "", err
	}

	addHeaders(req, token)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var gists []struct {
		URL   string                 `json:"url"`
		Files map[string]interface{} `json:"files"`
	}

	err = json.NewDecoder(resp.Body).Decode(&gists)
	if err != nil {
		return "", err
	}

	for _, gist := range gists {
		if _, ok := gist.Files[gistFileName]; ok {
			return gist.URL, nil
		}
	}

	return "", NotFound
}

func create(token string) (string, error) {
	content, err := json.MarshalIndent(node.NewRoot(), "", "  ")
	if err != nil {
		return "", err
	}

	var body struct {
		Files       map[string]file `json:"files"`
		Description string          `json:"description,omitempty"`
	}

	body.Files = map[string]file{gistFileName: {Content: string(content)}}
	body.Description = warnDescription

	bodyBytes, err := json.Marshal(&body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", gistsURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	addHeaders(req, token)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var respInfo struct {
		URL string `json:"url"`
	}

	err = json.NewDecoder(resp.Body).Decode(&respInfo)
	if err != nil {
		return "", err
	}

	return respInfo.URL, nil
}

func addHeaders(req *http.Request, token string) {
	req.Header.Add("Accept", accept)
	req.Header.Add("Authorization", "token "+token)
}
