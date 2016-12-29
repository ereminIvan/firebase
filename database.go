package firebase

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Method string

const (
	POST   Method = "POST"
	GET    Method = "GET"
	PATCH  Method = "PATCH"
	DELETE Method = "DELETET"
	PUT    Method = "PUT"
)

const (
	debugHeader = "X-Firebase-Auth-Debug"

	//Available query params
	ParamAccessToken = "auth"
	ParamFormat      = "format"
	//ParamShallow - This is an advanced feature, designed to help you work with large datasets without needing to
	//download everything. Set this to true to limit the depth of the data returned at a location. If the data at the
	//location is a JSON primitive (string, number or boolean), its value will simply be returned. If the data snapshot
	//at the location is a JSON object, the values for each key will be truncated to true.
	ParamShallow = "shallow"
	//Formats the data returned in the response from the server.
	//pretty : View the data in a human-readable format.
	//silent : Used to suppress the output from the server when writing data. The resulting response will be empty and
	//indicated by a 204 No Content HTTP status code.
	ParamPrint    = "print"
	ParamDownload = "download"
	paramOrderBy  = "orderBy" //todo implement
)

var availableParams = map[Method][]string{
	POST:   {ParamAccessToken, ParamPrint},
	GET:    {ParamAccessToken, ParamShallow, ParamPrint},
	PATCH:  {ParamAccessToken, ParamPrint},
	DELETE: {ParamAccessToken, ParamPrint},
	PUT:    {ParamAccessToken, ParamPrint},
}

// IRequestClient client interface
type IRequestClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

type dbClient struct {
	baseUrl     string
	client      IRequestClient
	accessToken string
	export      bool //If set to export, the server will encode priorities in the response.
	shallow     bool //Limit the depth of the response
}

// Retrieve a new Firebase Client
// baseUrl, accessToken - required
func NewDBClient(baseUrl, accessToken string, export bool, client IRequestClient) *dbClient {
	if client == nil {
		client = &http.Client{}
	}
	return &dbClient{
		baseUrl:     baseUrl,
		accessToken: accessToken,
		export:      export,
		client:      client,
	}
}

// Execute a new HTTP Request.
func (c *dbClient) executeRequest(method Method, path string, body []byte) ([]byte, error) {

	req, err := c.buildRequest(path, method, body)
	if err != nil {
		return nil, err
	}
	// Make actual HTTP request.
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if h := res.Header.Get(debugHeader); h != "" {
		log.Printf("Debug: %s", h)
	}
	// Check status code for errors.
	status := res.Status
	if strings.HasPrefix(status, "2") == false {
		return nil, errors.New(status)
	}

	// Read body.
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}

func (c *dbClient) buildRequest(path string, method Method, body []byte) (*http.Request, error) {
	//Build query params
	q := url.Values{}
	if c.accessToken != "" {
		q.Add(ParamAccessToken, c.accessToken)
	}
	if c.export {
		q.Add(ParamFormat, "export")
	}
	if c.shallow {
		q.Add(ParamShallow, "true")
	}
	// Prepare HTTP Request
	u := c.baseUrl + path + ".json" + "?" + q.Encode()

	return http.NewRequest(string(method), u, bytes.NewReader(body))
}

// Get the current value for this Reference.
// Data from our Firebase database can be read by issuing an HTTP GET request to and endpoint
// A successful request will be indicated by a 200 OK HTTP status code.
// The response will contain the data being retrieved
func (c *dbClient) Get(path string, v interface{}) error {
	resp, err := c.executeRequest(GET, path, nil)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(resp, v); err != nil {
		return err
	}
	return nil
}

// Write the value for this Reference (overwrites existing value).
// A successful request will be indicated by a 200 OK HTTP status code.
func (c *dbClient) Write(path string, v interface{}) error {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = c.executeRequest(PUT, path, jsonData)
	if err != nil {
		return err
	}

	return nil
}

// Create a new object to this Reference (effectively creates a list).
// A successful request will be indicated by a 200 OK HTTP status code.
func (c *dbClient) Create(path string, v interface{}) error {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err = c.executeRequest(POST, path, jsonData); err != nil {
		return err
	}

	return nil
}

// Update node with give data
func (c *dbClient) Update(path string, v interface{}) error {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err = c.executeRequest(PATCH, path, jsonData); err != nil {
		return err
	}

	return nil
}

// Delete any values for this node
func (c *dbClient) Delete(path string) error {
	_, err := c.executeRequest(DELETE, path, nil)
	return err
}
