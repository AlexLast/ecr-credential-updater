package updater_test

import (
	"errors"
	"testing"
	"time"

	"github.com/alexlast/ecr-credential-updater/internal/updater"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

// mockECRClient we can use for testing
type mockECRClient struct {
	ecriface.ECRAPI
	GetAuthorizationTokenReturnValue *ecr.GetAuthorizationTokenOutput
	GetAuthorizationTokenReturnError error
}

// GetAuthorizationToken is a mocked ECR GetAuthorizationToken function
func (m *mockECRClient) GetAuthorizationToken(*ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error) {
	return m.GetAuthorizationTokenReturnValue, m.GetAuthorizationTokenReturnError
}

// GetContext returns a test context we
// can use for unit testing
func GetContext(next time.Time, token string) *updater.Context {
	return &updater.Context{
		Config: &updater.Config{
			Registry:        "test",
			RegistryRegion:  "eu-west-2",
			SecretName:      "test-creds",
			SecretNamespace: "test-ns",
		},
		Kubernetes: new(fake.Clientset),
		ECR: &mockECRClient{
			GetAuthorizationTokenReturnValue: &ecr.GetAuthorizationTokenOutput{
				AuthorizationData: []*ecr.AuthorizationData{
					{
						AuthorizationToken: &token,
						ExpiresAt:          &next,
					},
				},
			},
			GetAuthorizationTokenReturnError: nil,
		},
	}
}

// TestUpdateCredentials tests the UpdateCredentials function
func TestUpdateCredentials(t *testing.T) {
	n := time.Now()
	c := GetContext(n, "test")

	// Update the credentials
	next, err := c.UpdateCredentials()

	assert.NoError(t, err)
	assert.Equal(t, n, next)
}

// TestUpdateCredentialsEmptyResponse tests the UpdateCredentials function
// when no auth tokens are returned in the response
func TestUpdateCredentialsEmptyResponse(t *testing.T) {
	n := time.Now()
	c := GetContext(n, "test")

	c.ECR.(*mockECRClient).GetAuthorizationTokenReturnValue = &ecr.GetAuthorizationTokenOutput{}

	next, err := c.UpdateCredentials()

	assert.Error(t, err)
	assert.WithinDuration(t, n, next, (time.Second * 125))
}

// TestUpdateCredentialsECRFailure tests the UpdateCredentials function
// when an ECR related error occurs
func TestUpdateCredentialsECRFailure(t *testing.T) {
	n := time.Now()
	c := GetContext(n, "test")

	c.ECR.(*mockECRClient).GetAuthorizationTokenReturnError = errors.New("AWS Error")

	next, err := c.UpdateCredentials()

	assert.Error(t, err)
	assert.WithinDuration(t, n, next, (time.Second * 125))
}

// TestUpdateCredentialsKubernetesFailure tests the UpdateCredentials function
// when a Kubernetes related error occurs
func TestUpdateCredentialsKubernetesFailure(t *testing.T) {
	n := time.Now()
	c := GetContext(n, "test")

	c.Kubernetes.(*fake.Clientset).AddReactor("create", "secrets", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("Kubernetes Error")
	})
	c.Kubernetes.(*fake.Clientset).AddReactor("update", "secrets", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("Kubernetes Error")
	})

	next, err := c.UpdateCredentials()

	assert.Error(t, err)
	assert.WithinDuration(t, n, next, (time.Second * 125))
}
