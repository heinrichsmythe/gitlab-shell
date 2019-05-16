package gitupdate

import (
	"context"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/accessverifier"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/handler"

	"gitlab.com/gitlab-org/gitaly/client"
	"google.golang.org/grpc"

	pb "gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
)

func (c *Command) receivePack(response *accessverifier.Response) error {
	if response.IsCustomAction() {
		return c.processCustomAction(response)
	}

	gc := &handler.GitalyCommand{
		Config:      c.Config,
		ServiceName: string(commandargs.ReceivePack),
		Address:     response.Gitaly.Address,
		Token:       response.Gitaly.Token,
	}

	repo := response.Gitaly.Repo
	request := &pb.SSHReceivePackRequest{
		Repository: &pb.Repository{
			StorageName:                   repo.StorageName,
			RelativePath:                  repo.RelativePath,
			GitObjectDirectory:            repo.GitObjectDirectory,
			GitAlternateObjectDirectories: repo.GitAlternateObjectDirectories,
			GlRepository:                  repo.RepoName,
			GlProjectPath:                 repo.ProjectPath,
		},
		GlId:             response.UserId,
		GlUsername:       response.Username,
		GitProtocol:      response.GitProtocol,
		GitConfigOptions: response.GitConfigOptions,
	}

	return gc.RunGitalyCommand(func(ctx context.Context, conn *grpc.ClientConn) (int32, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		rw := c.ReadWriter
		return client.ReceivePack(ctx, conn, rw.In, rw.Out, rw.ErrOut, request)
	})
}
