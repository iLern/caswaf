package service

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io/ioutil"
	"log"
)

func startModSecurityContainer(ctx context.Context, cli *client.Client) (string, error) {
	// Pull the ModSecurity image
	fmt.Println("Pull ModSec image")
	pullResponse, err := cli.ImagePull(ctx, "owasp/modsecurity-crs:nginx", types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	defer pullResponse.Close()
	_, err = ioutil.ReadAll(pullResponse)
	if err != nil {
		return "", err
	}

	fmt.Println("ModSecurity image pulled")

	// Create and start the container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "owasp/modsecurity-crs:nginx",
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func StartDocker() {
	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Start the ModSecurity container
	containerID, err := startModSecurityContainer(ctx, cli)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ModSecurity container started with ID: %s\n", containerID)

	// Close the Docker client connection
	defer cli.Close()
}
