package lambda

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type FunctionApi interface {
	GetFunction(ctx context.Context, params *lambda.GetFunctionInput, optFns ...func(options *lambda.Options)) (*lambda.GetFunctionOutput, error)
	UpdateFunctionCode(ctx context.Context, params *lambda.UpdateFunctionCodeInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionCodeOutput, error)
	CreateFunction(ctx context.Context, params *lambda.CreateFunctionInput, optFns ...func(*lambda.Options)) (*lambda.CreateFunctionOutput, error)
	DeleteFunction(ctx context.Context, params *lambda.DeleteFunctionInput, optFns ...func(*lambda.Options)) (*lambda.DeleteFunctionOutput, error)
	UpdateFunctionConfiguration(ctx context.Context, params *lambda.UpdateFunctionConfigurationInput, optFns ...func(*lambda.Options)) (*lambda.UpdateFunctionConfigurationOutput, error)
}
type ServiceWrapper struct {
	Client FunctionApi
}
