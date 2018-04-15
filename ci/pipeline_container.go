package ci

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// ContainerCreate creates a container
func (p *Pipeline) ContainerCreate(name string, config *container.Config, host *container.HostConfig) (container.ContainerCreateCreatedBody, error) {
	cont, err := p.Client.ContainerCreate(p.Context, config, host, nil, name)

	return cont, err
}

// NetworkConnect connects container with a specified network
func (p *Pipeline) NetworkConnect(contID, netName string) error {
	err := p.Client.NetworkConnect(p.Context, netName, contID, nil)
	if err != nil {
		return err
	}

	log.Printf("Connecting container %s to network %s", contID, netName)
	return nil
}

// ContainerStart starts the container
func (p *Pipeline) ContainerStart(contID string) error {
	err := p.Client.ContainerStart(p.Context, contID, types.ContainerStartOptions{})

	return err
}

// ContainerAttach attach to the container to read the stdout/stderr stream
func (p *Pipeline) ContainerAttach(contID string) (io.ReadCloser, io.ReadCloser, error) {
	attach, err := p.Client.ContainerAttach(p.Context, contID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  false,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return nil, nil, err
	}

	stdout, stderr := demuxDockerStream(attach.Reader)

	return stdout, stderr, nil
}

// demuxDockerStream registers separate io.Pipes per StdOut and StdErr.
// This code is based on implementation found in awslabs/aws-sam-local repo
func demuxDockerStream(input io.Reader) (io.ReadCloser, io.ReadCloser) {
	stdoutreader, stdoutwriter := io.Pipe()
	stderrreader, stderrwriter := io.Pipe()

	go func() {
		_, err := stdcopy.StdCopy(stdoutwriter, stderrwriter, input)
		if err != nil {
			log.Printf("Error reading I/O from runtime container: %s\n", err)
		}

		stdoutwriter.Write([]byte("\n"))

		stdoutwriter.Close()
		stderrwriter.Close()

	}()

	return stdoutreader, stderrreader
}

// LogWatch tails and reads logs from the container
func (p *Pipeline) LogWatch(stdoutTxt, stderrTxt io.Reader, contID, logFileLocation string) {
	var wg sync.WaitGroup

	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)

	if len(logFileLocation) > 0 {
		if logFile, err := os.Create(logFileLocation); err == nil {
			stderr = io.Writer(logFile)
			stdout = io.Writer(logFile)
			log.SetOutput(stderr)
		} else {
			log.Fatalf("Failed to open log file %s: %s\n", logFileLocation, err)
		}
	}

	wg.Add(1)
	go func() {
		io.Copy(stdout, stdoutTxt)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(stderr, stderrTxt)
		wg.Done()
	}()

	wg.Wait()
	fmt.Fprintf(stderr, "\n")
	p.CleanUp(contID)
}

// InterruptHandler register a handler of Interrupt signal (CTRL+C)
func (p *Pipeline) InterruptHandler(stdout, stderr io.ReadCloser, contID string) {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Printf("Received Interrupt signal, terminating...")
		stderr.Close()
		stdout.Close()
		p.CleanUp(contID)
		os.Exit(0)
	}()
}

// CleanUp removes the Docker container used by this runtime
func (p *Pipeline) CleanUp(contID string) {
	p.Client.ContainerKill(p.Context, contID, "SIGKILL")
	p.Client.ContainerRemove(p.Context, contID, types.ContainerRemoveOptions{})
}
