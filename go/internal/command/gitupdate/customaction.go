package gitupdate

import (
	"encoding/base64"
	"errors"
	"fmt"

	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet"
	"gitlab.com/gitlab-org/gitlab-shell/go/internal/gitlabnet/accessverifier"
)

type CustomRequest struct {
	Data   accessverifier.CustomPayloadData `json:"data"`
	Output string                           `json:"output"`
}

type CustomResponse struct {
	Result  string `json:"result"`
	Message string `json:"message"`
}

func (c *Command) processCustomAction(response *accessverifier.Response) error {
	data := response.Payload.Data
	apiEndpoints := data.ApiEndpoints

	if len(apiEndpoints) == 0 {
		return errors.New("Custom action error: Empty Api endpoints")
	}

	c.displayCustomInfoMessage(data.InfoMessage)

	return c.processApiEndpoints(response)
}

func (c *Command) displayCustomInfoMessage(infoMessage string) {
	if infoMessage != "" {
		fmt.Fprintf(c.ReadWriter.Out, "> GitLab: %v\n", infoMessage)
	}
}

func (c *Command) processApiEndpoints(response *accessverifier.Response) error {
	client, err := gitlabnet.GetClient(c.Config)

	if err != nil {
		return err
	}

	data := response.Payload.Data
	request := &CustomRequest{Data: data}
	request.Data.UserId = response.UserId

	for _, endpoint := range data.ApiEndpoints {
		cr := &CustomResponse{}
		_, err := client.DoRequest("POST", endpoint, request, cr)

		if err != nil {
			return err
		}

		output, err := c.displayCustomOutput(cr.Result)

		if err != nil {
			return err
		}

		request.Output = output
	}

	return nil
}

func (c *Command) displayCustomOutput(result string) (string, error) {
	decodedOutput, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(c.ReadWriter.Out, string(decodedOutput))

	var output string
	fmt.Fscan(c.ReadWriter.In, &output)

	return base64.StdEncoding.EncodeToString([]byte(output)), nil
}
