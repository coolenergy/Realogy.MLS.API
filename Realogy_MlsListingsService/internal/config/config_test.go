//go:build integration

package config_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"mlslisting/internal/config"
	"os"
	"testing"

	"log"
)

// Test the mapping of aws parameters to application configurations.
// for the purpose of this test, the aws call is being simulated to return the aws parameter stores.
func TestReadAndMapAWSParameterStore(t *testing.T) {
	t.Parallel()

	// mock aws env access details
	os.Setenv("AWS_ACCESS_KEY_ID", "foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "bar")
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")

	ssmResolverFn := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == "ssm" {
			return endpoints.ResolvedEndpoint{
				URL: "http://localhost:4566",
			}, nil
		}

		return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	}

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String("us-west-2"),
		EndpointResolver: endpoints.ResolverFunc(ssmResolverFn),
	}))

	svc := ssm.New(awsSession)
	basePath := "/realogy/services/mls-listings-service"
	url := basePath + "/mls-mongodb-url"
	user := basePath + "/mls-mongodb-user"
	pass := basePath + "/mls-mongodb-pass"
	svc.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(url),
	})
	svc.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(user),
	})
	svc.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(pass),
	})

	urlparam, err := svc.PutParameter(&ssm.PutParameterInput{
		Name:  aws.String(url),
		Value: aws.String("localhost:27017"),
		Type:  aws.String("SecureString"),
	})
	log.Printf("Url Param: %s", urlparam)
	assert.Nil(t, err)

	userparam, err := svc.PutParameter(&ssm.PutParameterInput{
		Name:  aws.String(user),
		Value: aws.String("root"),
		Type:  aws.String("SecureString"),
	})
	log.Printf("User Param: %s", userparam)
	assert.Nil(t, err)

	passparam, err := svc.PutParameter(&ssm.PutParameterInput{
		Name:  aws.String(pass),
		Value: aws.String("example"),
		Type:  aws.String("SecureString"),
	})
	log.Printf("Password Param: %s", passparam)
	assert.Nil(t, err)

	v := viper.New()
	v.Set("aws.ssm.basepath", "/realogy/services/mls-listings-service/")

	config.ReadAndMapParameterStore()(v, awsSession.Config)

	assert.Equal(t, "localhost:27017", v.GetString("mongodb.url"))
	assert.Equal(t, "root", v.GetString("mongodb.user"))
	assert.Equal(t, "example", v.GetString("mongodb.pass"))
}
