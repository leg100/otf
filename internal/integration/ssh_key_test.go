package integration

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/workspace"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestSSHKeyPrivateModule tests that a workspace with an SSH key assigned can
// successfully download a private terraform module via SSH git during a run.
func TestSSHKeyPrivateModule(t *testing.T) {
	integrationTest(t)

	// Generate an ED25519 key pair: the private key goes into OTF, the public
	// key authorises access to our test SSH git server.
	pubKeyRaw, privKeyRaw, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	privPEM, err := ssh.MarshalPrivateKey(privKeyRaw, "")
	require.NoError(t, err)
	privKeyBytes := pem.EncodeToMemory(privPEM)

	sshPub, err := ssh.NewPublicKey(pubKeyRaw)
	require.NoError(t, err)

	// Build a bare git repository containing a simple terraform module.
	repoDir := t.TempDir()
	mustRunGit(t, "init", "--bare", repoDir)

	workDir := t.TempDir()
	mustRunGit(t, "-C", workDir, "init")
	mustRunGit(t, "-C", workDir, "config", "user.email", "test@example.com")
	mustRunGit(t, "-C", workDir, "config", "user.name", "Test")
	err = os.WriteFile(filepath.Join(workDir, "main.tf"), []byte(`locals {
  message = "hello from ssh module"
}`), 0o600)
	require.NoError(t, err)
	mustRunGit(t, "-C", workDir, "add", ".")
	mustRunGit(t, "-C", workDir, "commit", "-m", "initial")
	mustRunGit(t, "-C", workDir, "push", repoDir, "HEAD:refs/heads/main")
	mustRunGit(t, "-C", repoDir, "symbolic-ref", "HEAD", "refs/heads/main")

	// Start an in-process SSH git server that only accepts the generated key.
	sshAddr := startSSHGitServer(t, repoDir, sshPub)

	// Start both a daemon and an agent.
	daemon, org, ctx := setup(t)
	agent, _ := daemon.startAgent(t, ctx, org.Name, nil, "")

	// Register the SSH key with OTF via the UI.
	browser.New(t, ctx, func(page playwright.Page) {
		_, err := page.Goto(daemon.URL(paths.Organization(org.Name)))
		require.NoError(t, err)

		err = page.Locator(`//li[@id='menu-item-ssh-keys']/a`).Click()
		require.NoError(t, err)

		err = page.Locator("input#name").Fill("test-key")
		require.NoError(t, err)

		err = page.Locator("textarea#private-key").Fill(string(privKeyBytes))
		require.NoError(t, err)

		err = page.Locator("button#create-button").Click()
		require.NoError(t, err)

		// confirm key created
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("created ssh key: test-key")
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//table//tr[@id='ssh-key-item-test-key']/td[1]`)).ToHaveText("test-key")
		require.NoError(t, err)
	})

	// two tests: one run on the daemon, one via the agent.
	tests := []struct {
		name   string
		mode   workspace.ExecutionMode
		poolID *resource.TfeID
	}{
		{
			"execute run via daemon", workspace.RemoteExecutionMode, nil,
		},
		{
			"execute run via agent", workspace.AgentExecutionMode, &agent.AgentPool.ID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create a workspace and assign the SSH key to it.
			ws, err := daemon.Workspaces.CreateWorkspace(ctx, workspace.CreateOptions{
				Name:          new("ws-" + string(tt.mode)),
				Organization:  &org.Name,
				ExecutionMode: new(tt.mode),
				AgentPoolID:   tt.poolID,
			})
			require.NoError(t, err)

			// Update workspace via UI to use ssh key.
			browser.New(t, ctx, func(page playwright.Page) {
				_, err := page.Goto(daemon.URL(paths.Workspace(ws.ID)))
				require.NoError(t, err)

				// navigate to workspace settings
				err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
				require.NoError(t, err)
				screenshot(t, page, "workspace_settings")

				// navigate to ssh settings
				err = page.Locator(`//li[@id='menu-item-ssh-key']/a`).Click()
				require.NoError(t, err)

				selectValues := []string{"test-key"}
				_, err = page.Locator(`//select[@id="ssh-key-id-selector"]`).SelectOption(playwright.SelectOptionValues{
					ValuesOrLabels: &selectValues,
				})
				require.NoError(t, err)

				err = page.Locator("button#update-ssh-key").Click()
				require.NoError(t, err)

				// confirm workspace updated to use key
				err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace")
				require.NoError(t, err)
			})

			// Build a terraform config that sources the private module over SSH.  The
			// module itself has no resources, so no providers need to be downloaded;
			// the sole purpose of init is to clone the module via the SSH key.
			tfconfig := fmt.Sprintf(`
terraform {
  cloud {
    hostname     = "%s"
    organization = "%s"
    workspaces {
      name = "%s"
    }
  }
}

module "private" {
  source = "git::ssh://git@%s/"
}
`, daemon.System.Hostname(), org.Name, ws.Name, sshAddr)

			// Pack and upload the config directly so the OTF runner executes the init.
			root := t.TempDir()
			err = os.WriteFile(filepath.Join(root, "main.tf"), []byte(tfconfig), 0o600)
			require.NoError(t, err)
			tarball, err := internal.Pack(root)
			require.NoError(t, err)

			cv := daemon.createConfigurationVersion(t, ctx, ws, &configversion.CreateOptions{})
			err = daemon.Configs.UploadConfig(ctx, cv.ID, tarball)
			require.NoError(t, err)

			// Trigger a run.  If the SSH key is wired up correctly, the runner will set
			// GIT_SSH_COMMAND and terraform init will successfully clone the module.
			// The plan will show no changes (the module has no resources) and the run
			// will reach PlannedAndFinished.
			run := daemon.createRun(t, ctx, ws, cv, nil)
			daemon.waitRunStatus(t, ctx, run.ID, runstatus.PlannedAndFinished)
		})
	}
}

// mustRunGit runs a git command and fails the test on error.
func mustRunGit(t *testing.T, args ...string) {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	require.NoError(t, err, string(out))
}

// startSSHGitServer starts an in-process SSH server that serves a git
// repository at repoPath, accepting only connections authenticated with
// authorizedKey.  It returns the "host:port" address of the listener.
func startSSHGitServer(t *testing.T, repoPath string, authorizedKey ssh.PublicKey) string {
	t.Helper()

	// Generate a throwaway host key for the server.
	_, hostPrivKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	hostSigner, err := ssh.NewSignerFromKey(hostPrivKey)
	require.NoError(t, err)

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if bytes.Equal(key.Marshal(), authorizedKey.Marshal()) {
				return &ssh.Permissions{}, nil
			}
			return nil, fmt.Errorf("unauthorized key")
		},
	}
	config.AddHostKey(hostSigner)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // listener closed
			}
			go handleSSHConn(conn, config, repoPath)
		}
	}()

	t.Cleanup(func() { listener.Close() })
	return listener.Addr().String()
}

func handleSSHConn(conn net.Conn, config *ssh.ServerConfig, repoPath string) {
	defer conn.Close()

	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		ch, requests, err := newChan.Accept()
		if err != nil {
			return
		}
		go handleSSHSession(ch, requests, repoPath)
	}
}

func handleSSHSession(ch ssh.Channel, requests <-chan *ssh.Request, repoPath string) {
	defer ch.Close()

	for req := range requests {
		if req.Type != "exec" {
			if req.WantReply {
				req.Reply(false, nil)
			}
			continue
		}
		if req.WantReply {
			req.Reply(true, nil)
		}

		// Serve the repository regardless of what path the client requested.
		cmd := exec.Command("git-upload-pack", repoPath)
		cmd.Stdin = ch
		cmd.Stdout = ch
		cmd.Stderr = ch.Stderr()

		var exitCode uint32
		if err := cmd.Run(); err != nil {
			exitCode = 1
		}

		exitMsg := make([]byte, 4)
		binary.BigEndian.PutUint32(exitMsg, exitCode)
		ch.SendRequest("exit-status", false, exitMsg)
		return
	}
}
