To insert or update a lambda function, execute the command

```
 go run main.go ul --config <<name of the yml file>>
```

Sample yaml file

```
name: FunctioName
autogenerate_execution_policy: true
runtime: go1.x
s3_bucket: S3BucketName
s3_key: S3Key
memory: 256
region: us-east-1
handler_name: HandlerName
time_out: 300
environment_variables: |
   {
      "var1":"val1",
      "var2": "val2"
    }
```

When autogenerate_execution_policy is set to true ,it will generate the following execution policy 

```
{
   "Version": "2012-10-17",
   "Statement": [
      {
         "Effect": "Allow",
         "Action": ["logs:CreateLogGroup"],
         "Resource": "arn:aws:logs:us-east-1:accountId:*"
      },
      {
         "Effect": "Allow",
         "Action": [
            "logs:CreateLogStream",
            "logs:PutLogEvents"
         ],
         "Resource": ["arn:aws:logs:us-east-1:accountId:log-group:/aws/lambda/function_name:*"]
         
      }
   ]
}
```



Policy can also be passed through the yaml file. We can also pass zip file instead of S3 bucket . Make sure to replace the account id and lambda function name

```
name: PillsGoCliLambda12
policy: |
   {
       "Version": "2012-10-17",
       "Statement": [
         {
           "Effect": "Allow",
           "Action": ["logs:CreateLogGroup"],
           "Resource": "arn:aws:logs:us-east-1:account_id:*"
         },
         {
           "Effect": "Allow",
           "Action": [
             "logs:CreateLogStream",
             "logs:PutLogEvents"
           ],
           "Resource": ["arn:aws:logs:us-east-1:account_id:log-group:/aws/lambda/PillsGoCliLambda12:*"]
   
         }
       ]
   }
runtime: go1.x
zip_file: D:/test/main.zip
memory: 256
region: us-east-1
handler_name: main
time_out: 300
environment_variables: |
   {
      "var1":"val1",
      "var2": "val2"
    }

    
```

Lambda can also be deleted by using the following command

```
go run main.go dl --name=<<Name of the Lambda Function>>  
```

