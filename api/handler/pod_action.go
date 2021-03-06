package handler

import (
	"github.com/gridworkz/kato/worker/server/pb"
	"strings"

	"github.com/gridworkz/kato/worker/client"
	"github.com/gridworkz/kato/worker/server"
)

// PodAction is an implementation of PodHandler
type PodAction struct {
	statusCli *client.AppRuntimeSyncClient
}

// PodDetail -
func (p *PodAction) PodDetail(serviceID, podName string) (*pb.PodDetail, error) {
	pd, err := p.statusCli.GetPodDetail(serviceID, podName)
	if err != nil {
		if strings.Contains(err.Error(), server.ErrPodNotFound.Error()) {
			return nil, server.ErrPodNotFound
		}
		return nil, err
	}
	return pd, nil
}
