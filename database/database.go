package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type client interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type DBClient struct {
	url          string
	postfix      string
	client       client
	secret       string
	export       bool
	response     *http.Response
	responseBody []byte
}

// Retrieve a new Firebase Client
func NewFBClient(url string) *DBClient {
	return &DBClient{
		url:     url,
		postfix: ".json",
		export:  false,
		client:  &http.Client{},
	}
}

// Set url for client request
func (c *DBClient) Url(url string) *DBClient {
	c.url = url
	return c
}

// Uses the Firebase secret or Auth Token to authenticate.
func (c *DBClient) Auth(token string) *DBClient {
	c.secret = token

	return c
}

// Set to true if you want priority data to be returned.
func (c *DBClient) Export(toggle bool) *DBClient {
	c.export = toggle

	return c
}

// Execute a new HTTP Request.
func (c *DBClient) executeRequest(method string, body []byte) ([]byte, error) {
	url, err := url.Parse(c.url + c.postfix)
	if err != nil {
		return nil, err
	}
	if c.secret != "" {
		url.Query().Set("auth", c.secret)
	}
	if c.export != "" {
		url.Query().Set("format", "export")
	}

	// Prepare HTTP Request
	req, err := http.NewRequest(method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Make actual HTTP request.
	if c.response, err = c.client.Do(req); err != nil {
		return nil, err
	}

	defer c.response.Body.Close()

	// Check status code for errors.
	status := c.response.Status
	if strings.HasPrefix(status, "2") == false {
		return nil, errors.New(status)
	}

	// Read body.
	if c.responseBody, err = ioutil.ReadAll(c.response.Body); err != nil {
		return nil, err
	}

	return c.responseBody, nil
}

// Retrieve the current value for this Reference.
func (c *DBClient) Value(v interface{}) error {

	// GET the data from Firebase.
	resp, err := c.executeRequest("GET", nil)
	if err != nil {
		return err
	}

	// JSON decode the data into given interface.
	err = json.Unmarshal(resp, v)
	if err != nil {
		return err
	}

	return nil
}

// Set the value for this Reference (overwrites existing value).
func (c *DBClient) Write(v interface{}) error {

	// JSON encode the data.
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// PUT the data to Firebase.
	_, err = c.executeRequest("PUT", jsonData)
	if err != nil {
		return err
	}

	return nil
}

// Pushes a new object to this Reference (effectively creates a list).
func (c *DBClient) Push(v interface{}) error {

	// JSON encode the data.
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// POST the data to Firebase.
	_, err = c.executeRequest("POST", jsonData)
	if err != nil {
		return err
	}

	return nil
}

// Update node with give data
func (c *DBClient) Update(v interface{}) error {

	// JSON encode the data.
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// PATCH the data on Firebase.
	_, err = c.executeRequest("PATCH", jsonData)
	if err != nil {
		return err
	}

	return nil
}

// Delete any values for this node
func (c *DBClient) Delete() error {
	_, err := c.executeRequest("DELETE", nil)
	if err != nil {
		return err
	}

	return nil
}
