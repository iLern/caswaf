package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	ContainerName = "modsecurity-container"
)

func isContainerExists(cli *client.Client, containerName string) (bool, string, error) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return false, "", err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if strings.Contains(name, containerName) {
				return true, container.ID, nil
			}
		}
	}

	return false, "", nil
}

func isContainerRunning(containerID string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return false, err
	}

	containerJSON, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return false, err
	}

	return containerJSON.State.Running, nil
}

func startModSecurityContainer(ctx context.Context, cli *client.Client) (string, error) {
	exists, id, err := isContainerExists(cli, ContainerName)
	if err != nil {
		return "", err
	}

	if exists {
		fmt.Printf("Container '%s' already exists. Reusing...\n", ContainerName)

		isRunning, err := isContainerRunning(id)
		if err != nil {
			return "", err
		}

		if !isRunning {
			err := cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
			if err != nil {
				return "", err
			}
		}

		return id, nil
	} else {
		// Pull the ModSecurity image
		_, err := cli.ImagePull(ctx, "owasp/modsecurity-crs:nginx", types.ImagePullOptions{})
		if err != nil {
			return "", err
		}

		// map container port 80 to host port 8080
		containerConfig := &container.Config{
			Image: "owasp/modsecurity-crs:nginx",
			ExposedPorts: nat.PortSet{
				"80/tcp": {}, // Expose the desired container port
			},
		}
		portBinding := nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "8088",
				},
			},
		}
		hostConfig := &container.HostConfig{
			PortBindings: portBinding,
		}

		resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, ContainerName)
		if err != nil {
			return "", err
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			return "", err
		}

		return resp.ID, nil
	}
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
