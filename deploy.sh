#!/bin/bash

echo "Building version ${TAG}..."

# Build the Docker image with the Git tag as the version
docker build --build-arg VERSION=${TAG} -t ${IMAGE}:${TAG} .

# Deployment to development
export ECR_ACCOUNT_ID='733930943835'
export AWS_PROFILE='shared-dev.SharedDevAdministrators'
aws ecr get-login-password   --region ${REGION} | docker login --username AWS --password-stdin  ${ECR_ACCOUNT_ID}.dkr.ecr.eu-west-2.amazonaws.com
# deploy tag to dev
docker tag ${IMAGE}:${TAG} ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}
docker push  ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}
# deploy latest to dev
docker tag ${IMAGE}:${TAG} ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:latest
docker push  ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:latest

# Deployment to production
export ECR_ACCOUNT_ID='120161110524'
export AWS_PROFILE='shared-prod.SharedProdAdministrators'
aws ecr get-login-password   --region ${REGION} --profile $AWS_PROFILE | docker login --username AWS --password-stdin  ${ECR_ACCOUNT_ID}.dkr.ecr.eu-west-2.amazonaws.com
docker tag ${IMAGE}:${TAG} ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}
docker push  ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:${TAG}
# deploy latest to prod
docker tag ${IMAGE}:${TAG}  ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:latest
docker push  ${ECR_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}/${IMAGE}:latest
