package lambda

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/smithy-go"

	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/a-pavithraa/lambda-deploy/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type Api interface {
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(options *lambda.Options)) (*lambda.GetFunctionOutput, error)
}

type ServiceWrapper struct {
	Client *lambda.Client
}

func Client(ctx context.Context) *lambda.Client {
	cfg, _ := config.LoadDefaultConfig(ctx)
	lambdaClient := lambda.NewFromConfig(cfg)
	return lambdaClient

}

func (wrapper ServiceWrapper) DoesExist(ctx context.Context, name string) (bool, error) {
	_, err := wrapper.Client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: &name,
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.(type) {
			case *types.ResourceNotFoundException:
				return false, nil
			default:
				log.Fatalf("Unknown  error: %v\n", err)
			}
		}
	}

	return true, nil

}

func CheckEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func ValidateInputParamsForCreate(lambdaParams DeployParams) error {
	var errorMessage strings.Builder
	if CheckEmptyString(lambdaParams.FunctionName) {
		errorMessage.WriteString("Function Name cannot be null.\n")
	}
	checkS3 := !CheckEmptyString(lambdaParams.BucketName) && !CheckEmptyString(lambdaParams.KeyName)
	if !checkS3 && CheckEmptyString(lambdaParams.ZipFile) {
		errorMessage.WriteString("Either S3 Bucket and Key or Zip file has to be included.\n")
	}
	if CheckEmptyString(lambdaParams.Runtime) {
		errorMessage.WriteString("Runtime must be specified.\n")
	}
	if CheckEmptyString(lambdaParams.HandlerName) {
		errorMessage.WriteString("HandlerName must be specified.\n")
	}
	if len(errorMessage.String()) > 0 {
		return &InputError{
			Message: errorMessage.String(),
		}

	}
	return nil

}

func (wrapper ServiceWrapper) New(ctx context.Context, lambdaParams DeployParams) (*lambda.CreateFunctionOutput, error) {
	roleArn, _ := iam.CreateRole(ctx, lambdaParams.FunctionName, lambdaParams.Policy)
	memory := int32(lambdaParams.Memory)
	functionInput := &lambda.CreateFunctionInput{

		FunctionName: &lambdaParams.FunctionName,
		Role:         roleArn,
		Runtime:      types.Runtime(lambdaParams.Runtime),
		Handler:      &lambdaParams.HandlerName,
		MemorySize:   &memory,
	}
	if lambdaParams.EnvironmentVariables != nil {
		functionInput.Environment = &types.Environment{Variables: lambdaParams.EnvironmentVariables}
	}
	if !CheckEmptyString(lambdaParams.BucketName) && !CheckEmptyString(lambdaParams.KeyName) {
		functionInput.Code = &types.FunctionCode{
			S3Bucket: &lambdaParams.BucketName,
			S3Key:    &lambdaParams.KeyName,
		}

	}
	if !CheckEmptyString(lambdaParams.ZipFile) {
		zipFile, err := os.Open(lambdaParams.ZipFile)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		defer zipFile.Close()
		zipFileInfo, _ := zipFile.Stat()
		fileSize := zipFileInfo.Size()
		fileBuffer := make([]byte, fileSize)
		zipFile.Read(fileBuffer)
		functionInput.Code = &types.FunctionCode{
			ZipFile: fileBuffer,
		}
	}
	output, err := wrapper.Client.CreateFunction(ctx, functionInput)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(output)
	return output, nil
}
