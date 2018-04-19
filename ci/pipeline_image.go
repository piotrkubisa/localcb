package ci

import (
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	colorable "github.com/mattn/go-colorable"
)

// PullImage tries to pull an image from the image registry (local, then remote)
// This code is based on implementation found in awslabs/aws-sam-local repo
func (p *Pipeline) PullImage(imageName string, force bool) error {
	images, err := p.lookForExistingImage(imageName)
	if err != nil {
		return err
	}

	pullImage := force

	if len(images) == 0 {
		log.Printf("Cannot find container image locally, trying to pull it\n")
		pullImage = true
	}

	if pullImage {
		return p.fetchImage(imageName, images)
	}

	return nil
}

// This code is based on implementation found in awslabs/aws-sam-local repo
func (p *Pipeline) lookForExistingImage(imageName string) ([]types.ImageSummary, error) {
	// Check if we have the required Docker image for this runtime
	filter := filters.NewArgs()
	filter.Add("reference", imageName)
	images, err := p.Client.ImageList(p.Context, types.ImageListOptions{
		Filters: filter,
	})

	return images, err
}

// This code is based on implementation found in awslabs/aws-sam-local repo
func (p *Pipeline) fetchImage(imageName string, images []types.ImageSummary) error {
	log.Printf("Fetching %s...\n", imageName)

	progress, err := p.Client.ImagePull(p.Context, imageName, types.ImagePullOptions{})

	if len(images) < 0 && err != nil {
		log.Fatalf("Could not fetch %s Docker image\n%s", imageName, err)
		return err
	}

	if err != nil {
		log.Printf("Could not fetch %s Docker image: %s\n", imageName, err)
		return err
	}

	origTerm := os.Getenv("TERM")
	os.Setenv("TERM", "NOT_EXISTING")
	defer os.Setenv("TERM", origTerm)

	// Show the Docker pull messages in green
	color.Output = colorable.NewColorableStderr()
	color.Set(color.FgGreen)
	defer color.Unset()

	jsonmessage.DisplayJSONMessagesStream(progress, os.Stderr, os.Stderr.Fd(), term.IsTerminal(os.Stderr.Fd()), nil)

	return nil
}
