# localcb: AWS CodeBuild sample

## tl;dr

```bash
# Change directory to current directory
cd $GOPATH/piotrkubisa/localcb

# Download and build required docker image
bash _examples/aws-codebuild/sample/pull_image.sh

# Install dependencies (i.e. using glide) if not installed
glide install

# Run localcb.
# There is some additional magic to
# * --env: These two env variables will be ased to docker container.
# * --image: Specify a docker image (required flag).
# * --file: Path to the buildspec.yml (let's say you use custom name;
#                                default - buildspec.yml in --basedir directory).
# * --basefid: Path to the directory where is located source code (by default - current directory).
go run ./cmd/localcb/main.go \
  run \
    --basedir ./_examples/aws-codebuild/sample/ \
    --file ./_examples/aws-codebuild/sample/buildspec.yml \
    --image aws/codebuild/docker:17.09.0 \
    --env "SOME_VARIABLE=Hello World" \
    --env CODEBUILD_RESOLVED_SOURCE_VERSION=`git rev-list --all --max-count=1`
```

## Explained

This example uses `aws/codebuild/docker:17.09.0` Docker image, which can be build following commands:

```bash
cd /tmp
git clone https://github.com/aws/aws-codebuild-docker-images.git
cd aws-codebuild-docker-images
cd ubuntu/docker/17.09.0
docker build -t aws/codebuild/docker:17.09.0 .
```

And to run `localcb` with a docker, please execute:

```bash
# Change directory to show --basedir as a feature
cd $GOPATH/piotrkubisa/localcb

# Assuming localcb binary has been built, otherwise go run might be used
# Override CODEBUILD_RESOLVED_SOURCE_VERSION variable
# Add extra SOME_VARIABLE variable which has a whitespace in its value
localcb \
  run \
    --basedir ./_examples/aws-codebuild/sample/ \
    --file ./_examples/aws-codebuild/sample/buildspec.yml \
    --image aws/codebuild/docker:17.09.0 \
    --env "SOME_VARIABLE=Hello World" \
    --env CODEBUILD_RESOLVED_SOURCE_VERSION=`git rev-list --all --max-count=1`
```
