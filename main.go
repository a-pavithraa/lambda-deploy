package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/a-pavithraa/lambda-deploy/lambda"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
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
			Name:    "create_lambda",
			Before:  altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config")),
			Aliases: []string{"cl"},
			Flags:   flags,
			Usage:   "Creates a Lambda",

			Action: func(cCtx *cli.Context) error {
				lambdaParams := lambda.DeployParams{
					FunctionName: cCtx.String("name"),
					Policy:       cCtx.String("policy"),
					Runtime:      cCtx.String("runtime"),
					BucketName:   cCtx.String("s3_bucket"),
					KeyName:      cCtx.String("s3_key"),
					Region:       cCtx.String("region"),
					ZipFile:      cCtx.String("zip_file"),
					Memory:       cCtx.Int("memory"),
					HandlerName:  cCtx.String("handler_name"),

					Action: cCtx.String("action_type"),
				}
				str := strings.TrimSpace(cCtx.String("environment_variables"))

				if len(str) > 0 {
					result := make(map[string]string)
					if err := json.Unmarshal([]byte(str), &result); err != nil {
						fmt.Println(err)
						return err
					}
					lambdaParams.EnvironmentVariables = result

				}

				//fmt.Println(lambdaParams)
				lambdaWrapper := lambda.ServiceWrapper{
					Client: lambda.Client(context.Background()),
				}
				alreadyExists, err := lambdaWrapper.DoesExist(context.Background(), lambdaParams.FunctionName)
				fmt.Println("alreadyExists==", alreadyExists)
				err = lambda.ValidateInputParamsForCreate(lambdaParams)
				if err != nil {
					fmt.Println(err)
					log.Fatal(err)

				}
				if !alreadyExists {
					lambdaWrapper.New(context.Background(), lambdaParams)

				}
				if err != nil {
					fmt.Println(err)
					log.Fatal(err)
				}
				return nil
			},
		},
	}

	app := &cli.App{
		Commands: commands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
