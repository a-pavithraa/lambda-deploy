package iam

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
)

type mockIAMClient struct {
	Client Api
}

func (m *mockIAMClient) DeleteRole(ctx context.Context, input *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return &iam.DeleteRoleOutput{}, nil
}

func (m *mockIAMClient) GetRole(ctx context.Context, input *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return &iam.GetRoleOutput{
		Role: &types.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/test"),
		},
	}, nil
}

func (m *mockIAMClient) CreatePolicy(ctx context.Context, input *iam.CreatePolicyInput, optFns ...func(*iam.Options)) (*iam.CreatePolicyOutput, error) {
	return &iam.CreatePolicyOutput{
		Policy: &types.Policy{
			Arn: aws.String("arn:aws:iam::123456789012:policy/test"),
		},
	}, nil
}

func (m *mockIAMClient) AttachRolePolicy(ctx context.Context, input *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	return &iam.AttachRolePolicyOutput{}, nil
}

func (m *mockIAMClient) CreateRole(ctx context.Context, input *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	return &iam.CreateRoleOutput{
		Role: &types.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/test"),
		},
	}, nil
}

func (m *mockIAMClient) ListAttachedRolePolicies(ctx context.Context, input *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	return &iam.ListAttachedRolePoliciesOutput{}, nil
}

func TestServiceWrapper_DeleteRole(t *testing.T) {
	sw := ServiceWrapper{
		Client: &mockIAMClient{},
	}
	err := sw.DeleteRole(context.TODO(), "test")
	assert.Nil(t, err)
}

func TestServiceWrapper_CheckRoleExists(t *testing.T) {
	sw := ServiceWrapper{
		Client: &mockIAMClient{},
	}
	arn := sw.CheckRoleExists(context.TODO(), "test")
	assert.EqualValues(t, *arn, "arn:aws:iam::123456789012:role/test")
}
func TestCreatePolicy(t *testing.T) {
	wrapper := ServiceWrapper{
		Client: &mockIAMClient{},
	}
	policyDocument := `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Action": "logs:CreateLogGroup",
			"Resource": "arn:aws:logs:us-west-2:123456789012:*"
		  }
		]
	  }`
	policy, err := wrapper.CreatePolicy(context.TODO(), policyDocument, "test_policy")
	if err != nil {
		t.Fatalf("Failed to create policy: %v", err)
	}
	if policy == nil {
		t.Fatalf("Policy is nil")
	}
}

func TestAttachRolePolicy(t *testing.T) {
	wrapper := ServiceWrapper{
		Client: &mockIAMClient{},
	}
	err := wrapper.AttachRolePolicy(context.TODO(), "test_policy_arn", "test_role")
	if err != nil {
		t.Fatalf("Failed to attach role policy: %v", err)
	}
}

func TestNewRole(t *testing.T) {
	wrapper := ServiceWrapper{
		Client: &mockIAMClient{},
	}
	trustPolicy := PolicyDocument{
		Version: "2012-10-17",
		Statement: []PolicyStatement{
			{
				Effect: "Allow",
				Principal: map[string]string{
					"Service": "lambda.amazonaws.com",
				},
				Action: []string{
					"sts:AssumeRole",
				},
			},
		},
	}
	_, err := wrapper.NewRole(context.TODO(), "Test", trustPolicy)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
}
