package common

type DeployParams struct {
	FunctionName                string
	BucketName                  string
	KeyName                     string
	Region                      string
	ZipFile                     string
	EnvironmentVariables        map[string]string
	Memory                      int
	Timeout                     int
	Policy                      string
	Runtime                     string
	HandlerName                 string
	Action                      string
	AutogenerateExecutionPolicy bool
	RoleArn                     string
}
