package iam

import "github.com/aws/aws-sdk-go-v2/service/iam"

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
