package gksync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

var ErrNetworkUnavailable = errors.New("network unavailable")

// PostAddDataTableUserIDJSONBody defines parameters for PostAddDataTableUserID.
type PostAddDataTableUserIDJSONBody map[string]string

// PutUpdateDataTableUserIDIdJSONBody defines parameters for PutUpdateDataTableUserIDId.
type PutUpdateDataTableUserIDIdJSONBody map[string]string

// PostAddDataTableUserIDJSONRequestBody defines body for PostAddDataTableUserID for application/json ContentType.
type PostAddDataTableUserIDJSONRequestBody PostAddDataTableUserIDJSONBody

// PutUpdateDataTableUserIDIdJSONRequestBody defines body for PutUpdateDataTableUserIDId for application/json ContentType.
type PutUpdateDataTableUserIDIdJSONRequestBody PutUpdateDataTableUserIDIdJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	syncWithServer bool

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, syncWithServer bool, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// PostAddDataTableUserIDWithBody request with any body
	PostAddDataTableUserIDWithBody(ctx context.Context, table string, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostAddDataTableUserID(ctx context.Context, table string, userID int, body PostAddDataTableUserIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteClearDataTableUserID request
	DeleteClearDataTableUserID(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteDeleteDataTableUserIDId request
	DeleteDeleteDataTableUserIDId(ctx context.Context, table string, userID int, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetGetAllDataTableUserID request
	GetGetAllDataTableUserID(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetGetDataTableUserID request
	GetGetDataTableUserID(ctx context.Context, userID int, table string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetGetPasswordUsername request
	GetGetPasswordUsername(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetGetUserIDUsername request
	GetGetUserIDUsername(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostSendFileUserIDWithBody request with any body
	PostSendFileUserIDWithBody(ctx context.Context, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PutUpdateDataTableUserIDIdWithBody request with any body
	PutUpdateDataTableUserIDIdWithBody(ctx context.Context, table string, userID int, id int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PutUpdateDataTableUserIDId(ctx context.Context, table string, userID int, id int, body PutUpdateDataTableUserIDIdJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) PostAddDataTableUserIDWithBody(ctx context.Context, table string, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostAddDataTableUserIDRequestWithBody(c.Server, table, userID, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostAddDataTableUserID(ctx context.Context, table string, userID int, body PostAddDataTableUserIDJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostAddDataTableUserIDRequest(c.Server, table, userID, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteClearDataTableUserID(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteClearDataTableUserIDRequest(c.Server, table, userID)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteDeleteDataTableUserIDId(ctx context.Context, table string, userID int, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteDeleteDataTableUserIDIdRequest(c.Server, table, userID, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetGetAllDataTableUserID(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetGetAllDataTableUserIDRequest(c.Server, table, userID)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetGetDataTableUserID(ctx context.Context, userID int, table string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	if !c.syncWithServer {
		return nil, nil
	}
	req, err := NewGetGetDataTableUserIDRequest(c.Server, table, userID)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, ErrNetworkUnavailable
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseData map[string]string
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) GetGetPasswordUsername(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetGetPasswordUsernameRequest(c.Server, username)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetGetUserIDUsername(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetGetUserIDUsernameRequest(c.Server, username)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostSendFileUserIDWithBody(ctx context.Context, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostSendFileUserIDRequestWithBody(c.Server, userID, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PutUpdateDataTableUserIDIdWithBody(ctx context.Context, table string, userID int, id int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPutUpdateDataTableUserIDIdRequestWithBody(c.Server, table, userID, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PutUpdateDataTableUserIDId(ctx context.Context, table string, userID int, id int, body PutUpdateDataTableUserIDIdJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPutUpdateDataTableUserIDIdRequest(c.Server, table, userID, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewPostAddDataTableUserIDRequest calls the generic PostAddDataTableUserID builder with application/json body
func NewPostAddDataTableUserIDRequest(server string, table string, userID int, body PostAddDataTableUserIDJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostAddDataTableUserIDRequestWithBody(server, table, userID, "application/json", bodyReader)
}

// NewPostAddDataTableUserIDRequestWithBody generates requests for PostAddDataTableUserID with any type of body
func NewPostAddDataTableUserIDRequestWithBody(server string, table string, userID int, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "table", runtime.ParamLocationPath, table)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "userID", runtime.ParamLocationPath, userID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/addData/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteClearDataTableUserIDRequest generates requests for DeleteClearDataTableUserID
func NewDeleteClearDataTableUserIDRequest(server string, table string, userID int) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "table", runtime.ParamLocationPath, table)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "userID", runtime.ParamLocationPath, userID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/clearData/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteDeleteDataTableUserIDIdRequest generates requests for DeleteDeleteDataTableUserIDId
func NewDeleteDeleteDataTableUserIDIdRequest(server string, table string, userID int, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "table", runtime.ParamLocationPath, table)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "userID", runtime.ParamLocationPath, userID)
	if err != nil {
		return nil, err
	}

	var pathParam2 string

	pathParam2, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/deleteData/%s/%s/%s", pathParam0, pathParam1, pathParam2)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetGetAllDataTableUserIDRequest generates requests for GetGetAllDataTableUserID
func NewGetGetAllDataTableUserIDRequest(server string, table string, userID int) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "table", runtime.ParamLocationPath, table)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "userID", runtime.ParamLocationPath, userID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/getAllData/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetGetDataTableUserIDRequest generates requests for GetGetDataTableUserID
func NewGetGetDataTableUserIDRequest(server string, table string, userID int) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "table", runtime.ParamLocationPath, table)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "userID", runtime.ParamLocationPath, userID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/getData/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetGetPasswordUsernameRequest generates requests for GetGetPasswordUsername
func NewGetGetPasswordUsernameRequest(server string, username string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "username", runtime.ParamLocationPath, username)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/getPassword/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetGetUserIDUsernameRequest generates requests for GetGetUserIDUsername
func NewGetGetUserIDUsernameRequest(server string, username string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "username", runtime.ParamLocationPath, username)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/getUserID/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostSendFileUserIDRequestWithBody generates requests for PostSendFileUserID with any type of body
func NewPostSendFileUserIDRequestWithBody(server string, userID int, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "userID", runtime.ParamLocationPath, userID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sendFile/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewPutUpdateDataTableUserIDIdRequest calls the generic PutUpdateDataTableUserIDId builder with application/json body
func NewPutUpdateDataTableUserIDIdRequest(server string, table string, userID int, id int, body PutUpdateDataTableUserIDIdJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPutUpdateDataTableUserIDIdRequestWithBody(server, table, userID, id, "application/json", bodyReader)
}

// NewPutUpdateDataTableUserIDIdRequestWithBody generates requests for PutUpdateDataTableUserIDId with any type of body
func NewPutUpdateDataTableUserIDIdRequestWithBody(server string, table string, userID int, id int, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "table", runtime.ParamLocationPath, table)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "userID", runtime.ParamLocationPath, userID)
	if err != nil {
		return nil, err
	}

	var pathParam2 string

	pathParam2, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/updateData/%s/%s/%s", pathParam0, pathParam1, pathParam2)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, syncWithServer bool, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, syncWithServer, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// PostAddDataTableUserIDWithBodyWithResponse request with any body
	PostAddDataTableUserIDWithBodyWithResponse(ctx context.Context, table string, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostAddDataTableUserIDResponse, error)

	PostAddDataTableUserIDWithResponse(ctx context.Context, table string, userID int, body PostAddDataTableUserIDJSONRequestBody, reqEditors ...RequestEditorFn) (*PostAddDataTableUserIDResponse, error)

	// DeleteClearDataTableUserIDWithResponse request
	DeleteClearDataTableUserIDWithResponse(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*DeleteClearDataTableUserIDResponse, error)

	// DeleteDeleteDataTableUserIDIdWithResponse request
	DeleteDeleteDataTableUserIDIdWithResponse(ctx context.Context, table string, userID int, id string, reqEditors ...RequestEditorFn) (*DeleteDeleteDataTableUserIDIdResponse, error)

	// GetGetAllDataTableUserIDWithResponse request
	GetGetAllDataTableUserIDWithResponse(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*GetGetAllDataTableUserIDResponse, error)

	// GetGetDataTableUserIDWithResponse request
	GetGetDataTableUserIDWithResponse(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*GetGetDataTableUserIDResponse, error)

	// GetGetPasswordUsernameWithResponse request
	GetGetPasswordUsernameWithResponse(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*GetGetPasswordUsernameResponse, error)

	// GetGetUserIDUsernameWithResponse request
	GetGetUserIDUsernameWithResponse(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*GetGetUserIDUsernameResponse, error)

	// PostSendFileUserIDWithBodyWithResponse request with any body
	PostSendFileUserIDWithBodyWithResponse(ctx context.Context, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostSendFileUserIDResponse, error)

	// PutUpdateDataTableUserIDIdWithBodyWithResponse request with any body
	PutUpdateDataTableUserIDIdWithBodyWithResponse(ctx context.Context, table string, userID int, id int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PutUpdateDataTableUserIDIdResponse, error)

	PutUpdateDataTableUserIDIdWithResponse(ctx context.Context, table string, userID int, id int, body PutUpdateDataTableUserIDIdJSONRequestBody, reqEditors ...RequestEditorFn) (*PutUpdateDataTableUserIDIdResponse, error)
}

type PostAddDataTableUserIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r PostAddDataTableUserIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostAddDataTableUserIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteClearDataTableUserIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r DeleteClearDataTableUserIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteClearDataTableUserIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteDeleteDataTableUserIDIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r DeleteDeleteDataTableUserIDIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteDeleteDataTableUserIDIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetGetAllDataTableUserIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]map[string]string
}

// Status returns HTTPResponse.Status
func (r GetGetAllDataTableUserIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetGetAllDataTableUserIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetGetDataTableUserIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *map[string]string
}

// Status returns HTTPResponse.Status
func (r GetGetDataTableUserIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetGetDataTableUserIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetGetPasswordUsernameResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *string
}

// Status returns HTTPResponse.Status
func (r GetGetPasswordUsernameResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetGetPasswordUsernameResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetGetUserIDUsernameResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *int
}

// Status returns HTTPResponse.Status
func (r GetGetUserIDUsernameResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetGetUserIDUsernameResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostSendFileUserIDResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r PostSendFileUserIDResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostSendFileUserIDResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PutUpdateDataTableUserIDIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r PutUpdateDataTableUserIDIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PutUpdateDataTableUserIDIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// PostAddDataTableUserIDWithBodyWithResponse request with arbitrary body returning *PostAddDataTableUserIDResponse
func (c *ClientWithResponses) PostAddDataTableUserIDWithBodyWithResponse(ctx context.Context, table string, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostAddDataTableUserIDResponse, error) {
	rsp, err := c.PostAddDataTableUserIDWithBody(ctx, table, userID, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostAddDataTableUserIDResponse(rsp)
}

func (c *ClientWithResponses) PostAddDataTableUserIDWithResponse(ctx context.Context, table string, userID int, body PostAddDataTableUserIDJSONRequestBody, reqEditors ...RequestEditorFn) (*PostAddDataTableUserIDResponse, error) {
	rsp, err := c.PostAddDataTableUserID(ctx, table, userID, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostAddDataTableUserIDResponse(rsp)
}

// DeleteClearDataTableUserIDWithResponse request returning *DeleteClearDataTableUserIDResponse
func (c *ClientWithResponses) DeleteClearDataTableUserIDWithResponse(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*DeleteClearDataTableUserIDResponse, error) {
	rsp, err := c.DeleteClearDataTableUserID(ctx, table, userID, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteClearDataTableUserIDResponse(rsp)
}

// DeleteDeleteDataTableUserIDIdWithResponse request returning *DeleteDeleteDataTableUserIDIdResponse
func (c *ClientWithResponses) DeleteDeleteDataTableUserIDIdWithResponse(ctx context.Context, table string, userID int, id string, reqEditors ...RequestEditorFn) (*DeleteDeleteDataTableUserIDIdResponse, error) {
	rsp, err := c.DeleteDeleteDataTableUserIDId(ctx, table, userID, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteDeleteDataTableUserIDIdResponse(rsp)
}

// GetGetAllDataTableUserIDWithResponse request returning *GetGetAllDataTableUserIDResponse
func (c *ClientWithResponses) GetGetAllDataTableUserIDWithResponse(ctx context.Context, table string, userID int, reqEditors ...RequestEditorFn) (*GetGetAllDataTableUserIDResponse, error) {
	rsp, err := c.GetGetAllDataTableUserID(ctx, table, userID, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetGetAllDataTableUserIDResponse(rsp)
}

// GetGetDataTableUserIDWithResponse request returning *GetGetDataTableUserIDResponse
func (c *ClientWithResponses) GetGetDataTableUserIDWithResponse(ctx context.Context, userID int, table string, reqEditors ...RequestEditorFn) (*GetGetDataTableUserIDResponse, error) {
	rsp, err := c.GetGetDataTableUserID(ctx, userID, table, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetGetDataTableUserIDResponse(rsp)
}

// GetGetPasswordUsernameWithResponse request returning *GetGetPasswordUsernameResponse
func (c *ClientWithResponses) GetGetPasswordUsernameWithResponse(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*GetGetPasswordUsernameResponse, error) {
	rsp, err := c.GetGetPasswordUsername(ctx, username, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetGetPasswordUsernameResponse(rsp)
}

// GetGetUserIDUsernameWithResponse request returning *GetGetUserIDUsernameResponse
func (c *ClientWithResponses) GetGetUserIDUsernameWithResponse(ctx context.Context, username string, reqEditors ...RequestEditorFn) (*GetGetUserIDUsernameResponse, error) {
	rsp, err := c.GetGetUserIDUsername(ctx, username, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetGetUserIDUsernameResponse(rsp)
}

// PostSendFileUserIDWithBodyWithResponse request with arbitrary body returning *PostSendFileUserIDResponse
func (c *ClientWithResponses) PostSendFileUserIDWithBodyWithResponse(ctx context.Context, userID int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostSendFileUserIDResponse, error) {
	rsp, err := c.PostSendFileUserIDWithBody(ctx, userID, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostSendFileUserIDResponse(rsp)
}

// PutUpdateDataTableUserIDIdWithBodyWithResponse request with arbitrary body returning *PutUpdateDataTableUserIDIdResponse
func (c *ClientWithResponses) PutUpdateDataTableUserIDIdWithBodyWithResponse(ctx context.Context, table string, userID int, id int, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PutUpdateDataTableUserIDIdResponse, error) {
	rsp, err := c.PutUpdateDataTableUserIDIdWithBody(ctx, table, userID, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePutUpdateDataTableUserIDIdResponse(rsp)
}

func (c *ClientWithResponses) PutUpdateDataTableUserIDIdWithResponse(ctx context.Context, table string, userID int, id int, body PutUpdateDataTableUserIDIdJSONRequestBody, reqEditors ...RequestEditorFn) (*PutUpdateDataTableUserIDIdResponse, error) {
	rsp, err := c.PutUpdateDataTableUserIDId(ctx, table, userID, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePutUpdateDataTableUserIDIdResponse(rsp)
}

// ParsePostAddDataTableUserIDResponse parses an HTTP response from a PostAddDataTableUserIDWithResponse call
func ParsePostAddDataTableUserIDResponse(rsp *http.Response) (*PostAddDataTableUserIDResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostAddDataTableUserIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseDeleteClearDataTableUserIDResponse parses an HTTP response from a DeleteClearDataTableUserIDWithResponse call
func ParseDeleteClearDataTableUserIDResponse(rsp *http.Response) (*DeleteClearDataTableUserIDResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteClearDataTableUserIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseDeleteDeleteDataTableUserIDIdResponse parses an HTTP response from a DeleteDeleteDataTableUserIDIdWithResponse call
func ParseDeleteDeleteDataTableUserIDIdResponse(rsp *http.Response) (*DeleteDeleteDataTableUserIDIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteDeleteDataTableUserIDIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseGetGetAllDataTableUserIDResponse parses an HTTP response from a GetGetAllDataTableUserIDWithResponse call
func ParseGetGetAllDataTableUserIDResponse(rsp *http.Response) (*GetGetAllDataTableUserIDResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetGetAllDataTableUserIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []map[string]string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetGetDataTableUserIDResponse parses an HTTP response from a GetGetDataTableUserIDWithResponse call
func ParseGetGetDataTableUserIDResponse(rsp *http.Response) (*GetGetDataTableUserIDResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetGetDataTableUserIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest map[string]string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetGetPasswordUsernameResponse parses an HTTP response from a GetGetPasswordUsernameWithResponse call
func ParseGetGetPasswordUsernameResponse(rsp *http.Response) (*GetGetPasswordUsernameResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetGetPasswordUsernameResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest string
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetGetUserIDUsernameResponse parses an HTTP response from a GetGetUserIDUsernameWithResponse call
func ParseGetGetUserIDUsernameResponse(rsp *http.Response) (*GetGetUserIDUsernameResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetGetUserIDUsernameResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest int
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePostSendFileUserIDResponse parses an HTTP response from a PostSendFileUserIDWithResponse call
func ParsePostSendFileUserIDResponse(rsp *http.Response) (*PostSendFileUserIDResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostSendFileUserIDResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParsePutUpdateDataTableUserIDIdResponse parses an HTTP response from a PutUpdateDataTableUserIDIdWithResponse call
func ParsePutUpdateDataTableUserIDIdResponse(rsp *http.Response) (*PutUpdateDataTableUserIDIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PutUpdateDataTableUserIDIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}
