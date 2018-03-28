package main

import (
	"flag"
	"fmt"

	"github.com/piotrkubisa/buildspec/codebuild"
)

func main() {
	var (
		buildSpecLocation string
	)
	flag.StringVar(&buildSpecLocation, "buildspec", "./buildspec.yml", "Location to the buildspec.yml file")
	flag.Parse()

	vars := map[string]string{
		"CODEBUILD_RESOLVED_SOURCE_VERSION": "HEAD",

		"GITHUB_REPOSITORY_NAME": "buildspec",
		"SOME_VARIABLE":          "whatever",
	}

	cmds := codebuild.LoadBuildSpec(buildSpecLocation, vars)

	for _, cmd := range cmds {
		fmt.Println(cmd)
	}
}
