package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

func execCommand(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Printf("Error creating stdout pipe for command: %v, stdout: %v", cmd, stdout)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error creating stderr pipe for command: %v, error: %v", cmd, err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting command: %v, error: %v", cmd, err)
		return
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for command: %v to finish. error: %v", cmd, err)
		return
	}
}

func main() {
	numberOfClusters := 2
	clusterPrefix := "kcp-test"
	// var wg sync.WaitGroup
	// wg.Add(numberOfClusters)
	for idx := 0; idx < numberOfClusters; idx++ {
		clusterName := clusterPrefix + "-" + strconv.Itoa(idx)
		cmd := exec.Command("kind", "create", "cluster", "--name", clusterName, "--kubeconfig", clusterName+"-kubeconfig")
		execCommand(cmd)
		// go func() {
		// 	defer wg.Done()

		// }()
	}
	// wg.Wait()
}
