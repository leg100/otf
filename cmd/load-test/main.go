package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/hashicorp/go-tfe"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal/integration"
)

const (
	numWorkspaces = 100
	namespace     = "loadtest"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		// Stop handling ^C; another ^C will exit the program.
		cancel()
	}()

	if err := run(ctx); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	deploy, _, err := integration.NewKubeDeploy(ctx, integration.KubeDeployConfig{
		Namespace:          namespace,
		OpenBrowser:        true,
		CacheVolumeEnabled: true,
	})
	if err != nil {
		return fmt.Errorf("deploying to kind: %w", err)
	}
	defer func() {
		if err := deploy.Close(false); err != nil {
			log.Printf("error closing deploy: %s\n", err.Error())
		}
	}()

	// Create randomly named organization.
	orgName := petname.Generate(1, "")
	org, err := deploy.Organizations.Create(ctx, tfe.OrganizationCreateOptions{
		Name:  new(orgName),
		Email: new("bollocks@morebollocks.bollocks"),
	})
	if err != nil {
		return err
	}
	log.Printf("created organization: %s\n", org.Name)

	// Create umpteen workspaces, each with a made up name.
	workspaces := make([]*tfe.Workspace, numWorkspaces)
	for i := range numWorkspaces {
		ws, err := deploy.Workspaces.Create(ctx, org.Name, tfe.WorkspaceCreateOptions{
			Name: new(petname.Generate(2, "-")),
		})
		if err != nil {
			return err
		}
		log.Printf("created workspace: %s\n", ws.Name)
		workspaces[i] = ws
	}

	// For each workspace, create a run and wait for its kubernetes pod to
	// succeed.
	runs := make([]*tfe.Run, len(workspaces))

	// Prime cache by creating one run and wait to complete and only then create
	// the others.
	run, err := createRun(ctx, deploy, workspaces[0])
	if err != nil {
		return fmt.Errorf("creating first run: %w", err)
	}
	log.Printf("creating first run: %s\n", run.ID)
	// Pod should succeed and run should reach planned status
	_, err = deploy.WaitPodSucceed(ctx, run.ID, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("waiting for pod to succeed: %w", err)
	}
	log.Printf("first run completed: %s\n", run.ID)

	log.Printf("creating other runs\n")
	for i, ws := range workspaces[1:] {
		run, err := createRun(ctx, deploy, ws)
		if err != nil {
			return fmt.Errorf("creating run: %w", err)
		}
		runs[i] = run
	}
	// Wait until user sends ^C
	return deploy.Wait()
}

func createRun(ctx context.Context, deploy *integration.KubeDeploy, ws *tfe.Workspace) (*tfe.Run, error) {
	cv, err := deploy.ConfigurationVersions.Create(ctx, ws.ID, tfe.ConfigurationVersionCreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("creating config version: %w", err)
	}

	tarball, err := os.Open("./loadtest/loadtest.tar.gz")
	if err != nil {
		return nil, fmt.Errorf("opening tarball: %w", err)
	}
	err = deploy.ConfigurationVersions.UploadTarGzip(ctx, cv.UploadURL, tarball)
	if err != nil {
		return nil, fmt.Errorf("uploading config version: %w", err)
	}
	return deploy.Runs.Create(ctx, tfe.RunCreateOptions{
		Workspace:            ws,
		ConfigurationVersion: cv,
	})
}
