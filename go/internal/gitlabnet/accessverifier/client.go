package accessverifier

import (
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet"
)

const (
	sshEnv     = "ssh"
	anyChanges = "_any"
)

type Client struct {
	config     *config.Config
	client     *gitlabnet.GitlabClient
	readWriter *readwriter.ReadWriter
}

type Request struct {
	Action   commandargs.CommandType `json:"action"`
	Repo     string                  `json:"project"`
	Changes  string                  `json:"changes"`
	Env      string                  `json:"env"`
	KeyId    string                  `json:"key_id,omitempty"`
	Username string                  `json:"username,omitempty"`
}

type GitalyRepo struct {
	StorageName                   string   `json:"storage_name"`
	RelativePath                  string   `json:"relative_path"`
	GitObjectDirectory            string   `json:"git_object_directory"`
	GitAlternateObjectDirectories []string `json:"git_alternate_object_directories"`
	RepoName                      string   `json:"gl_repository"`
	ProjectPath                   string   `json:"gl_project_path"`
}

type Gitaly struct {
	Repo    GitalyRepo `json:"repository"`
	Address string     `json:"address"`
	Token   string     `json:"token"`
}

type CustomPayloadData struct {
	ApiEndpoints []string `json:"api_endpoints"`
	Username     string   `json:"gl_username"`
	PrimaryRepo  string   `json:"primary_repo"`
	InfoMessage  string   `json:"info_message"`
	UserId       string   `json:"gl_id,omitempty"`
}

type CustomPayload struct {
	Action string            `json:"action"`
	Data   CustomPayloadData `json:"data"`
}

type Response struct {
	Success          bool          `json:"status"`
	Message          string        `json:"message"`
	Repo             string        `json:"gl_repository"`
	UserId           string        `json:"gl_id"`
	Username         string        `json:"gl_username"`
	GitConfigOptions []string      `json:"git_config_options"`
	Gitaly           Gitaly        `json:"gitaly"`
	GitProtocol      string        `json:"git_protocol"`
	Payload          CustomPayload `json:"payload"`
	ConsoleMessages  []string      `json:"gl_console_messages"`
	StatusCode       int
}

func NewClient(config *config.Config, readWriter *readwriter.ReadWriter) (*Client, error) {
	client, err := gitlabnet.GetClient(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating http client: %v", err)
	}

	return &Client{config: config, client: client, readWriter: readWriter}, nil
}

func (c *Client) Execute(args *commandargs.CommandArgs, action commandargs.CommandType, repo string) (*Response, error) {
	request := &Request{Action: action, Repo: repo, Env: sshEnv, Changes: anyChanges}
	if args.GitlabUsername != "" {
		request.Username = args.GitlabUsername
	} else {
		request.KeyId = args.GitlabKeyId
	}

	response, err := c.performRequest(request)

	if err != nil {
		return nil, err
	}

	for _, msg := range response.ConsoleMessages {
		fmt.Fprintf(c.readWriter.Out, "> GitLab: %v\n", msg)
	}

	return response, nil
}

func (c *Client) performRequest(request *Request) (*Response, error) {
	response := &Response{}
	httpResponse, err := c.client.Post("/allowed", request, response)

	if err != nil {
		return nil, err
	}

	response.StatusCode = httpResponse.StatusCode

	if response.Success {
		return response, nil
	} else {
		return nil, errors.New(response.Message)
	}
}

func (r *Response) IsCustomAction() bool {
	return r.StatusCode == http.StatusMultipleChoices
}
