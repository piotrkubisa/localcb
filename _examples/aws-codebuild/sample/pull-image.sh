#!/usr/bin/env sh

git clone https://github.com/aws/aws-codebuild-docker-images.git
cd aws-codebuild-docker-images
cd ubuntu/docker/17.09.0
docker build -t aws/codebuild/docker:17.09.0 .
