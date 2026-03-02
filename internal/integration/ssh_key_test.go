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
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sshkey"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestSSHKeyPrivateModule tests that a workspace with an SSH key assigned can
// successfully download a private terraform module via SSH git during a run.
func TestSSHKeyPrivateModule(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

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

	// Register the SSH key with OTF.
	key, err := svc.SSHKeys.Create(ctx, sshkey.CreateOptions{
		Organization: org.Name,
		Name:         "test-key",
		PrivateKey:   string(privKeyBytes),
	})
	require.NoError(t, err)

	// Create a workspace and assign the SSH key to it.
	ws := svc.createWorkspace(t, ctx, org)
	_, err = svc.Workspaces.Update(ctx, ws.ID, workspace.UpdateOptions{
		UpdateSSHKeyOptions: &workspace.UpdateSSHKeyOptions{
			SSHKeyID: &key.ID,
		},
	})
	require.NoError(t, err)

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
`, svc.System.Hostname(), org.Name, ws.Name, sshAddr)

	// Pack and upload the config directly so the OTF runner executes the init.
	root := t.TempDir()
	err = os.WriteFile(filepath.Join(root, "main.tf"), []byte(tfconfig), 0o600)
	require.NoError(t, err)
	tarball, err := internal.Pack(root)
	require.NoError(t, err)

	cv := svc.createConfigurationVersion(t, ctx, ws, &configversion.CreateOptions{})
	err = svc.Configs.UploadConfig(ctx, cv.ID, tarball)
	require.NoError(t, err)

	// Trigger a run.  If the SSH key is wired up correctly, the runner will set
	// GIT_SSH_COMMAND and terraform init will successfully clone the module.
	// The plan will show no changes (the module has no resources) and the run
	// will reach PlannedAndFinished.
	run := svc.createRun(t, ctx, ws, cv, nil)
	svc.waitRunStatus(t, ctx, run.ID, runstatus.PlannedAndFinished)
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
