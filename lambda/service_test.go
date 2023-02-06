package lambda

import (
	"context"
	"github.com/a-pavithraa/lambda-deploy/common"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockFunctionApi struct {
	Client FunctionApi
}

func (m *mockFunctionApi) GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(options *lambda.Options)) (*lambda.GetFunctionOutput, error) {
	return &lambda.GetFunctionOutput{}, nil
}

func (m *mockFunctionApi) UpdateFunctionCode(ctx context.Context, params *lambda.UpdateFunctionCodeInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionCodeOutput, error) {
	return &lambda.UpdateFunctionCodeOutput{
		FunctionName: params.FunctionName,
	}, nil
}
func (m *mockFunctionApi) UpdateFunctionConfiguration(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error) {
	return &lambda.UpdateFunctionConfigurationOutput{
		FunctionName: params.FunctionName,
	}, nil
}
func (m *mockFunctionApi) CreateFunction(ctx context.Context, params *lambda.CreateFunctionInput, optFns ...func(*lambda.Options)) (*lambda.CreateFunctionOutput, error) {
	return &lambda.CreateFunctionOutput{
		FunctionName: params.FunctionName,
	}, nil
}

func (m *mockFunctionApi) DeleteFunction(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error) {
	return nil, nil
}

func TestGetFunctionDetails(t *testing.T) {
	service := ServiceWrapper{Client: &mockFunctionApi{}}
	ctx := context.TODO()

	t.Run("found", func(t *testing.T) {
		resp, err := service.GetFunctionDetails(ctx, "found")
		assert.NoError(t, err)
		assert.NotNil(t, resp)

	})

}

func TestUpdateFunction(t *testing.T) {
	mock := &mockFunctionApi{}
	wrapper := ServiceWrapper{Client: mock}
	ctx := context.Background()
	functionName := "test-function"
	lambdaParams := common.DeployParams{FunctionName: functionName}
	err := wrapper.UpdateFunction(ctx, lambdaParams)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestValidateInputParamsValidInput(t *testing.T) {
	lambdaParams := common.DeployParams{
		FunctionName: "test-function",
		ZipFile:      "test.zip",
		HandlerName:  "handler",
		Runtime:      "go",
	}
	err := ValidateInputParams(lambdaParams, true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestValidateInputParamsInvalidValidInput(t *testing.T) {
	lambdaParams := common.DeployParams{

		ZipFile:     "test.zip",
		HandlerName: "handler",
		Runtime:     "go",
	}

	assert.Error(t, ValidateInputParams(lambdaParams, true))
}
