package identities

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// AWSAuthType defines the authentication method used for AWS credential resolution.
type AWSAuthType string

const (
	// AWSAuthDefault uses the default AWS credential chain.
	AWSAuthDefault AWSAuthType = "default"
	// AWSAuthStatic uses explicitly provided access key and secret key credentials.
	AWSAuthStatic AWSAuthType = "static"
	// AWSAuthProfile uses a named profile from the shared AWS configuration.
	AWSAuthProfile AWSAuthType = "profile"
)

// AWSCredentialSource provides AWS credentials and SigV4 request signing.
type AWSCredentialSource struct {
	creds  aws.CredentialsProvider
	signer v4.Signer
	region string
}

// NewAWSCredentialSource creates an AWSCredentialSource configured with
// the given region, authentication type, and optional provider-specific options.
func NewAWSCredentialSource(
	ctx context.Context,
	region string,
	authType AWSAuthType,
	options map[string]any,
) (*AWSCredentialSource, error) {
	var creds aws.CredentialsProvider

	switch authType {
	case AWSAuthDefault, "":
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			return nil, fmt.Errorf("load default AWS config: %w", err)
		}
		creds = cfg.Credentials

	case AWSAuthStatic:
		accessKey, _ := options["access_key_id"].(string)
		secretKey, _ := options["secret_access_key"].(string)
		sessionToken, _ := options["session_token"].(string)

		if accessKey == "" || secretKey == "" {
			return nil, fmt.Errorf("access_key_id and secret_access_key are required for static auth")
		}

		creds = credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken)

	case AWSAuthProfile:
		profile, _ := options["profile"].(string)
		if profile == "" {
			return nil, fmt.Errorf("profile is required for profile auth")
		}

		cfg, err := config.LoadDefaultConfig(
			ctx,
			config.WithRegion(region),
			config.WithSharedConfigProfile(profile),
		)
		if err != nil {
			return nil, fmt.Errorf("load AWS config for profile %q: %w", profile, err)
		}
		creds = cfg.Credentials

	default:
		return nil, fmt.Errorf("unsupported AWS auth_type: %q", authType)
	}

	return &AWSCredentialSource{
		creds:  creds,
		signer: *v4.NewSigner(),
		region: region,
	}, nil
}

// SignRequest applies AWS SigV4 signing to the given HTTP request for the specified service.
func (s *AWSCredentialSource) SignRequest(
	ctx context.Context,
	req *http.Request,
	service string,
) error {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("read request body for signing: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	hash := sha256.Sum256(bodyBytes)
	payloadHash := fmt.Sprintf("%x", hash)

	cred, err := s.creds.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("retrieve AWS credentials: %w", err)
	}

	err = s.signer.SignHTTP(
		ctx,
		cred,
		req,
		payloadHash,
		service,
		s.region,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	return nil
}
