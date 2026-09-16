// Package apiclient is a typed client for the SecureEnv HTTP API v1.
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PoCInnovation/SecureEnv/internal/apiv1"
	"github.com/PoCInnovation/SecureEnv/internal/domain"
)

const defaultTimeout = 30 * time.Second

// Error is a non 2xx API response. It unwraps to the matching domain error so
// callers can use errors.Is(err, domain.ErrProjectNotFound).
type Error struct {
	Status  int
	Code    apiv1.Code
	Message string
}

func (e *Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("api: %d %s", e.Status, http.StatusText(e.Status))
	}
	return "api: " + e.Message
}

func (e *Error) Unwrap() error { return apiv1.ToError(e.Code) }

// Client calls the API with a Vault token.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
	userAgent  string
}

// Option customises a Client.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(cl *Client) { cl.httpClient = c } }

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option { return func(cl *Client) { cl.userAgent = ua } }

// New returns a client for the API at baseURL.
func New(baseURL, token string, opts ...Option) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("invalid API URL %q: expected http(s)://host[:port]", baseURL)
	}
	if token == "" {
		return nil, fmt.Errorf("%w: no token provided", domain.ErrUnauthorized)
	}

	c := &Client{
		baseURL:    parsed,
		token:      token,
		httpClient: &http.Client{Timeout: defaultTimeout},
		userAgent:  "secureenv",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// ListProjects returns every project name visible to the token.
func (c *Client) ListProjects(ctx context.Context) ([]string, error) {
	var out apiv1.ProjectList
	_, err := c.do(ctx, http.MethodGet, apiv1.BasePath+"/projects", nil, nil, &out)
	return out.Projects, err
}

// CreateProject creates an empty project.
func (c *Client) CreateProject(ctx context.Context, name string) error {
	if _, err := domain.NewProjectName(name); err != nil {
		return err
	}
	_, err := c.do(ctx, http.MethodPost, apiv1.BasePath+"/projects", nil, apiv1.ProjectNameRequest{Name: name}, nil)
	return err
}

// Project returns a project's metadata.
func (c *Client) Project(ctx context.Context, name string) (apiv1.Project, error) {
	path, err := projectPath(name)
	if err != nil {
		return apiv1.Project{}, err
	}
	var out apiv1.Project
	_, err = c.do(ctx, http.MethodGet, path, nil, nil, &out)
	return out, err
}

// RenameProject renames a project.
func (c *Client) RenameProject(ctx context.Context, from, to string) error {
	path, err := projectPath(from)
	if err != nil {
		return err
	}
	if _, err := domain.NewProjectName(to); err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodPatch, path, nil, apiv1.ProjectNameRequest{Name: to}, nil)
	return err
}

// DeleteProject permanently deletes a project.
func (c *Client) DeleteProject(ctx context.Context, name string) error {
	path, err := projectPath(name)
	if err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodDelete, path, nil, nil, nil)
	return err
}

// Variables returns the latest variables of a project.
func (c *Client) Variables(ctx context.Context, project string) (domain.Snapshot, error) {
	path, err := projectPath(project, "")
	if err != nil {
		return domain.Snapshot{}, err
	}
	var out apiv1.Variables
	if _, err := c.do(ctx, http.MethodGet, path, nil, nil, &out); err != nil {
		return domain.Snapshot{}, err
	}
	vars, err := domain.NewVariables(out.Variables)
	if err != nil {
		return domain.Snapshot{}, fmt.Errorf("api: invalid variables in response: %w", err)
	}
	return domain.Snapshot{Version: domain.Version(out.Version), Variables: vars}, nil
}

// ReplaceVariables overwrites the variables of a project. Unless expected is
// domain.AnyVersion, the call fails with domain.ErrVersionConflict when the
// project changed since that version.
func (c *Client) ReplaceVariables(ctx context.Context, project string, vars map[string]string, expected domain.Version) (domain.Version, error) {
	path, err := projectPath(project, "")
	if err != nil {
		return 0, err
	}
	header := http.Header{}
	if expected != domain.AnyVersion {
		header.Set("If-Match", strconv.Quote(strconv.Itoa(int(expected))))
	}
	return c.version(c.do(ctx, http.MethodPut, path, header, apiv1.ReplaceVariablesRequest{Variables: vars}, nil))
}

// Variable returns the value of one variable.
func (c *Client) Variable(ctx context.Context, project, key string) (string, error) {
	path, err := projectPath(project, key)
	if err != nil {
		return "", err
	}
	var out apiv1.Variable
	_, err = c.do(ctx, http.MethodGet, path, nil, nil, &out)
	return out.Value, err
}

// SetVariable creates or updates one variable.
func (c *Client) SetVariable(ctx context.Context, project, key, value string) (domain.Version, error) {
	path, err := projectPath(project, key)
	if err != nil {
		return 0, err
	}
	return c.version(c.do(ctx, http.MethodPut, path, nil, apiv1.SetVariableRequest{Value: value}, nil))
}

// DeleteVariable removes one variable.
func (c *Client) DeleteVariable(ctx context.Context, project, key string) (domain.Version, error) {
	path, err := projectPath(project, key)
	if err != nil {
		return 0, err
	}
	return c.version(c.do(ctx, http.MethodDelete, path, nil, nil, nil))
}

func (c *Client) version(body []byte, err error) (domain.Version, error) {
	if err != nil {
		return 0, err
	}
	var out apiv1.VersionResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, fmt.Errorf("api: decode version: %w", err)
	}
	return domain.Version(out.Version), nil
}

// do sends a request and decodes a JSON response into out when it is not
// nil. It returns the raw body for callers that decode it themselves.
func (c *Client) do(ctx context.Context, method, path string, header http.Header, in, out any) ([]byte, error) {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return nil, fmt.Errorf("api: encode request: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL.JoinPath(path).String(), body)
	if err != nil {
		return nil, fmt.Errorf("api: build request: %w", err)
	}
	for key, values := range header {
		req.Header[key] = values
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("api: read response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, decodeError(resp.StatusCode, raw)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return nil, fmt.Errorf("api: decode response: %w", err)
		}
	}
	return raw, nil
}

func decodeError(status int, raw []byte) error {
	var body apiv1.ErrorResponse
	if err := json.Unmarshal(raw, &body); err != nil || body.Error.Code == "" {
		return &Error{Status: status, Message: strings.TrimSpace(string(raw))}
	}
	return &Error{Status: status, Code: body.Error.Code, Message: body.Error.Message}
}

// projectPath validates the project name (and variable key) before building a
// route, so invalid input fails early with a domain error instead of reaching
// an unexpected route through "/" or "..".
func projectPath(project string, key ...string) (string, error) {
	if _, err := domain.NewProjectName(project); err != nil {
		return "", err
	}
	path := apiv1.BasePath + "/projects/" + project
	if len(key) == 0 {
		return path, nil
	}
	if key[0] == "" {
		return path + "/variables", nil
	}
	if _, err := domain.NewVariableKey(key[0]); err != nil {
		return "", err
	}
	return path + "/variables/" + key[0], nil
}
