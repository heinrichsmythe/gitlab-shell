require_relative 'spec_helper'

require 'open3'
require 'json'
require 'base64'

describe 'Custom bin/gitlab-shell git-receive-pack' do
  include_context 'gitlab shell'

  def mock_server(server)
    server.mount_proc('/geo/proxy_git_push_ssh/info_refs') do |req, res|
      res.content_type = 'application/json'
      res.status = 200

      res.body = {"result" => "#{Base64.encode64('custom')}"}.to_json
    end

    server.mount_proc('/geo/proxy_git_push_ssh/push') do |req, res|
      res.content_type = 'application/json'
      res.status = 200

      output = JSON.parse(req.body)['output']

      res.body = {"result" => output}.to_json
    end

    server.mount_proc('/api/v4/internal/allowed') do |req, res|
      res.content_type = 'application/json'

      key_id = req.query['key_id'] || req.query['username']

      unless key_id
        body = JSON.parse(req.body)
        key_id = body['key_id'] || body['username'].to_s
      end

      case key_id
      when '100', 'someone' then
        res.status = 300
        body = {
          "gl_id" => "user-100",
          "status" => true,
          "payload" => {
            "action" => "geo_proxy_to_primary",
            "data" => {
              "api_endpoints" => ["/geo/proxy_git_push_ssh/info_refs", "/geo/proxy_git_push_ssh/push"],
              "gl_username" =>   "custom",
              "primary_repo" =>  "https://repo/path",
              "info_message" => "info_message",
            },
          },
          "gl_console_messages" => ["console", "message"]
        }
        res.body = body.to_json
      else
        res.status = 403
      end
    end
  end

  shared_examples 'dialog for performing a custom action' do
    context 'when the user agrees to regenerate keys' do
      def verify_successful_call!(cmd)
        Open3.popen2e(env, cmd) do |stdin, stdout|
          expect(stdout.gets).to eq("> GitLab: console\n")
          expect(stdout.gets).to eq("> GitLab: message\n")

          expect(stdout.gets).to eq("> GitLab: info_message\n")
          expect(stdout.gets).to eq("custom\n")

          stdin.puts("input")
          stdin.close

          expect(stdout.flush.read).to eq("input\n")
        end
      end

      context 'when key is provided' do
        let(:cmd) { "#{gitlab_shell_path} -c/usr/share/webapps/gitlab-shell/bin/gitlab-shell key-100" }

        it 'custom action is performed' do
          verify_successful_call!(cmd)
        end
      end

      context 'when username is provided' do
        let(:cmd) { "#{gitlab_shell_path} -c/usr/share/webapps/gitlab-shell/bin/gitlab-shell username-someone" }

        it 'custom action is performed' do
          verify_successful_call!(cmd)
        end
      end
    end

    context 'when API error occurs' do
      let(:cmd) { "#{gitlab_shell_path} -c/usr/share/webapps/gitlab-shell/bin/gitlab-shell key-101" }

      it 'custom action is not performed' do
        Open3.popen2e(env, cmd) do |stdin, stdout|
          expect(stdout.gets).to eq(inaccessible_error)
        end
      end
    end
  end

  let(:env) { {'SSH_CONNECTION' => 'fake', 'SSH_ORIGINAL_COMMAND' => 'git-receive-pack group/repo' } }

  describe 'without go features' do
    before(:context) do
      write_config(
        "gitlab_url" => "http+unix://#{CGI.escape(tmp_socket_path)}",
      )
    end

    it_behaves_like 'dialog for performing a custom action' do
      let(:inaccessible_error) { "> GitLab: API is not accessible\n" }
    end
  end

  describe 'with go features' do
    before(:context) do
      write_config(
        "gitlab_url" => "http+unix://#{CGI.escape(tmp_socket_path)}",
        "migration" => { "enabled" => true,
                        "features" => ["git-receive-pack"] }
      )
    end

    it_behaves_like 'dialog for performing a custom action' do
      let(:inaccessible_error) { "Internal API error (403)\n" }
    end
  end
end
