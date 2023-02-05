package lambda

import (
	"context"
	"errors"
	"fmt"
	"github.com/a-pavithraa/lambda-deploy/common"
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

func (wrapper ServiceWrapper) GetFunctionDetails(ctx context.Context, name string) (*lambda.GetFunctionOutput, error) {
	resp, err := wrapper.Client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: &name,
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.(type) {
			case *types.ResourceNotFoundException:
				return nil, nil
			default:
				log.Fatalf("Unknown  error: %v\n", err)
				return nil, err
			}
		}
	}

	return resp, nil

}

func ValidateInputParams(lambdaParams DeployParams, createFlag bool) error {
	var errorMessage strings.Builder
	if common.TrimAndCheckEmptyString(&lambdaParams.FunctionName) {
		errorMessage.WriteString("Function Name cannot be null.\n")
	}
	checkS3 := !common.TrimAndCheckEmptyString(&lambdaParams.BucketName) && !common.TrimAndCheckEmptyString(&lambdaParams.KeyName)
	if !checkS3 && common.TrimAndCheckEmptyString(&lambdaParams.ZipFile) {
		errorMessage.WriteString("Either S3 Bucket and Key or Zip file has to be included.\n")
	}
	if createFlag {
		if common.TrimAndCheckEmptyString(&lambdaParams.Runtime) {
			errorMessage.WriteString("Runtime must be specified.\n")
		}
		if common.TrimAndCheckEmptyString(&lambdaParams.HandlerName) {
			errorMessage.WriteString("HandlerName must be specified.\n")
		}
	}

	if len(errorMessage.String()) > 0 {
		return &common.InputError{
			Message: errorMessage.String(),
		}

	}
	return nil

}
func (wrapper ServiceWrapper) UpdateFunction(ctx context.Context, lambdaParams DeployParams) error {
	functionInput := &lambda.UpdateFunctionCodeInput{
		FunctionName: &lambdaParams.FunctionName,
	}

	if !common.TrimAndCheckEmptyString(&lambdaParams.BucketName) && !common.TrimAndCheckEmptyString(&lambdaParams.KeyName) {
		functionInput.S3Bucket = &lambdaParams.BucketName
		functionInput.S3Key = &lambdaParams.KeyName

	}
	if !common.TrimAndCheckEmptyString(&lambdaParams.ZipFile) {

		contents, err := GetFunctionCodeFromZip(lambdaParams.ZipFile)
		if err != nil {
			log.Println(err)
			return err
		}
		functionInput.ZipFile = contents
	}
	wrapper.Client.UpdateFunctionCode(ctx, functionInput)
	return nil

}

func GetFunctionCodeFromZip(fileName string) ([]byte, error) {

	zipFile, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer zipFile.Close()
	zipFileInfo, _ := zipFile.Stat()
	fileSize := zipFileInfo.Size()
	fileBuffer := make([]byte, fileSize)
	zipFile.Read(fileBuffer)

	return fileBuffer, nil

}

func (wrapper ServiceWrapper) New(ctx context.Context, lambdaParams DeployParams, iamWrapper iam.ServiceWrapper) (*lambda.CreateFunctionOutput, error) {

	roleArn, _ := iamWrapper.CreateRole(ctx, lambdaParams)
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
	if !common.TrimAndCheckEmptyString(&lambdaParams.BucketName) && !common.TrimAndCheckEmptyString(&lambdaParams.KeyName) {
		functionInput.Code = &types.FunctionCode{
			S3Bucket: &lambdaParams.BucketName,
			S3Key:    &lambdaParams.KeyName,
		}

	}
	if !common.TrimAndCheckEmptyString(&lambdaParams.ZipFile) {

		contents, err := GetFunctionCodeFromZip(lambdaParams.ZipFile)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		functionInput.Code = &types.FunctionCode{
			ZipFile: contents,
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
func (wrapper ServiceWrapper) Delete(ctx context.Context, name string) (*lambda.GetFunctionOutput, error) {
	functionDetails, err := wrapper.Client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: &name,
	})
	if err != nil {
		log.Printf("Not able to delete the function. The reason is %s", err.Error())
		return nil, err
	}

	_, err = wrapper.Client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: &name,
	})

	if err != nil {
		log.Printf("Not able to delete the function. The reason is %s", err.Error())
		return nil, err
	}
	log.Println(functionDetails)

	return functionDetails, nil
}
