package updater

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Config defines all configuration options
type Config struct {
	Registry        string `required:"true"`
	RegistryRegion  string `required:"true" split_words:"true"`
	SecretName      string `required:"true" split_words:"true"`
	SecretNamespace string `required:"true" split_words:"true"`
}

// RegistryAuth defines the structure
// of the docker config secret
type RegistryAuth struct {
	Auths map[string]Auth `json:"auths"`
}

// Auth defines the required auth fields
// for a docker config secret
type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

// Context defines the contexts for all
// clients and configuration
type Context struct {
	Config     *Config
	Kubernetes kubernetes.Interface
	ECR        ecriface.ECRAPI
}

const (
	ecrUser            = "AWS"
	ecrEmail           = "deprecated@example.com"
	tokenErrFmt        = "Unable to generate auth token, error: %v"
	buildSecretErrFmt  = "Unable to build dockerconfig secret, error: %v"
	updateSecretErrFmt = "Unable to update dockerconfig secret, error: %v"
)

// updateSecret will update or create the dockerconfig secret
func (c *Context) updateSecret(secret *v1.Secret) error {
	_, err := c.Kubernetes.CoreV1().Secrets(c.Config.SecretNamespace).Create(
		context.Background(),
		secret,
		metav1.CreateOptions{},
	)

	if err != nil {
		// Lets try and update
		_, err := c.Kubernetes.CoreV1().Secrets(c.Config.SecretNamespace).Update(
			context.Background(),
			secret,
			metav1.UpdateOptions{},
		)

		return err
	}

	return err
}

// buildDockerSecret will return a secret of dockerconfigjson type
// based on the provided ECR auth data
func (c *Context) buildDockerSecret(auth *ecr.AuthorizationData) (*v1.Secret, error) {
	// Marshal docker config into json
	config, err := json.Marshal(
		RegistryAuth{
			Auths: map[string]Auth{
				c.Config.Registry: {
					Username: ecrUser,
					Email:    ecrEmail,
					Password: *auth.AuthorizationToken,
					Auth: base64.StdEncoding.EncodeToString([]byte(
						fmt.Sprintf("%s:%s", ecrUser, *auth.AuthorizationToken)),
					),
				},
			},
		})

	if err != nil {
		return nil, err
	}

	// Build the kubernetes secret
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Config.SecretName,
			Namespace: c.Config.SecretNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "ecr-credential-updater",
			},
		},
		Data: map[string][]byte{
			".dockerconfigjson": config,
		},
		Type: v1.SecretTypeDockerConfigJson,
	}, nil
}

// UpdateCredentials will generate a new authorization
// token and update the image pull secret
func (c *Context) UpdateCredentials() (time.Time, error) {
	next := time.Now().Add(time.Second * 120)
	response, err := c.ECR.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})

	if err != nil {
		return next, fmt.Errorf(tokenErrFmt, err)
	}

	// For legacy reasons authorization data can
	// contain multiple tokens, ensure at least 1 is returned
	if len(response.AuthorizationData) == 0 {
		return next, fmt.Errorf(tokenErrFmt, "empty authorization response")
	}

	secret, err := c.buildDockerSecret(response.AuthorizationData[0])

	if err != nil {
		return next, fmt.Errorf(buildSecretErrFmt, err)
	}

	// Create or update the secret
	err = c.updateSecret(secret)

	if err != nil {
		return next, fmt.Errorf(updateSecretErrFmt, err)
	}

	// Set the credential update for when
	// this token expires
	return *response.AuthorizationData[0].ExpiresAt, nil
}

// Start starts the updater and ensures
// credentials will be updated
func (c *Context) Start() {
	log.Infoln("Starting ecr-credential-updater")

	for {
		log.Infof("Updating ECR credentials for secret %s", c.Config.SecretName)

		// Update the credentials
		next, err := c.UpdateCredentials()

		if err != nil {
			log.Errorln(err)
		} else {
			log.Infof("Credentials updated, next update scheduled for %s", next.Format("2006-01-02T15:04:05"))
		}

		// Sleep until the token expires
		time.Sleep(next.Sub(time.Now()))
	}
}
