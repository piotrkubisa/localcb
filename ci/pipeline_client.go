package ci

func (p *Pipeline) DockerVersion() (string, error) {
	response, err := p.Client.Ping(p.Context)
	return response.APIVersion, err
}
