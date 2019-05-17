package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/discover"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/fallback"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/gitupdate"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/twofactorrecover"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/testhelper"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		desc         string
		arguments    []string
		config       *config.Config
		environment  map[string]string
		expectedType interface{}
	}{
		{
			desc:      "it returns a Discover command if the feature is enabled",
			arguments: []string{},
			config: &config.Config{
				GitlabUrl: "http+unix://gitlab.socket",
				Migration: config.MigrationConfig{Enabled: true, Features: []string{"discover"}},
			},
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "",
			},
			expectedType: &discover.Command{},
		},
		{
			desc:      "it returns a Fallback command no feature is enabled",
			arguments: []string{},
			config: &config.Config{
				GitlabUrl: "http+unix://gitlab.socket",
				Migration: config.MigrationConfig{Enabled: false},
			},
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "",
			},
			expectedType: &fallback.Command{},
		},
		{
			desc:      "it returns a TwoFactorRecover command if the feature is enabled",
			arguments: []string{},
			config: &config.Config{
				GitlabUrl: "http+unix://gitlab.socket",
				Migration: config.MigrationConfig{Enabled: true, Features: []string{"2fa_recovery_codes"}},
			},
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "2fa_recovery_codes",
			},
			expectedType: &twofactorrecover.Command{},
		},
		{
			desc:      "it returns a generate gitupdate command if the git-receive-pack feature is enabled",
			arguments: []string{},
			config: &config.Config{
				GitlabUrl: "http+unix://gitlab.socket",
				Migration: config.MigrationConfig{Enabled: true, Features: []string{"git-receive-pack"}},
			},
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "git-receive-pack",
			},
			expectedType: &gitupdate.Command{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			restoreEnv := testhelper.TempEnv(tc.environment)
			defer restoreEnv()

			command, err := New(tc.arguments, tc.config, nil)

			assert.NoError(t, err)
			assert.IsType(t, tc.expectedType, command)
		})
	}
}

func TestFailingNew(t *testing.T) {
	t.Run("It returns an error when SSH_CONNECTION is not set", func(t *testing.T) {
		restoreEnv := testhelper.TempEnv(map[string]string{})
		defer restoreEnv()

		_, err := New([]string{}, &config.Config{}, nil)

		assert.Error(t, err, "Only ssh allowed")
	})
}
