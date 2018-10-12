package codebuild

import (
	"io/ioutil"

	"github.com/sanathkr/yaml"
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

	// Cache is not supported
	// Cache     interface{}     `yaml:"cache"`
}

// ParseBuildSpec unmarshals contents of the `buildspec.yml` file to the newly-
// created BuildSpec struct.
func ParseBuildSpec(filePath string) (bs BuildSpec, err error) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return bs, err
	}

	err = yaml.Unmarshal(contents, &bs)
	if err != nil {
		return bs, err
	}

	return bs, nil
}

// Env describes contents specified in top-root `env` key of the `builspec.yml`
// file.
type Env struct {
	Variables map[string]string `yaml:"variables"`

	// TODO; ParameterStore is not supported right now, so make it work :)
	// ParameterStore map[string]string `yaml:"parameter-store"`
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
	Files         interface{} `yaml:"files"`
	DiscardPaths  string      `yaml:"discard-paths"` // default=yes
	BaseDirectory string      `yaml:"base-directory"`
}

// List returns list of files defined in the OutputArtifact
func (a Artifacts) List() ([]string, error) {
	switch a.Files.(type) {
	case string:
		// TODO: Apply glob-pattern here
		return []string{a.Files.(string)}, nil
	case []string:
		return a.Files.([]string), nil
	}

	return []string{}, nil
}
