package gitlabnet

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/testserver"
)

const (
	username = "basic_auth_user"
	password = "basic_auth_password"
)

func TestBasicAuthSettings(t *testing.T) {
	requests := []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/get_endpoint",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)

				body := map[string]string{
					"value": r.Header.Get("Authorization"),
				}
				json.NewEncoder(w).Encode(body)
			},
		},
		{
			Path: "/api/v4/internal/post_endpoint",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)

				body := map[string]string{
					"value": r.Header.Get("Authorization"),
				}
				json.NewEncoder(w).Encode(body)
			},
		},
	}
	config := &config.Config{HttpSettings: config.HttpSettingsConfig{User: username, Password: password}}

	client, cleanup := setup(t, config, requests)
	defer cleanup()

	var response map[string]string
	_, err := client.Get("/get_endpoint", &response)
	require.NoError(t, err)
	testBasicAuthHeaders(t, response["value"])

	_, err = client.Post("/post_endpoint", nil, &response)
	require.NoError(t, err)
	testBasicAuthHeaders(t, response["value"])
}

func testBasicAuthHeaders(t *testing.T, basicAuth string) {
	headerParts := strings.Split(basicAuth, " ")
	assert.Equal(t, "Basic", headerParts[0])

	credentials, err := base64.StdEncoding.DecodeString(headerParts[1])
	require.NoError(t, err)

	assert.Equal(t, username+":"+password, string(credentials))
}

func TestEmptyBasicAuthSettings(t *testing.T) {
	requests := []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/empty_basic_auth",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "", r.Header.Get("Authorization"))
				json.NewEncoder(w).Encode(map[string]string{})
			},
		},
	}

	client, cleanup := setup(t, &config.Config{}, requests)
	defer cleanup()

	var response map[string]string
	_, err := client.Get("/empty_basic_auth", &response)
	require.NoError(t, err)
}

func setup(t *testing.T, config *config.Config, requests []testserver.TestRequestHandler) (*GitlabClient, func()) {
	cleanup, url, err := testserver.StartHttpServer(requests)
	require.NoError(t, err)

	config.GitlabUrl = url
	client, err := GetClient(config)
	require.NoError(t, err)

	return client, cleanup
}
