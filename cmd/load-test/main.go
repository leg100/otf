package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/hashicorp/go-tfe"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/integration"
)

const numWorkspaces = 100

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
	deploy, err := integration.NewKubeDeploy(ctx, "", "otfd", ".")
	if err != nil {
		return fmt.Errorf("deploying to kind: %w", err)
	}

	org, err := deploy.Organizations.Create(ctx, tfe.OrganizationCreateOptions{
		Name:  internal.Ptr("acme"),
		Email: internal.Ptr("bollocks@morebollocks.bollocks"),
	})
	if err != nil {
		return err
	}
	log.Printf("created organization: %s\n", org.Name)

	// Create umpteen workspaces, each with a made up name.
	workspaces := make([]*tfe.Workspace, numWorkspaces)
	for i := range numWorkspaces {
		ws, err := deploy.Workspaces.Create(ctx, org.Name, tfe.WorkspaceCreateOptions{
			Name: internal.Ptr(petname.Generate(2, "-")),
		})
		if err != nil {
			return err
		}
		log.Printf("created workspace: %s\n", ws.Name)
		workspaces[i] = ws
	}
	return nil
}
