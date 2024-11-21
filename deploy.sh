#!/bin/bash

export TAG="1.5.1"
export IMAGE="ec2-restart-manager"
export REPO='platform'
export REGION='eu-west-2'
docker build -t ec2-restart-manager:${TAG} .

# deploy to production
export ECR_ACCOUNT_ID='120161110524'
aws ecr get-login-password   --region ${REGION} | docker login --username AWS --password-stdin  ${ECR_ACCOUNT_ID}.dkr.ecr.eu-west-2.amazonaws.com
docker tag ${IMAGE}:${TAG} ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}
docker push  ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}

# deploy to development
export ECR_ACCOUNT_ID='733930943835'
aws ecr get-login-password   --region ${REGION} | docker login --username AWS --password-stdin  ${ECR_ACCOUNT_ID}.dkr.ecr.eu-west-2.amazonaws.com
docker tag ${IMAGE}:${TAG} ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}
docker push  ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}