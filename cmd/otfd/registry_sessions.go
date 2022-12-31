package main

func RegistrySessionsCommand(factory http.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Organization management",
	}

	cmd.AddCommand(OrganizationNewCommand(factory))

	return cmd
}
	cmd.AddCommand(&cobra.Command{
		Use:   "registry-sessions",
		Short: "Registry session tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCfg, err := http.NewConfig()
			if err != nil {
				return err
			}
			// NewClient sends unauthenticated ping to server
			client, err := clientCfg.NewClient()
			if err != nil {
				return err
			}
			client.CreateRegistrySession(cmd.Context(), organi

			return nil
		},
	})
