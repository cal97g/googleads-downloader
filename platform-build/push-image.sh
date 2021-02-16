#!/bin/bash

if [ $# -eq 0 ]
  then
    echo "No arguments supplied, usage: push-image <env>"
    exit
fi

repo=adwords-downloader-$1

AWS_ACCOUNT_ID=534129819123
DOCKER_SERVER=$AWS_ACCOUNT_ID.dkr.ecr.eu-west-1.amazonaws.com

echo "Getting ECR login."
aws ecr get-login-password --region eu-west-1 \
  | docker login --username AWS --password-stdin $DOCKER_SERVER

echo "Tagging image."
docker tag percept/adwords-downloader-platform:latest $DOCKER_SERVER/$repo:latest

echo "Pushing image to repo $repo."
docker push $DOCKER_SERVER/$repo:latest
