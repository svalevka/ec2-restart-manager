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

## Docker

Build image

```
export TAG="1.0"
export IMAGE="ec2-restart-manager"
docker build -t ec2-restart-manager:${TAG} .


```
