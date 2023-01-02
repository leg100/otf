package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/leg100/otf"
	"github.com/spf13/cobra"
)

func RegistrySessionsCommand(svc otf.RegistrySessionService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry-sessions",
		Short: "Registry session tokens",
	}

	var organization string
	cmd.PersistentFlags().StringVar(&organization, "organization", "", "The organization of the registry. Required.")
	cmd.MarkPersistentFlagRequired("organization")

	cmd.AddCommand(RegistrySessionsGetCommand(svc, &organization))

	return cmd
}

func RegistrySessionsGetCommand(svc otf.RegistrySessionService, organization *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [hostname]",
		Short: "Retrieve registry session token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if organization == nil {
				return fmt.Errorf("organization not set")
			}

			session, err := svc.CreateRegistrySession(cmd.Context(), *organization)
			if err != nil {
				return err
			}

			// terraform expects a json object containing token to be printed, like so:
			// {
			//   "token": "example-token-value"
			// }

			type output struct {
				Token string
			}
			out, err := json.MarshalIndent(output{Token: session.Token()}, "", "\t")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(out))

			return nil
		},
	}

	return cmd
}
