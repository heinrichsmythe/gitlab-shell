package commandargs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/testhelper"
)

func TestParseSuccess(t *testing.T) {
	testCases := []struct {
		desc         string
		arguments    []string
		environment  map[string]string
		expectedArgs *CommandArgs
	}{
		// Setting the used env variables for every case to ensure we're
		// not using anything set in the original env.
		{
			desc: "It sets discover as the command when the command string was empty",
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "",
			},
			expectedArgs: &CommandArgs{CommandType: Discover},
		},
		{
			desc: "It finds the key id in any passed arguments",
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "",
			},
			arguments:    []string{"hello", "key-123"},
			expectedArgs: &CommandArgs{CommandType: Discover, GitlabKeyId: "123"},
		}, {
			desc: "It finds the username in any passed arguments",
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "",
			},
			arguments:    []string{"hello", "username-jane-doe"},
			expectedArgs: &CommandArgs{CommandType: Discover, GitlabUsername: "jane-doe"},
		}, {
			desc: "It parses 2fa_recovery_codes command",
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "2fa_recovery_codes",
			},
			expectedArgs: &CommandArgs{SshArgs: []string{"2fa_recovery_codes"}, CommandType: TwoFactorRecover},
		}, {
			desc: "It parses git-receive-pack command",
			environment: map[string]string{
				"SSH_CONNECTION":       "1",
				"SSH_ORIGINAL_COMMAND": "git receive-pack 'group/repo'",
			},
			expectedArgs: &CommandArgs{SshArgs: []string{"git-receive-pack", "group/repo"}, CommandType: ReceivePack},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			restoreEnv := testhelper.TempEnv(tc.environment)
			defer restoreEnv()

			result, err := Parse(tc.arguments)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedArgs, result)
		})
	}
}

func TestParseFailure(t *testing.T) {
	t.Run("It fails if SSH connection is not set", func(t *testing.T) {
		_, err := Parse([]string{})

		assert.Error(t, err, "Only ssh allowed")
	})

}
