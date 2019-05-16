package gitupdate

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/accessverifier"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/testserver"
)

func TestCustomReceivePack(t *testing.T) {
	repo := "group/repo"
	userId := "1"

	requests := []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/allowed",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				require.NoError(t, err)

				var request *accessverifier.Request
				json.Unmarshal(b, &request)

				assert.Equal(t, "1", request.KeyId)

				body := map[string]interface{}{
					"status": true,
					"gl_id":  "1",
					"payload": map[string]interface{}{
						"action": "geo_proxy_to_primary",
						"data": map[string]interface{}{
							"api_endpoints": []string{"/geo/proxy_git_push_ssh/info_refs", "/geo/proxy_git_push_ssh/push"},
							"gl_username":   "custom",
							"primary_repo":  "https://repo/path",
							"info_message":  "info_message",
						},
					},
				}
				w.WriteHeader(http.StatusMultipleChoices)
				json.NewEncoder(w).Encode(body)
			},
		},
		{
			Path: "/geo/proxy_git_push_ssh/info_refs",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				require.NoError(t, err)

				var request *CustomRequest
				json.Unmarshal(b, &request)

				assert.Equal(t, userId, request.Data.UserId)
				assert.Equal(t, "", request.Output)

				body := map[string]interface{}{
					"result": "Y3VzdG9t",
				}
				json.NewEncoder(w).Encode(body)
			},
		},
		{
			Path: "/geo/proxy_git_push_ssh/push",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				require.NoError(t, err)

				var request *CustomRequest
				json.Unmarshal(b, &request)

				assert.Equal(t, userId, request.Data.UserId)
				assert.Equal(t, "aW5wdXQ=", request.Output)

				body := map[string]interface{}{
					"result": "b3V0cHV0",
				}
				json.NewEncoder(w).Encode(body)
			},
		},
	}

	cleanup, url := setup(t, requests)
	defer cleanup()

	output := &bytes.Buffer{}
	input := bytes.NewBufferString("input")

	cmd := &Command{
		Config:     &config.Config{GitlabUrl: url},
		Args:       &commandargs.CommandArgs{GitlabKeyId: userId, CommandType: commandargs.ReceivePack, SshArgs: []string{"git-receive-pack", repo}},
		ReadWriter: &readwriter.ReadWriter{ErrOut: output, Out: output, In: input},
	}

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "> GitLab: info_message\ncustom\noutput\n", output.String())
}
