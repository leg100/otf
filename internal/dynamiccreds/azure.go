package dynamiccreds

import (
	"context"
	"fmt"
)

const azure provider = "azure"

type azureVariablesCredentialsConfig struct {
	ClientIDFilePath  string `json:"client_id_file_path"`
	OIDCTokenFilePath string `json:"oidc_token_file_path"`
}

func configureAzure(ctx context.Context, h helper, audience string) (azureVariablesCredentialsConfig, []string, error) {
	clientID, err := h.getRunVar("client_id")
	if err != nil {
		return azureVariablesCredentialsConfig{}, nil, err
	}
	clientIDPath, err := h.writeFile("client_id", []byte(clientID))
	if err != nil {
		return azureVariablesCredentialsConfig{}, nil, err
	}
	if audience == "" {
		audience = "api://AzureADTokenExchange"
	}
	token, err := h.generateToken(ctx, audience)
	if err != nil {
		return azureVariablesCredentialsConfig{}, nil, err
	}
	tokenPath, err := h.writeFile("token", []byte(token))
	if err != nil {
		return azureVariablesCredentialsConfig{}, nil, err
	}
	cfg := azureVariablesCredentialsConfig{
		ClientIDFilePath:  clientIDPath,
		OIDCTokenFilePath: tokenPath,
	}
	envs := []string{
		fmt.Sprintf("ARM_CLIENT_ID=%s", clientID),
		fmt.Sprintf("ARM_OIDC_TOKEN=%s", token),
		"ARM_USE_OIDC=true",
	}
	return cfg, envs, nil
}
