// localci is a program which mimics behaviour of popular CI/CD pipelines and
// thanks to docker, brings them into the local environment
package main

import (
	"log"
	"os"

	"github.com/piotrkubisa/localci/provider/codebuild"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "localci"
	app.Usage = "Run CI pipeline directly on your local machine"
	app.Version = "1.0.0"
	app.Commands = []cli.Command{
		{
			Name:  "codebuild",
			Usage: "Use buildspec.yml definition (AWS CodeBuild).",
			Subcommands: []cli.Command{
				codebuild.RunCommand(),
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
