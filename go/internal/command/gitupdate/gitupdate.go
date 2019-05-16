package gitupdate

import (
	"errors"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/accessverifier"
)

var (
	disallowedCommandError = errors.New("> GitLab: Disallowed command")
)

type Command struct {
	Config     *config.Config
	Args       *commandargs.CommandArgs
	ReadWriter *readwriter.ReadWriter
}

func (c *Command) Execute() error {
	args := c.Args.SshArgs

	if len(args) != 2 {
		return disallowedCommandError
	}

	repo := args[1]
	response, err := c.verifyAccess(repo)

	if err != nil {
		return err
	}

	return c.execCommand(response)
}

func (c *Command) verifyAccess(repo string) (*accessverifier.Response, error) {
	client, err := accessverifier.NewClient(c.Config, c.ReadWriter)

	if err != nil {
		return nil, err
	}

	return client.Execute(c.Args, c.Args.CommandType, repo)
}

func (c *Command) execCommand(response *accessverifier.Response) error {
	switch c.Args.CommandType {
	case commandargs.ReceivePack:
		return c.receivePack(response)
	}

	return nil
}
