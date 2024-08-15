package auth

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
)

const AWS_CODEARTIFACT_REGISTRY_REGEX = `\.codeartifact.*\.amazonaws\.com`
const AWS_CODEARTIFACT_CAPTURE_REGEX = `([a-z0-9-]+)-(.+)\.d\.codeartifact\.(.+)\.amazonaws\.com`

type codeArtifactAuthenticator struct {
	domain  *string
	account *string
	client  *codeartifact.Client
}

func (c *codeArtifactAuthenticator) Auth(ctx context.Context, envKey string, envVars map[string]string) (map[string]string, error) {
	resp, err := c.client.GetAuthorizationToken(ctx, &codeartifact.GetAuthorizationTokenInput{
		Domain:      c.domain,
		DomainOwner: c.account,
	})
	if err != nil {
		return envVars, err
	}
	envVars[envKey] = *resp.AuthorizationToken
	return envVars, nil
}

func IsCodeArtifactURL(url string) bool {
	regex := regexp.MustCompile(AWS_CODEARTIFACT_REGISTRY_REGEX)
	return regex.MatchString(url)
}

func NewCodeArtifact(ctx context.Context, url string) (Authenticator, error) {
	if !IsCodeArtifactURL(url) {
		return nil, fmt.Errorf("registry URL is not a CodeArtifact URL, got: %s", url)
	}

	spec, err := parseRegistryUrl(url)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(spec.region))
	if err != nil {
		panic(err)
	}

	return &codeArtifactAuthenticator{
		domain:  &spec.domain,
		account: &spec.account,
		client:  codeartifact.NewFromConfig(cfg),
	}, nil
}

type codeArtifactSpec struct {
	domain  string
	account string
	region  string
}

func parseRegistryUrl(url string) (*codeArtifactSpec, error) {
	regex := regexp.MustCompile(AWS_CODEARTIFACT_CAPTURE_REGEX)
	matches := regex.FindStringSubmatch(url)
	if len(matches) == 0 {
		return nil, fmt.Errorf("registry URL is not a valid CodeArtifact URL, got: %s", url)
	}
	return &codeArtifactSpec{
		domain:  matches[1],
		account: matches[2],
		region:  matches[3],
	}, nil
}
