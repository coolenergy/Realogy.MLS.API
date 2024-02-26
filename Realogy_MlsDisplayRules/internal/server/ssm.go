package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"reflect"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

type Context struct {
	Environment string
}

// Struct used to provide database details
type DatabaseDetails struct {
	MongoDbHost string `ssm:"/realogy/services/{{.Environment}}/mls-display-rules/mls-mongodb-host" validate:"required"`
	MongoDbUser string `ssm:"/realogy/services/{{.Environment}}/mls-display-rules/mls-mongodb-user" validate:"required"`
	MongoDbPw   string `ssm:"/realogy/services/{{.Environment}}/mls-display-rules/mls-mongodb-pw" validate:"required"`
}

var (
	errUnhandledType = fmt.Errorf("unhandled type")
	errRequiresPTR   = fmt.Errorf("ref must be ptr type")
)

// Name of the struct tag used for parameters
const tagName = "ssm"

func LoadParametersCache(api ssmiface.SSMAPI, c Context, ref interface{}) error {
	parametersValue := reflect.ValueOf(ref)
	if parametersValue.Kind() != reflect.Ptr {
		return errRequiresPTR
	}
	parametersValue = parametersValue.Elem()
	parametersType := parametersValue.Type()
	for i := 0; i < parametersType.NumField(); i++ {
		field := parametersType.Field(i)

		tagValue := field.Tag.Get(tagName)
		if tagValue == "" {
			continue
		}

		parameterName, err := interpolate(c, tagValue)
		if err != nil {
			return fmt.Errorf("unable to load parameters: %w", err)
		}

		v, err := ssmLookup(api, parameterName)
		if err != nil {
			var ae awserr.Error
			if ok := errors.As(err, &ae); ok && ae.Code() == ssm.ErrCodeParameterNotFound {
				return fmt.Errorf("required parameter not found: %v", parameterName)
			}
			return fmt.Errorf("failed to lookup parameters: %w", err)
		}
		cleanV := strings.TrimSpace(v)
		parametersValue.Field(i).SetString(strings.TrimSpace(cleanV))
	}
	return nil
}

// interpolate returns the result of evaluating the provided text template
func interpolate(c Context, text string) (string, error) {
	t, err := template.New("text").Parse(text)
	if err != nil {
		return "", fmt.Errorf("unable to load parameters: unable to parse tag, %v: %w", text, err)
	}

	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, c)
	if err != nil {
		return "", fmt.Errorf("unable to load parameters: unable to execute template, %v: %w", text, err)
	}

	return buf.String(), nil
}

func ssmLookup(api ssmiface.SSMAPI, name string) (string, error) {
	input := ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	}

	output, err := api.GetParameter(&input)
	if err != nil {
		return "", fmt.Errorf("unable to fetch parameter, %v: %w", name, err)
	}
	return aws.StringValue(output.Parameter.Value), nil
}
