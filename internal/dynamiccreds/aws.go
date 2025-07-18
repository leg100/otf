package dynamiccreds

import (
	"bytes"
	"context"
	"fmt"

	"gopkg.in/ini.v1"
)

const aws provider = "aws"

type awsVariablesSharedConfigFile struct {
	SharedConfigFile string `json:"shared_config_file"`
}

func configureAWS(ctx context.Context, h helper, audience string) (awsVariablesSharedConfigFile, []string, error) {
	arn, err := h.getRunVar("role_arn")
	if err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}
	if audience == "" {
		audience = "aws.workload.identity"
	}

	// Write token to disk.
	token, err := h.generateToken(ctx, audience)
	if err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}
	tokenPath, err := h.writeFile("token", []byte(token))
	if err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}

	// Construct and write config.
	inidata := ini.Empty()
	section, err := inidata.NewSection("default")
	if err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}
	_, err = section.NewKey("web_identity_token_file", tokenPath)
	if err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}
	_, err = section.NewKey("role_arn", arn)
	if err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}
	var buf bytes.Buffer
	if _, err := inidata.WriteTo(&buf); err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}
	configPath, err := h.writeFile("config.ini", buf.Bytes())
	if err != nil {
		return awsVariablesSharedConfigFile{}, nil, err
	}

	cfg := awsVariablesSharedConfigFile{
		SharedConfigFile: configPath,
	}
	envs := []string{
		fmt.Sprintf("AWS_ROLE_ARN=%s", arn),
		fmt.Sprintf("AWS_WEB_IDENTITY_TOKEN_FILE=%s", tokenPath),
	}
	return cfg, envs, nil
}
