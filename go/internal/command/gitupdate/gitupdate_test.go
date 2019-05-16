package gitupdate

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/accessverifier"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/testserver"

	"google.golang.org/grpc"

	pb "gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
)

type testSSHServer struct{}

func (s *testSSHServer) SSHUploadPack(stream pb.SSHService_SSHUploadPackServer) error {
	return nil
}

func (s *testSSHServer) SSHUploadArchive(stream pb.SSHService_SSHUploadArchiveServer) error {
	return nil
}

func startSSHServer() (string, error) {
	server := grpc.NewServer()

	tempDir, _ := ioutil.TempDir("", "gitlab-shell-test-api")
	gitalySocketPath := path.Join(tempDir, "gitaly.sock")

	listener, err := net.Listen("unix", gitalySocketPath)

	if err != nil {
		return "", err
	}

	pb.RegisterSSHServiceServer(server, &testSSHServer{})

	go server.Serve(listener)

	gitalySocketUrl := "unix:" + gitalySocketPath

	return gitalySocketUrl, nil
}

func setup(t *testing.T, requests []testserver.TestRequestHandler) (func(), string) {
	cleanup, url, err := testserver.StartHttpServer(requests)
	require.NoError(t, err)

	return cleanup, url
}

func TestForbiddenAccess(t *testing.T) {
	requests := []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/allowed",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				require.NoError(t, err)

				var request *accessverifier.Request
				json.Unmarshal(b, &request)

				assert.Equal(t, "disallowed", request.KeyId)

				body := map[string]interface{}{
					"status":  false,
					"message": "Disallowed by API call",
				}
				w.WriteHeader(http.StatusForbidden)
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
		Args:       &commandargs.CommandArgs{GitlabKeyId: "disallowed", SshArgs: []string{"git-receive-pack", "group/repo"}},
		ReadWriter: &readwriter.ReadWriter{ErrOut: output, Out: output, In: input},
	}

	err := cmd.Execute()
	assert.Equal(t, "Disallowed by API call", err.Error())
}
