package lambda

import "fmt"

type DeployParams struct {
	FunctionName         string
	BucketName           string
	KeyName              string
	Region               string
	ZipFile              string
	EnvironmentVariables map[string]string
	Memory               int
	Policy               string
	Runtime              string
	HandlerName          string
	Action               string
}
type InputError struct {
	Message string
}

func (e *InputError) Error() string {
	return fmt.Sprintf(
		"Error in inputs: %s",
		e.Message)
}
