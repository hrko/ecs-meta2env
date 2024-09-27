package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const prefix = "X_ECS_"

type TaskMetadata struct {
	Cluster     string `json:"Cluster"`
	TaskARN     string `json:"TaskARN"`
	Family      string `json:"Family"`
	Revision    string `json:"Revision"`
	ServiceName string `json:"ServiceName"`
}

type ContainerMetadata struct {
	Name         string `json:"Name"`
	DockerName   string `json:"DockerName"`
	ContainerARN string `json:"ContainerARN"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	taskMetadataURI := os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
	if taskMetadataURI == "" {
		fmt.Println("ECS_CONTAINER_METADATA_URI_V4 environment variable not set")
		os.Exit(1)
	}

	taskMetadata, err := fetchMetadataWithRetry[TaskMetadata](taskMetadataURI + "/task")
	if err != nil {
		fmt.Println("Error fetching task metadata:", err)
		os.Exit(1)
	}

	containerMetadata, err := fetchMetadataWithRetry[ContainerMetadata](taskMetadataURI)
	if err != nil {
		fmt.Println("Error fetching container metadata:", err)
		os.Exit(1)
	}

	env := os.Environ()
	env = append(env, fmt.Sprintf("%sCLUSTER=%s", prefix, taskMetadata.Cluster))
	env = append(env, fmt.Sprintf("%sTASK_ARN=%s", prefix, taskMetadata.TaskARN))
	env = append(env, fmt.Sprintf("%sFAMILY=%s", prefix, taskMetadata.Family))
	env = append(env, fmt.Sprintf("%sREVISION=%s", prefix, taskMetadata.Revision))
	env = append(env, fmt.Sprintf("%sSERVICE_NAME=%s", prefix, taskMetadata.ServiceName))
	env = append(env, fmt.Sprintf("%sCONTAINER_NAME=%s", prefix, containerMetadata.Name))
	env = append(env, fmt.Sprintf("%sCONTAINER_DOCKER_NAME=%s", prefix, containerMetadata.DockerName))
	env = append(env, fmt.Sprintf("%sCONTAINER_ARN=%s", prefix, containerMetadata.ContainerARN))

	binary, err := exec.LookPath(os.Args[1])
	if err != nil {
		fmt.Println("Error finding binary:", err)
		os.Exit(1)
	}
	execErr := syscall.Exec(binary, os.Args[1:], env)
	if execErr != nil {
		fmt.Println("Error executing command:", execErr)
		os.Exit(1)
	}
}

func fetchMetadataWithRetry[T any](url string) (*T, error) {
	const maxRetries = 3
	const retryInterval = 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		data, err := fetchMetadata[T](url)
		if err == nil {
			return data, nil
		}

		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("failed to fetch metadata after %d retries", maxRetries)
}

func fetchMetadata[T any](url string) (*T, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data T
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
