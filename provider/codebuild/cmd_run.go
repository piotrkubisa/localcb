package codebuild

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/piotrkubisa/localcb/cmd"
	"github.com/urfave/cli"
)

const (
	scriptFile    = "localcb.sh"
	containerName = ""
)

var runFlags = struct {
	ProjectName      cmd.FlagPair
	LogFile          cmd.FlagPair
	BaseDir          cmd.FlagPair
	DryRun           cmd.FlagPair
	BuildspecFile    cmd.FlagPair
	DockerImage      cmd.FlagPair
	Env              cmd.FlagPair
	DockerWorkingDir cmd.FlagPair
	DockerVolumes    cmd.FlagPair
	DockerNetwork    cmd.FlagPair
	ForcePullImage   cmd.FlagPair
}{
	ProjectName:      cmd.NewFlagPair("project-name", "p"),
	LogFile:          cmd.NewFlagPair("log-file", "l"),
	BaseDir:          cmd.NewFlagPair("basedir", "b"),
	DryRun:           cmd.NewFlagPair("dry-run", ""),
	BuildspecFile:    cmd.NewFlagPair("file", "f"),
	DockerImage:      cmd.NewFlagPair("image", "i"),
	Env:              cmd.NewFlagPair("env", "e"),
	DockerWorkingDir: cmd.NewFlagPair("working-directory", "w"),
	DockerVolumes:    cmd.NewFlagPair("volume", "v"),
	DockerNetwork:    cmd.NewFlagPair("network", "net"),
	ForcePullImage:   cmd.NewFlagPair("force-pull-image", "u"),
}

// RunCommand registers a cli.Command
func RunCommand() cli.Command {
	return cli.Command{
		Name:  "run",
		Usage: "Loads buildspec.yml and runs it in the Docker container",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  runFlags.ProjectName.Join(),
				Value: "codebuild-localcb-project",
				Usage: `Optional. Name of the project`,
			},
			cli.StringFlag{
				Name:  runFlags.LogFile.Join(),
				Usage: `Optional. Locaton to file where runtime logs should be saved instead of stdout/stderr.`,
			},
			cli.StringFlag{
				Name:  runFlags.BaseDir.Join(),
				Value: "./",
				Usage: `Optional. Define location where localcb should recognize as working directory.`,
			},
			cli.BoolFlag{
				Name:  runFlags.DryRun.Join(),
				Usage: "Optional. Does localcb should bail out before starting a container?",
			},
			cli.StringFlag{
				Name:  runFlags.BuildspecFile.Join(),
				Value: "buildspec.yml",
				Usage: `Optional. Location to the buildspec.yml - AWS CodeBuild definition file.`,
			},
			cli.StringFlag{
				Name:  runFlags.DockerImage.Join(),
				Usage: `Required. Name of the Docker container image which will be used to execute commands inside.`,
			},
			cli.StringSliceFlag{
				Name:  runFlags.Env.Join(),
				Usage: `Optional. Define env-variables which will be passed to docker container during its creation. Use Var=Value syntax, where Var is key and Value is value of the env variable `,
			},
			cli.StringFlag{
				Name:  runFlags.DockerWorkingDir.Join(),
				Usage: "Optional. Define a working-directory (as in docker cli) for a docker container.",
			},
			cli.StringSliceFlag{
				Name:  runFlags.DockerVolumes.Join(),
				Usage: "Optional. Define volumes (as in docker cli) for a docker container.",
			},
			cli.StringFlag{
				Name:  runFlags.DockerNetwork.Join(),
				Usage: "Optional. Define a id or name of docker network for a container. By default is uses bridge network",
			},
			cli.BoolFlag{
				Name:  runFlags.ForcePullImage.Join(),
				Usage: "Optional. Does docker client should try to pull new version of container image even if it is already in local registry (to try update it)?",
			},
		},
		Action: runCommand,
	}
}

func runCommand(c *cli.Context) error {
	// Compute a working directory for the localcb
	baseDir := filepath.Dir(runFlags.BuildspecFile.Long)
	if c.String(runFlags.BaseDir.Long) != "" {
		baseDir = c.String(runFlags.BaseDir.Long)
	}
	if strings.HasSuffix(baseDir, "/") == false {
		baseDir += "/"
	}

	// Mimic definition as is in the CloudFormation
	project := cloudformation.AWSCodeBuildProject{
		Name: c.String(runFlags.ProjectName.Long),
		Environment: &cloudformation.AWSCodeBuildProject_Environment{
			Image:          c.String(runFlags.DockerImage.Long),
			PrivilegedMode: true,
			Type:           "LINUX_CONTAINER",
			ComputeType:    "BUILD_LOCAL",

			// TODO: Not sure if it will be needed but support for the following
			// syntax: Name=%s,Value=%s would be nice to support in the future.
			// EnvironmentVariables: []cloudformation.AWSCodeBuildProject_EnvironmentVariable{},
		},
		Source: &cloudformation.AWSCodeBuildProject_Source{
			Type:      "local",
			Location:  baseDir,
			BuildSpec: c.String(runFlags.BuildspecFile.Long),
		},
		Artifacts: &cloudformation.AWSCodeBuildProject_Artifacts{
			// TODO: It would be nice to produce an artifact as a result
			Type: "NO_ARTIFACTS",
		},
	}

	cb, err := NewCodeBuild(project)
	if err != nil {
		log.Fatal(err)
	}

	cb.PhasesAsStages()

	err = cb.StagesAsScript(baseDir, scriptFile)
	if err != nil {
		log.Fatal(err)
	}

	// Bail-out if running in dry-run mode
	if c.Bool(runFlags.DryRun.Long) == true {
		return nil
	}

	volumes, err := cb.Volumes(c.StringSlice(runFlags.DockerVolumes.Long))
	if err != nil {
		log.Fatal(err)
	}

	cfg := RunConfiguration{
		LogFile:          c.String(runFlags.LogFile.Long),
		EnvVariables:     cb.EnvVariables(c.StringSlice(runFlags.Env.Long)),
		WorkingDirectory: cb.WorkingDirectory(c.String(runFlags.DockerWorkingDir.Long)),
		Volume:           volumes,
		ForcePullImage:   c.Bool(runFlags.ForcePullImage.Long),
		NetworkName:      c.String(runFlags.DockerNetwork.Long),
		ContainerName:    containerName,
	}

	err = cb.Validate(cfg)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return cb.RunInContainer(cfg)
}
