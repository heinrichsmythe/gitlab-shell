package gitlabnet

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/testserver"
)

func TestClients(t *testing.T) {
	requests := []testserver.TestRequestHandler{
		{
			Path: "/api/v4/internal/hello",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)

				body := map[string]string{
					"value": "Hello",
				}
				json.NewEncoder(w).Encode(body)
			},
		},
		{
			Path: "/api/v4/internal/post_endpoint",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)

				b, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()

				require.NoError(t, err)

				body := map[string]string{
					"value": "Echo: " + string(b),
				}
				json.NewEncoder(w).Encode(body)
			},
		},
		{
			Path: "/api/v4/internal/auth",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				body := map[string]string{
					"value": r.Header.Get(secretHeaderName),
				}
				json.NewEncoder(w).Encode(body)
			},
		},
		{
			Path: "/api/v4/internal/error",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				body := map[string]string{
					"message": "Don't do that",
				}
				json.NewEncoder(w).Encode(body)
			},
		},
		{
			Path: "/api/v4/internal/broken",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				panic("Broken")
			},
		},
	}

	testCases := []struct {
		desc   string
		secret string
		server func([]testserver.TestRequestHandler) (func(), string, error)
	}{
		{
			desc:   "Socket client",
			secret: "sssh, it's a secret",
			server: testserver.StartSocketHttpServer,
		},
		{
			desc:   "Http client",
			secret: "sssh, it's a secret",
			server: testserver.StartHttpServer,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cleanup, url, err := tc.server(requests)
			defer cleanup()
			require.NoError(t, err)

			client, err := GetClient(&config.Config{GitlabUrl: url, Secret: tc.secret})
			require.NoError(t, err)

			testBrokenRequest(t, client)
			testSuccessfulGet(t, client)
			testSuccessfulPost(t, client)
			testMissing(t, client)
			testErrorMessage(t, client)
			testAuthenticationHeader(t, client)
		})
	}
}

func testSuccessfulGet(t *testing.T, client *GitlabClient) {
	t.Run("Successful get", func(t *testing.T) {
		response := map[string]string{}
		httpResponse, err := client.Get("/hello", &response)

		require.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, "Hello", response["value"])
		assert.Equal(t, httpResponse.StatusCode, 200)
	})
}

func testSuccessfulPost(t *testing.T, client *GitlabClient) {
	t.Run("Successful Post", func(t *testing.T) {
		data := map[string]string{"key": "value"}

		response := map[string]string{}
		httpResponse, err := client.Post("/post_endpoint", data, &response)

		require.NoError(t, err)

		assert.NoError(t, err)
		assert.Equal(t, "Echo: {\"key\":\"value\"}", response["value"])
		assert.Equal(t, httpResponse.StatusCode, 200)
	})
}

func testMissing(t *testing.T, client *GitlabClient) {
	t.Run("Missing error for GET", func(t *testing.T) {
		var response map[string]string
		_, err := client.Get("/missing", &response)

		assert.EqualError(t, err, "Internal API error (404)")
		assert.Nil(t, response)
	})

	t.Run("Missing error for POST", func(t *testing.T) {
		var response map[string]string
		_, err := client.Post("/missing", map[string]string{}, &response)

		assert.EqualError(t, err, "Internal API error (404)")
		assert.Nil(t, response)
	})
}

func testErrorMessage(t *testing.T, client *GitlabClient) {
	t.Run("Error with message for GET", func(t *testing.T) {
		var response map[string]string
		_, err := client.Get("/error", &response)

		assert.EqualError(t, err, "Don't do that")
		assert.Nil(t, response)
	})

	t.Run("Error with message for POST", func(t *testing.T) {
		var response map[string]string
		_, err := client.Post("/error", map[string]string{}, &response)

		assert.EqualError(t, err, "Don't do that")
		assert.Nil(t, response)
	})
}

func testBrokenRequest(t *testing.T, client *GitlabClient) {
	t.Run("Broken request for GET", func(t *testing.T) {
		var response map[string]string
		_, err := client.Get("/broken", &response)

		assert.EqualError(t, err, "Internal API unreachable")
		assert.Nil(t, response)
	})

	t.Run("Broken request for POST", func(t *testing.T) {
		var response map[string]string
		_, err := client.Post("/broken", map[string]string{}, &response)

		assert.EqualError(t, err, "Internal API unreachable")
		assert.Nil(t, response)
	})
}

func testAuthenticationHeader(t *testing.T, client *GitlabClient) {
	t.Run("Authentication headers for GET", func(t *testing.T) {
		var response map[string]string
		_, err := client.Get("/auth", &response)

		require.NoError(t, err)

		header, err := base64.StdEncoding.DecodeString(string(response["value"]))
		require.NoError(t, err)
		assert.Equal(t, "sssh, it's a secret", string(header))
	})

	t.Run("Authentication headers for POST", func(t *testing.T) {
		var response map[string]string
		_, err := client.Post("/auth", map[string]string{}, &response)

		require.NoError(t, err)

		header, err := base64.StdEncoding.DecodeString(string(response["value"]))
		require.NoError(t, err)
		assert.Equal(t, "sssh, it's a secret", string(header))
	})
}
