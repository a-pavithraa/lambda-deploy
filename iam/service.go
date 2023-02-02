package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type ServiceWrapper struct {
	IamClient *iam.Client
}
type PolicyDocument struct {
	Version   string
	Statement []PolicyStatement
}

// PolicyStatement defines a statement in a policy document.
type PolicyStatement struct {
	Effect    string
	Action    []string
	Principal map[string]string `json:",omitempty"`
	Resource  *string           `json:",omitempty"`
}

func (wrapper ServiceWrapper) validatePolicy(lambdaExecutionRolePolicy string) error {
	lambdaBasicRole := PolicyDocument{}
	if err := json.Unmarshal([]byte(lambdaExecutionRolePolicy), &lambdaBasicRole); err != nil {
		return err

	}
	return nil
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
	fmt.Println(*role.Arn)
	fmt.Println(*role.AssumeRolePolicyDocument)
	return role.Arn
}
func (wrapper ServiceWrapper) CreatePolicy(ctx context.Context, policyDocument string, policyName string) (*types.Policy, error) {
	result, err := wrapper.IamClient.CreatePolicy(context.TODO(), &iam.CreatePolicyInput{
		PolicyDocument: aws.String(policyDocument),
		PolicyName:     aws.String(policyName),
	})
	if err != nil {
		log.Fatalf("Couldn't create policy %v. Here's why: %v\n", policyName, err)

	} else {
		policy := result.Policy
		return policy, nil
	}
	return nil, err
}
func (wrapper ServiceWrapper) AttachRolePolicy(ctx context.Context, policyArn string, roleName string) error {
	_, err := wrapper.IamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	})
	if err != nil {
		log.Fatalf("Couldn't attach policy %v to role %v. Here's why: %v\n", policyArn, roleName, err)
	}
	return err
}
func (wrapper ServiceWrapper) NewRole(ctx context.Context, roleName string, trustPolicy PolicyDocument) (*types.Role, error) {
	var role *types.Role

	policyBytes, err := json.Marshal(trustPolicy)
	if err != nil {
		log.Fatalf("Couldn't create trust policy for %v. Here's why: %v\n", trustPolicy, err)
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
		log.Fatalf("Couldn't list attached policies for role %v. Here's why: %v\n", roleName, err)
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
func CreateRole(ctx context.Context, roleName string, rolePolicy string) (*string, error) {
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

	roleArn := roleWrapper.CheckRoleExists(roleName)
	if roleArn == nil {

		role, err := roleWrapper.NewRole(ctx, roleName, trustPolicy)
		if err != nil {
			log.Fatal(err)
		}

		roleArn = role.Arn
	}

	//accountId := strings.Split(*roleArn, ":")[4]
	policies, err := roleWrapper.ListAttachedRolePolicies(ctx, roleName)
	if err != nil {

		log.Fatal(err)
	}
	if len(policies) == 0 {
		err := roleWrapper.SetupPolicesAndAttachPolicy(ctx, roleName, rolePolicy)
		if err != nil {

			log.Fatal(err)
		}
	}
	return roleArn, nil

}
