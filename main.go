package main

import (
	"context"
	"encoding/json"
	"github.com/a-pavithraa/lambda-deploy/common"
	"github.com/a-pavithraa/lambda-deploy/iam"
	"github.com/a-pavithraa/lambda-deploy/lambda"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"log"
	"os"
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
				Usage:   "Name of the Lambda function",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "s3_key",
				Aliases: []string{"key"},
				Value:   "",
				Usage:   "Name of the Lambda function",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "zip_file",
				Aliases: []string{"zip"},
				Value:   "",
				Usage:   "Name of the Lambda function",
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
				Usage:   "Environment variables of the Lambda function",
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "action_type",
				Aliases: []string{"at"},
				Value:   "",
				Usage:   "Action Type",
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
	lambdaWrapper := lambda.ServiceWrapper{
		Client: lambda.Client(context.Background()),
	}
	iamWrapper := iam.ServiceWrapper{
		IamClient: iam.Client(context.Background()),
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

	}

	return nil

}

func SetLambdaParams(cCtx *cli.Context) (*lambda.DeployParams, error) {
	lambdaParams := lambda.DeployParams{
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
			iamWrapper := iam.ServiceWrapper{
				IamClient: iam.Client(context.Background()),
			}
			iamWrapper.DeleteRole(context.Background(), *functionDetails.Configuration.Role)

		}

	}
	return nil

}
