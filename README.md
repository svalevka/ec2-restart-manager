## S3
```
cd .
go mod init  ec2-restart-manager
export GOSUMDB=sum.golang.org
export GOPROXY=https://proxy.golang.org,direct


go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
```