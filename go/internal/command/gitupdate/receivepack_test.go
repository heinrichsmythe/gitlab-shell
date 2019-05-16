package gitupdate

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/accessverifier"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/testserver"

	pb "gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
)

func (s *testSSHServer) SSHReceivePack(stream pb.SSHService_SSHReceivePackServer) error {
	req, err := stream.Recv()

	if err != nil {
		log.Fatal(err)
	}

	response := []byte("ReceivePack: " + req.GlId + " " + req.Repository.GlRepository)
	stream.Send(&pb.SSHReceivePackResponse{Stdout: response})

	return nil
}

func TestReceivePack(t *testing.T) {
	gitalyAddress, err := startSSHServer()
	require.NoError(t, err)

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
					"gitaly": map[string]interface{}{
						"repository": map[string]interface{}{
							"storage_name":                     "storage_name",
							"relative_path":                    "relative_path",
							"git_object_directory":             "path/to/git_object_directory",
							"git_alternate_object_directories": []string{"path/to/git_alternate_object_directory"},
							"gl_repository":                    repo,
							"gl_project_path":                  "group/project-path",
						},
						"address": gitalyAddress,
						"token":   "token",
					},
				}
				json.NewEncoder(w).Encode(body)
			},
		},
	}

	cleanup, url := setup(t, requests)
	defer cleanup()

	output := &bytes.Buffer{}
	input := &bytes.Buffer{}

	cmd := &Command{
		Config:     &config.Config{GitlabUrl: url},
		Args:       &commandargs.CommandArgs{GitlabKeyId: userId, CommandType: commandargs.ReceivePack, SshArgs: []string{"git-receive-pack", repo}},
		ReadWriter: &readwriter.ReadWriter{ErrOut: output, Out: output, In: input},
	}

	err = cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "ReceivePack: "+userId+" "+repo, output.String())
}
