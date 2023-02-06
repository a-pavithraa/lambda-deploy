package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/a-pavithraa/lambda-deploy/common"
	"github.com/a-pavithraa/lambda-deploy/iam"
	"github.com/a-pavithraa/lambda-deploy/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"log"
	"os"
	"time"
)

func main() {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "yaml config file name",
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Value:   "",
				Usage:   "Name of the Lambda function",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "policy",
				Aliases: []string{"p"},
				Value:   "",
				Usage:   "Execution policy of Lambda",
			},
		),
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:    "autogenerate_execution_policy",
				Aliases: []string{"dep"},
				Value:   false,
				Usage:   "Autogenerate default execution policy",
			},
		),

		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "runtime",
				Aliases: []string{"rt"},
				Value:   "",
				Usage:   "Runtime of the Lambda function",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "handler_name",
				Aliases: []string{"hn"},
				Value:   "",
				Usage:   "Name of the handler",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "s3_bucket",
				Aliases: []string{"s3"},
				Value:   "",
				Usage:   "Name of the S3 bucket ",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "s3_key",
				Aliases: []string{"key"},
				Value:   "",
				Usage:   "S3 Key",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "zip_file",
				Aliases: []string{"zip"},
				Value:   "",
				Usage:   "Name of the Zip File",
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Name:    "memory",
				Aliases: []string{"mem"},
				Value:   128,
				Usage:   "Memory of the Lambda function",
			},
		),
		altsrc.NewIntFlag(
			&cli.IntFlag{
				Name:    "time_out",
				Aliases: []string{"to"},
				Value:   60,
				Usage:   "Timeout of the Lambda function",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "environment_variables",
				Aliases: []string{"ev"},
				Value:   "",
				Usage:   "Environment variables of the Lambda function",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "region",
				Aliases: []string{"r"},
				Value:   "",
				Usage:   "Region",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "role_arn",
				Aliases: []string{"ra"},
				Value:   "",
				Usage:   "Role ARN",
			},
		),
	}
	commands := []*cli.Command{
		{
			Name:    "upsert_lambda",
			Before:  altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config")),
			Aliases: []string{"ul"},
			Flags:   flags,
			Usage:   "Creates or Updates a Lambda",

			Action: UpsertLambda,
		},

		{
			Name:    "delete_lambda",
			Aliases: []string{"dl"},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "name",
					Usage: "Name of the Lambda function",
				},
				&cli.StringFlag{
					Name:        "delete_role",
					Usage:       "Flag to indicate whether role and policy should also be deleted - Possible values Y and N",
					DefaultText: "N",
				},
			},
			Usage: "Deletes a Lambda",

			Action: DeleteLambda,
		},
	}

	app := &cli.App{
		Commands: commands,
	}

	if err := app.Run(os.Args); err != nil {

		log.Fatalf("Not able to run the command . The reason is %s", err.Error())
	}
}

func UpsertLambda(cCtx *cli.Context) error {

	lambdaParams, err2 := SetLambdaParams(cCtx)
	if err2 != nil {
		return err2
	}

	//fmt.Println(lambdaParams)

	iamWrapper := iam.ServiceWrapper{
		Client: iam.Client(context.Background()),
	}
	lambdaWrapper := lambda.ServiceWrapper{
		Client: lambda.Client(context.Background()),
	}

	functionDetails, err := lambdaWrapper.GetFunctionDetails(context.Background(), lambdaParams.FunctionName)
	if err != nil {
		log.Println(err)
		return err

	}

	if functionDetails == nil {
		err = lambda.ValidateInputParams(*lambdaParams, true)
		if err != nil {
			log.Println(err)
			return err
		}
		_, err := lambdaWrapper.New(context.Background(), *lambdaParams, iamWrapper)
		if err != nil {
			log.Println(err)
			return err
		}

	} else {
		err = lambda.ValidateInputParams(*lambdaParams, false)
		if err != nil {
			log.Println(err)
			return err
		}
		err := lambdaWrapper.UpdateFunction(context.Background(), *lambdaParams)
		if err != nil {
			log.Println(err)
			return err
		}
		err = UpdateFunctionConfiguration(lambdaWrapper, lambdaParams)

		return err

	}

	return nil

}

func UpdateFunctionConfiguration(lambdaWrapper lambda.ServiceWrapper, lambdaParams *common.DeployParams) error {
	// Not able to perform 2 updates in succession immediately . So retrying till it is successful
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for {
		err := lambdaWrapper.UpdateFunctionConfiguration(ctx, *lambdaParams)

		if err != nil {

			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				switch apiErr.(type) {
				case *types.ResourceConflictException:
					log.Println("Resource Conflict Exception. Not able to update")
					time.Sleep(2 * time.Second)

				default:
					break

				}
			}
		} else {
			log.Println("Resource Updated successfully")
			break
		}
	}
	return err
}

func SetLambdaParams(cCtx *cli.Context) (*common.DeployParams, error) {
	lambdaParams := common.DeployParams{
		FunctionName:                cCtx.String("name"),
		Policy:                      cCtx.String("policy"),
		Runtime:                     cCtx.String("runtime"),
		BucketName:                  cCtx.String("s3_bucket"),
		KeyName:                     cCtx.String("s3_key"),
		Region:                      cCtx.String("region"),
		ZipFile:                     cCtx.String("zip_file"),
		Memory:                      cCtx.Int("memory"),
		HandlerName:                 cCtx.String("handler_name"),
		AutogenerateExecutionPolicy: cCtx.Bool("autogenerate_execution_policy"),
		Action:                      cCtx.String("action_type"),
		Timeout:                     cCtx.Int("time_out"),
		RoleArn:                     cCtx.String("role_arn"),
	}
	envVariables := cCtx.String("environment_variables")

	if common.TrimAndCheckEmptyString(&envVariables) {
		result := make(map[string]string)
		if err := json.Unmarshal([]byte(envVariables), &result); err != nil {
			log.Println(err)
			return nil, err
		}
		lambdaParams.EnvironmentVariables = result

	}
	return &lambdaParams, nil
}

func DeleteLambda(cCtx *cli.Context) error {
	name := cCtx.String("name")
	deleteRole := cCtx.String("delete_role")
	log.Println("Deleting Lambda-----", name)
	if common.TrimAndCheckEmptyString(&name) {
		return &common.InputError{
			Message: "Function Name cannot be null",
		}

	}
	lambdaWrapper := lambda.ServiceWrapper{
		Client: lambda.Client(context.Background()),
	}

	functionDetails, err := lambdaWrapper.Delete(context.Background(), name)
	if err != nil {

		return err
	}
	if common.TrimAndCheckEmptyString(&deleteRole) {
		if deleteRole == "Y" {

			wrapper := iam.ServiceWrapper{
				Client: iam.Client(context.Background()),
			}

			wrapper.DeleteRole(context.Background(), *functionDetails.Configuration.Role)

		}

	}
	return nil

}
