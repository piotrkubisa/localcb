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

	var (
		knownImages = []string{
			"android-java-8:24.4.1",
			"android-java-8:26.1.1",
			"docker:1.12.1",
			"docker:17.09.0",
			"docker:18.09.0",
			"dot-net:core-1",
			"dot-net:core-2.1",
			"dot-net:core-2",
			"golang:1.10",
			"golang:1.11",
			"golang:1.5.4",
			"golang:1.6.3",
			"golang:1.7.3",
			"java:openjdk-11",
			"java:openjdk-6",
			"java:openjdk-7",
			"java:openjdk-8",
			"java:openjdk-9",
			"nodejs:10.1.0",
			"nodejs:10.14.1",
			"nodejs:4.3.2",
			"nodejs:4.4.7",
			"nodejs:5.12.0",
			"nodejs:6.3.1",
			"nodejs:7.0.0",
			"nodejs:8.11.0",
			"php:5.6",
			"php:7.0",
			"php:7.1",
			"python:2.7.12",
			"python:3.3.6",
			"python:3.4.5",
			"python:3.5.2",
			"python:3.6.5",
			"python:3.7.1",
			"ruby:2.1.10",
			"ruby:2.2.5",
			"ruby:2.3.1",
			"ruby:2.5.1",
			"ruby:2.5.3",
			"ubuntu-base:14.04",
		}

		imagePath   string
		dockerImage string
	)
	shortImage := strings.TrimPrefix(image, "aws/codebuild/")
	for _, i := range knownImages {
		if i == shortImage {
			imagePath = "ubuntu/Unsupported\\ Images/" + strings.Replace(shortImage, ":", "/", 1)
			dockerImage = "aws/codebuild/" + shortImage
			break
		}
	}
	if dockerImage == "" {
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

	fmt.Println("# Run following commands:")
	fmt.Println()
	return tpl.Execute(os.Stdout, data)
}
