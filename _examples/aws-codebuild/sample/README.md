# localci: AWS CodeBuild sample

This example uses `aws/codebuild/docker:17.09.0` Docker image, which can be build following commands:

```bash
cd /tmp
git clone https://github.com/aws/aws-codebuild-docker-images.git
cd aws-codebuild-docker-images
cd ubuntu/docker/17.09.0
docker build -t aws/codebuild/docker:17.09.0 .
```

And to run `localci` with a docker, please execute:

```bash
# Change directory to show --basedir as a feature
cd $GOPATH/piotrkubisa/localci

# Assuming localci binary has been built, otherwise go run might be used
# Override CODEBUILD_RESOLVED_SOURCE_VERSION variable and add extra SOME_VARIABLE variable which is has whitespace in value
localci \
    codebuild run \
        --basedir ./_examples/aws-codebuild/sample/ \
        --file ./_examples/aws-codebuild/sample/buildspec.yml \
        --image aws/codebuild/docker:17.09.0 \
        --env "SOME_VARIABLE=Hello World" \
        --env CODEBUILD_RESOLVED_SOURCE_VERSION=`git rev-list --all --max-count=1`
```
