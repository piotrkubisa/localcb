package codebuild

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/go-yaml/yaml"
)

// BuildSpec defines top-root contents of `buildspec.yml` file.
//
// See:
// - https://docs.aws.amazon.com/codebuild/latest/userguide/build-spec-ref.html
type BuildSpec struct {
	Version   string    `yaml:"version"`
	Env       Env       `yaml:"env"`
	Phases    Phases    `yaml:"phases"`
	Artifacts Artifacts `yaml:"artifacts"`
}

func LoadBuildSpec(filePath string, customVars map[string]string) []string {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	bs, err := ParseBuildSpec(contents)
	if err != nil {
		log.Fatal(err)
	}

	vars := bs.ParseVariables(customVars)

	return bs.Commands(vars)
}

func ParseBuildSpec(contents []byte) (bs BuildSpec, err error) {
	err = yaml.Unmarshal(contents, &bs)
	return bs, err
}

func (bs *BuildSpec) ParseVariables(overridables map[string]string) map[string]string {
	vars := DefaultEnvVariables()

	if overridables != nil {
		for k, v := range overridables {
			vars[k] = v
		}
	}

	for k, v := range bs.Env.Variables {
		vars[k] = v
	}

	return vars
}

// Commands extracts all commands and variables as a shell-script commands.
func (bs *BuildSpec) Commands(vars map[string]string) []string {
	commands := []string{}

	if vars != nil {
		for k, v := range vars {
			commands = append(commands, fmt.Sprintf("export %s=%s", k, v))
		}
	}

	for _, c := range bs.Phases.Install.Commands {
		commands = append(commands, c)
	}
	for _, c := range bs.Phases.PreBuild.Commands {
		commands = append(commands, c)
	}
	for _, c := range bs.Phases.Build.Commands {
		commands = append(commands, c)
	}
	for _, c := range bs.Phases.PostBuild.Commands {
		commands = append(commands, c)
	}

	return commands
}

// Env describes contents specified in top-root `env` key of the `builspec.yml`
// file.
type Env struct {
	Variables map[string]string `yaml:"variables"`

	// TODO; ParameterStore is not supported right now,
	// it would be great to support it
	// ParameterStore map[string]string `yaml:"parameter-store"`
}

// DefaultEnvVariables defines environmental variables assigned by CodeBuild
// in each build.
//
// See:
// - https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-env-vars.html
func DefaultEnvVariables() map[string]string {
	return map[string]string{
		"AWS_DEFAULT_REGION": "us-east-1",
		"AWS_REGION":         "us-east-1",

		"CODEBUILD_BUILD_ARN":               "arn:aws:codebuild:region-ID:account-ID:build/codebuild-demo-project:b1e6661e-e4f2-4156-9ab9-82a19EXAMPLE)",
		"CODEBUILD_BUILD_ID":                "codebuild-demo-project:b1e6661e-e4f2-4156-9ab9-82a19EXAMPLE",
		"CODEBUILD_BUILD_IMAGE":             "aws/codebuild/java:openjdk-8",
		"CODEBUILD_BUILD_SUCCEEDING":        "1",
		"CODEBUILD_INITIATOR":               "codepipeline/my-demo-pipeline",
		"CODEBUILD_KMS_KEY_ID":              "arn:aws:kms:region-ID:account-ID:key/key-ID or alias/key-alias)",
		"CODEBUILD_RESOLVED_SOURCE_VERSION": "ffffffff",
		"CODEBUILD_SOURCE_REPO_URL":         "s3://bucket_name/input_artifact.zip",
		"CODEBUILD_SOURCE_VERSION":          "ffffffff",
		"CODEBUILD_SRC_DIR":                 "/tmp/src123456789/src",

		"HOME": "/root",
	}
}

// Phases describes contents specified in top-root `phases` key
// of the `buildspec.yml` file.
type Phases struct {
	Install   Phase `yaml:"install"`
	PreBuild  Phase `yaml:"pre_build"`
	Build     Phase `yaml:"build"`
	PostBuild Phase `yaml:"post_build"`
}

// Phase provides command list which will be run during the build.
type Phase struct {
	Commands []string `yaml:"commands"`
}

// Artifacts describes contents specified in top-root `artifacts` key
// of the `buildspec.yml` file.
type Artifacts struct {
	// TODO: Files could be an array or just a string (glob pattern).
	Files         interface{} `yaml:"files"`
	DiscardPaths  string      `yaml:"discard-paths"` // default=yes
	BaseDirectory string      `yaml:"base-directory"`
}
