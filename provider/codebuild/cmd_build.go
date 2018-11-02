package codebuild

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/piotrkubisa/localcb/cmd"
	"github.com/urfave/cli"
)

var buildFlags = struct {
	DockerImage cmd.FlagPair
}{
	DockerImage: cmd.NewFlagPair("image", "i"),
}

var buildTemplate = `cd /tmp
git clone https://github.com/aws/aws-codebuild-docker-images.git
cd aws-codebuild-docker-images
cd {{.ImagePath}}
docker build -t {{.DockerImage}} .
`

// BuildCommand registers a cli.Command
func BuildCommand() cli.Command {
	return cli.Command{
		Name:      "build",
		Usage:     "Prepares shell script to build one of the AWS CodeBuild curated Docker image",
		ArgsUsage: "<docker-image>",
		Action:    buildCommand,
	}
}

func buildCommand(c *cli.Context) error {
	image := c.Args().Get(0)
	if image == "" {
		return fmt.Errorf("Provide name of the docker image, i.e. aws/codebuild/golang:1.10 or golang:1.10")
	}

	var imagePath string
	var dockerImage string
	shortImage := strings.TrimPrefix(image, "aws/codebuild/")

	switch image {
	case "android-java-8:24.4.1":
		fallthrough
	case "android-java-8:26.1.1":
		fallthrough
	case "docker:17.09.0":
		fallthrough
	case "dot-net:core-1":
		fallthrough
	case "dot-net:core-2.1":
		fallthrough
	case "dot-net:core-2":
		fallthrough
	case "golang:1.10":
		fallthrough
	case "java:openjdk-8":
		fallthrough
	case "java:openjdk-9":
		fallthrough
	case "nodejs:10.1.0":
		fallthrough
	case "nodejs:6.3.1":
		fallthrough
	case "nodejs:8.11.0":
		fallthrough
	case "php:5.6":
		fallthrough
	case "php:7.0":
		fallthrough
	case "python:2.7.12":
		fallthrough
	case "python:3.3.6":
		fallthrough
	case "python:3.4.5":
		fallthrough
	case "python:3.5.2":
		fallthrough
	case "python:3.6.5":
		fallthrough
	case "ruby:2.2.5":
		fallthrough
	case "ruby:2.3.1":
		fallthrough
	case "ruby:2.5.1":
		fallthrough
	case "ubuntu-base:14.04":
		imagePath = "ubuntu/" + strings.Replace(shortImage, ":", "/", 1)
		dockerImage = "aws/codebuild/" + shortImage
	default:
		return fmt.Errorf("Unknown image name (%s), you might need to build it manually", image)
	}

	tpl, err := template.New("script").Parse(buildTemplate)
	if err != nil {
		return err
	}

	data := struct {
		ImagePath   string
		DockerImage string
	}{imagePath, dockerImage}
	return tpl.Execute(os.Stdout, data)
}
