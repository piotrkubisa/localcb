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
	Version   string    `json:"version"`
	Env       Env       `json:"env"`
	Phases    Phases    `json:"phases"`
	Artifacts Artifacts `json:"artifacts"`

	// Cache is not supported
	// Cache     interface{}     `json:"cache"`
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
	Variables map[string]string `json:"variables"`

	// TODO; ParameterStore is not supported right now, so make it work :)
	// ParameterStore map[string]string `json:"parameter-store"`
}

// Phases describes contents specified in top-root `phases` key
// of the `buildspec.yml` file.
type Phases struct {
	Install   Phase `json:"install"`
	PreBuild  Phase `json:"pre_build"`
	Build     Phase `json:"build"`
	PostBuild Phase `json:"post_build"`
}

// Phase provides command list which will be run during the build.
type Phase struct {
	Commands []string `json:"commands"`
}

// Artifacts describes contents specified in top-root `artifacts` key
// of the `buildspec.yml` file.
type Artifacts struct {
	Files         interface{} `json:"files"`
	DiscardPaths  string      `json:"discard-paths"` // default=yes
	BaseDirectory string      `json:"base-directory"`
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
