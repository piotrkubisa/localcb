package codebuild

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/docker/docker/api/types/container"
	"github.com/piotrkubisa/localci/ci"
	"github.com/pkg/errors"
)

const (
	envTag = "env"

	guestWorkingDirectory = "/tmp/src"
)

// CodeBuild mimics AWS CodeBuild runtime for the localci
type CodeBuild struct {
	Definition BuildSpec
	Project    cloudformation.AWSCodeBuildProject
	Pipeline   *ci.Pipeline
	Script     *ShellScript
}

// NewCodeBuild creates new CodeBuild pipeline runtime
func NewCodeBuild(project cloudformation.AWSCodeBuildProject) (*CodeBuild, error) {
	// Load and parse buildspec.yml file
	bs, err := ParseBuildSpec(project.Source.BuildSpec)
	if err != nil {
		return nil, errors.Wrap(err, "localci: ParseBuildSpec")
	}

	// Create a wrapper on top of the Docker client
	pipeline, err := ci.NewPipeline()
	if err != nil {
		return nil, errors.Wrap(err, "localci: ci.NewPipeline")
	}

	cb := &CodeBuild{bs, project, pipeline, NewShellScript()}
	return cb, nil
}

// WorkingDirectory returns working directory for the Docker client
func (cb *CodeBuild) WorkingDirectory(customWorkDir string) string {
	if len(customWorkDir) > 0 {
		return customWorkDir
	}

	return guestWorkingDirectory
}

// Volumes returns volumes for the Docker client
func (cb *CodeBuild) Volumes(customVolumes []string) ([]string, error) {
	if len(customVolumes) > 0 {
		return customVolumes, nil
	}

	cwd, err := filepath.Abs(cb.Project.Source.Location)
	if err != nil {
		return nil, err
	}

	defaultVolumes := []string{
		// Working directory
		cwd + ":" + guestWorkingDirectory,
	}

	return defaultVolumes, nil
}

// EnvVariables binds all variables to one common []string slice which can be
// further passed to the Docker client.
func (cb *CodeBuild) EnvVariables(customVariables []string) []string {
	// Bind all default env variables
	vars := cb.NewDefaultVariables().KeyValues()

	// Bind all env variables from the buildspec.yml definition file
	for k, v := range cb.Definition.Env.Variables {
		vars = append(vars, k+"="+v)
	}

	// Bind all env variables given by the user via --env flags
	if len(customVariables) > 0 {
		vars = append(vars, customVariables...)
	}

	return vars
}

// NewDefaultVariables parses all variables provided by default by the CodeBuild
// and returns them formatted for the Docker client.
// See: https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html
func (cb *CodeBuild) NewDefaultVariables() DefaultVariables {
	var (
		awsRegion    = "local"
		awsAccountID = "000000000"
		requestID    = "00000000-0000-0000-0000-00000"
		pipelineName = "localci-pipeline"
		kmsKeyID     = "notExistingID"
		commitID     = "ffffffff"
	)

	dv := DefaultVariables{
		AwsDefaultRegion: awsRegion,
		AwsRegion:        awsRegion,

		CodeBuildBuildArn: fmt.Sprintf("arn:aws:codebuild:%s:%s:build/%s:%s",
			awsRegion,
			awsAccountID,
			cb.Project.Name,
			requestID,
		),
		CodeBuildBuildID:         fmt.Sprintf("%s:%s", cb.Project.Name, requestID),
		CodeBuildBuildImage:      cb.Project.Environment.Image,
		CodeBuildBuildSucceeding: "1",
		CodeBuildInitiator:       fmt.Sprintf("codepipeline:%s", pipelineName),
		CodeBuildKMSKeyID: fmt.Sprintf("arn:aws:kms:%s:%s:key/%s",
			awsRegion,
			awsAccountID,
			kmsKeyID,
		),
		CodeBuildResolvedSourceVersion: commitID,
		CodeBuildSourceVersion:         commitID,
		CodeBuildSourceRepoURL:         "s3://bucket_name/input_artifact.zip",
		CodeBuildSrcDir:                guestWorkingDirectory,

		Home: "/root",
	}

	return dv
}

// DefaultVariables contains default variables provided by the CodeBuild
// See: https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html
type DefaultVariables struct {
	AwsDefaultRegion               string `env:"AWS_DEFAULT_REGION"`
	AwsRegion                      string `env:"AWS_REGION"`
	CodeBuildBuildArn              string `env:"CODEBUILD_BUILD_ARN"`
	CodeBuildBuildID               string `env:"CODEBUILD_BUILD_ID"`
	CodeBuildBuildImage            string `env:"CODEBUILD_BUILD_IMAGE"`
	CodeBuildBuildSucceeding       string `env:"CODEBUILD_BUILD_SUCCEEDING"`
	CodeBuildInitiator             string `env:"CODEBUILD_INITIATOR"`
	CodeBuildKMSKeyID              string `env:"CODEBUILD_KMS_KEY_ID"`
	CodeBuildResolvedSourceVersion string `env:"CODEBUILD_RESOLVED_SOURCE_VERSION"`
	CodeBuildSourceRepoURL         string `env:"CODEBUILD_SOURCE_REPO_URL"`
	CodeBuildSourceVersion         string `env:"CODEBUILD_SOURCE_VERSION"`
	CodeBuildSrcDir                string `env:"CODEBUILD_SRC_DIR"`
	Home                           string `env:"HOME"`
}

// KeyValues returns a list of environment variables for the Docker client in
// key=value format.
func (dv DefaultVariables) KeyValues() []string {
	varlist := []string{}
	t := reflect.TypeOf(dv)
	v := reflect.ValueOf(dv)
	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)

		entry := strings.Builder{}
		entry.WriteString(f.Tag.Get(envTag))
		entry.WriteString("=")
		entry.WriteString(v.Field(i).String())

		varlist = append(varlist, entry.String())
	}

	return varlist
}

// PhasesAsStages converts all codebuild.Phases to ci.Stages
func (cb *CodeBuild) PhasesAsStages() {
	cb.PhaseToStage(cb.Definition.Phases.Install, "install")
	cb.PhaseToStage(cb.Definition.Phases.PreBuild, "pre_build")
	cb.PhaseToStage(cb.Definition.Phases.Build, "build")
	cb.PhaseToStage(cb.Definition.Phases.PostBuild, "post_build")
}

// PhaseToStage parses codebuild.Phase as a ci.Stage
func (cb *CodeBuild) PhaseToStage(p Phase, name string) {
	commands := []ci.Command{}
	for _, cmd := range p.Commands {
		commands = append(commands, ci.NewCommand(cmd))
	}

	stage := ci.NewStage(name, commands)
	cb.Pipeline.AddStage(stage)
}

// StagesAsScript saves a localci.sh shell script into given basedir
func (cb *CodeBuild) StagesAsScript(baseDir, scriptFile string) error {
	for _, stage := range cb.Pipeline.Stages {
		cb.Script.ExtractStage(stage)
	}
	err := cb.SaveScript(baseDir + scriptFile)
	if err != nil {
		return errors.Wrap(err, "localci: cb.StagesAsScript")
	}
	return nil
}

// SaveScript creates script (i.e. shell) file on host, which can be futher used
// by the localci during the runtime.
func (cb *CodeBuild) SaveScript(location string) error {
	f, err := os.Create(location)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, cb.Script.Buffer)
	if err != nil {
		return err
	}

	return nil
}

type RunConfiguration struct {
	LogFile          string
	EnvVariables     []string
	WorkingDirectory string
	Volume           []string
	ForcePullImage   bool
	NetworkName      string

	ContainerName string
}

func (cb *CodeBuild) Validate(cfg RunConfiguration) error {
	if len(cb.Project.Environment.Image) == 0 {
		return errors.New("Please specify value for --image flag")
	}

	return nil
}

// RunInContainer starts Docker container and executes localci.sh shell script
func (cb *CodeBuild) RunInContainer(cfg RunConfiguration) error {
	_, err := cb.Pipeline.DockerVersion()
	if err != nil {
		log.Printf("localci requires Docker. Do you have docker installed and running as a service on your machine?")
		return errors.Wrap(err, "localci: cb.Pipeline.DockerVersion")
	}

	if err := cb.Pipeline.PullImage(cb.Project.Environment.Image, cfg.ForcePullImage); err != nil {
		return errors.Wrap(err, "localci: cb.Pipeline.PullImage")
	}

	cont, err := cb.CreateContainer(cfg)
	if err != nil {
		return errors.Wrap(err, "localci: cb.Pipeline.CreateContainer")
	}

	if cfg.NetworkName != "" {
		err = cb.Pipeline.NetworkConnect(cont.ID, cfg.NetworkName)
		if err != nil {
			return errors.Wrap(err, "localci: cb.Pipeline.NetworkConnect")
		}
	}

	err = cb.Pipeline.ContainerStart(cont.ID)
	if err != nil {
		return errors.Wrap(err, "localci: cb.Pipeline.ContainerStart")
	}

	stdout, stderr, err := cb.Pipeline.ContainerAttach(cont.ID)
	if err != nil {
		return errors.Wrap(err, "localci: cb.Pipeline.ContainerAttach")
	}

	cb.Pipeline.InterruptHandler(stdout, stderr, cont.ID)
	cb.Pipeline.LogWatch(stdout, stderr, cont.ID, cfg.LogFile)

	defer cb.Pipeline.CleanUp(cont.ID)

	return nil
}

// CreateContainer creates a Docker container with given configuration
func (cb *CodeBuild) CreateContainer(cfg RunConfiguration) (container.ContainerCreateCreatedBody, error) {
	config := &container.Config{
		Image:      cb.Project.Environment.Image,
		Tty:        false,
		Env:        cfg.EnvVariables,
		WorkingDir: cfg.WorkingDirectory,
		Cmd:        []string{"sh", "./localci.sh"},
	}
	host := &container.HostConfig{
		Binds:      cfg.Volume,
		Privileged: cb.Project.Environment.PrivilegedMode,
	}

	container, err := cb.Pipeline.ContainerCreate(cfg.ContainerName, config, host)

	return container, err
}
