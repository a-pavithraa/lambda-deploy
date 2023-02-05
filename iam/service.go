package iam

import (
	"context"
	"encoding/json"
	"github.com/a-pavithraa/lambda-deploy/lambda"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func (wrapper ServiceWrapper) validatePolicy(lambdaExecutionRolePolicy string) error {
	lambdaBasicRole := PolicyDocument{}
	if err := json.Unmarshal([]byte(lambdaExecutionRolePolicy), &lambdaBasicRole); err != nil {
		return err

	}
	return nil
}
func (wrapper ServiceWrapper) DeleteRole(ctx context.Context, roleName string) error {
	_, err := wrapper.IamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: &roleName,
	})
	return err

}
func (wrapper ServiceWrapper) CheckRoleExists(roleName string) *string {
	var role *types.Role
	result, err := wrapper.IamClient.GetRole(context.TODO(),
		&iam.GetRoleInput{RoleName: aws.String(roleName)})
	if err != nil {
		return nil
	} else {
		role = result.Role
	}
	log.Println(*role.Arn)
	log.Println(*role.AssumeRolePolicyDocument)
	return role.Arn
}
func (wrapper ServiceWrapper) CreatePolicy(ctx context.Context, policyDocument string, policyName string) (*types.Policy, error) {
	result, err := wrapper.IamClient.CreatePolicy(context.TODO(), &iam.CreatePolicyInput{
		PolicyDocument: aws.String(policyDocument),
		PolicyName:     aws.String(policyName),
	})
	if err != nil {
		log.Printf("Couldn't create policy %v. Here's why: %v\n", policyName, err)
		return nil, err

	} else {
		policy := result.Policy
		return policy, nil
	}

}
func (wrapper ServiceWrapper) AttachRolePolicy(ctx context.Context, policyArn string, roleName string) error {
	_, err := wrapper.IamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	})
	if err != nil {
		log.Printf("Couldn't attach policy %v to role %v. Here's why: %v\n", policyArn, roleName, err)

	}
	return err
}

func (wrapper ServiceWrapper) NewRole(ctx context.Context, roleName string, trustPolicy PolicyDocument) (*types.Role, error) {
	var role *types.Role

	policyBytes, err := json.Marshal(trustPolicy)
	if err != nil {
		log.Printf("Couldn't create trust policy for %v. Here's why: %v\n", trustPolicy, err)
		return nil, err
	}
	result, err := wrapper.IamClient.CreateRole(ctx, &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(string(policyBytes)),

		RoleName: aws.String(roleName),
	})

	if err != nil {
		log.Fatalf("Couldn't create role %v. Here's why: %v\n", roleName, err)
	} else {
		role = result.Role
	}
	return role, err
}
func (wrapper ServiceWrapper) ListAttachedRolePolicies(ctx context.Context, roleName string) ([]types.AttachedPolicy, error) {
	var policies []types.AttachedPolicy
	result, err := wrapper.IamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		log.Printf("Couldn't list attached policies for role %v. Here's why: %v\n", roleName, err)
		return nil, err
	} else {
		policies = result.AttachedPolicies
	}
	return policies, err
}
func (wrapper ServiceWrapper) SetupPolicesAndAttachPolicy(ctx context.Context, roleName string, lambdaExecutionRolePolicy string) error {

	policy, err := wrapper.CreatePolicy(ctx, strings.TrimSpace(lambdaExecutionRolePolicy), roleName+"_policy")
	if err != nil {
		return err
	}
	err = wrapper.AttachRolePolicy(ctx, *policy.Arn, roleName)
	if err != nil {
		return err
	}
	return nil
}
func Client(ctx context.Context) *iam.Client {
	cfg, _ := config.LoadDefaultConfig(ctx)
	iamClient := iam.NewFromConfig(cfg)
	return iamClient
}
func (wrapper ServiceWrapper) CreateRole(ctx context.Context, lambdaParams lambda.DeployParams) (*string, error) {
	cfg, _ := config.LoadDefaultConfig(ctx)
	roleWrapper := ServiceWrapper{
		IamClient: iam.NewFromConfig(cfg),
	}

	trustPolicy := PolicyDocument{
		Version: "2012-10-17",
		Statement: []PolicyStatement{{
			Effect:    "Allow",
			Principal: map[string]string{"Service": "lambda.amazonaws.com"},
			Action:    []string{"sts:AssumeRole"},
		}},
	}

	roleArn := roleWrapper.CheckRoleExists(lambdaParams.FunctionName)
	if roleArn == nil {

		role, err := roleWrapper.NewRole(ctx, lambdaParams.FunctionName, trustPolicy)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		roleArn = role.Arn
	}

	accountId := strings.Split(*roleArn, ":")[4]
	policies, err := roleWrapper.ListAttachedRolePolicies(ctx, lambdaParams.FunctionName)
	if err != nil {

		log.Println(err)
		return nil, err
	}
	// To overwrite or not to overwrite the existing policy - going with not to overwrite
	if len(policies) == 0 {
		if lambdaParams.AutogenerateExecutionPolicy {
			err = AutoGenerateBasicPolicy(ctx, lambdaParams.FunctionName, accountId, roleWrapper)
			if err != nil {
				return nil, err
			}
		}
		if strings.TrimSpace(lambdaParams.Policy) != "" {
			err := roleWrapper.SetupPolicesAndAttachPolicy(ctx, lambdaParams.FunctionName, lambdaParams.Policy)
			if err != nil {

				log.Println(err)
				return nil, err
			}
		}

	}
	return roleArn, nil

}

func AutoGenerateBasicPolicy(ctx context.Context, name string, accountId string, roleWrapper ServiceWrapper) error {
	lambdaExecutionRolePolicy := `
	{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": ["logs:CreateLogGroup"],
				"Resource": "arn:aws:logs:us-east-1:` + accountId + `:*"
			},
			{
				"Effect": "Allow",
				"Action": [
					"logs:CreateLogStream",
					"logs:PutLogEvents"
				],
				"Resource": ["arn:aws:logs:us-east-1:` + accountId + `:log-group:/aws/lambda/` + name + `:*"]
				
			}
		]
	}
`
	err := roleWrapper.SetupPolicesAndAttachPolicy(ctx, name, lambdaExecutionRolePolicy)
	if err != nil {

		log.Println(err)

	}
	return err
}
