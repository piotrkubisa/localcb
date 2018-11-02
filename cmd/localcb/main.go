// localcb is a program which mimics behaviour of popular CI/CD pipelines and
// thanks to docker, brings them into the local environment
package main

import (
	"log"
	"os"

	"github.com/piotrkubisa/localcb/provider/codebuild"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "localcb"
	app.Usage = "Run AWS CodeBuild pipeline directly on your local machine"
	app.Version = "0.4.1"
	app.Commands = []cli.Command{
		codebuild.BuildCommand(),
		codebuild.RunCommand(),
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
