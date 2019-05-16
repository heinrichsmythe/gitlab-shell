package commandargs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSshArgs(t *testing.T) {
	testCases := []struct {
		desc       string
		sshCommand string
		args       []string
	}{
		{
			desc:       "Multiple words",
			sshCommand: "git-lfs-authenticate group/repo upload",
			args:       []string{"git-lfs-authenticate", "group/repo", "upload"},
		}, {
			desc:       "An argument with single quotes",
			sshCommand: "git-lfs-authenticate 'group repo' upload",
			args:       []string{"git-lfs-authenticate", "group repo", "upload"},
		}, {
			desc:       "An argument with double quotes",
			sshCommand: `git-lfs-authenticate "group repo" upload`,
			args:       []string{"git-lfs-authenticate", "group repo", "upload"},
		}, {
			desc:       "Additional spaces",
			sshCommand: "       git-receive-pack group/repo  ",
			args:       []string{"git-receive-pack", "group/repo"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, splitSshArgs(tc.sshCommand), tc.args)
		})
	}
}
