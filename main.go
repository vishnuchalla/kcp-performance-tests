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
		fmt.Printf("Error creating stdout pipe for command: %v, stdout: %v\n", cmd, stdout)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("Error creating stderr pipe for command: %v, error: %v\n", cmd, err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting command: %v, error: %v\n", cmd, err)
		return
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for command: %v to finish. error: %v\n", cmd, err)
		return
	}
}

func execCommandBackground(cmd *exec.Cmd) {
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v, error: %v\n", cmd, err)
		return
	}
	fmt.Printf("Command: %v started in background\n", cmd)
}

func createKindCluster(clusterName string) {
	cmd := exec.Command("kind", "create", "cluster", "--name", clusterName, "--kubeconfig", clusterName+"-kubeconfig")
	execCommand(cmd)
}

func createOrg(kcpPath string, orgName string) {
	fmt.Printf("Creating Organization Workspace: %v\n", orgName)
	cmd := exec.Command(kcpPath+"/bin/kubectl-kcp", "ws", "create", orgName, "--type", "Organization", "--enter")
	execCommand(cmd)
}

func createWs(kcpPath string, wsName string) {
	fmt.Printf("Creating Universal Workspace: %v\n", wsName)
	cmd := exec.Command(kcpPath+"/bin/kubectl-kcp", "ws", "create", wsName)
	execCommand(cmd)
}

func setEnv(variable string, value string) {
	err := os.Setenv(variable, value)
	if err != nil {
		fmt.Printf("Error setting environment variable: %v, error: %v\n", variable, err)
		return
	}
	fmt.Printf("Env variable %v=%v set successfully\n", variable, os.Getenv(variable))
}

func initializeKcp(kcpPath string) {
	// fmt.Println("Initializing KCP")
	// kcpCmd := exec.Command(kcpPath+"/bin/kcp", "start", "--authorization-always-allow-paths=/metrics,/readyz,/livez,/healthz")
	// execCommandBackground(kcpCmd)
	// time.Sleep(20 * time.Second)
	setEnv("KUBECONFIG", kcpPath+"/.kcp/admin.kubeconfig")
	rootCmd := exec.Command(kcpPath+"/bin/kubectl-kcp", "ws", "use", "root")
	execCommand(rootCmd)
}

func setupSyncer(kcpPath string, orgName string, wsName string, syncerImage string) {
	currentWs := "root:" + orgName + ":" + wsName
	fmt.Printf("Swithching to the workspace: %v\n", currentWs)
	cmd := exec.Command(kcpPath+"/bin/kubectl-kcp", "ws", "use", currentWs)
	execCommand(cmd)
	syncerCmd := exec.Command(kcpPath+"/bin/kubectl-kcp", "workload", "sync", wsName, "--syncer-image", syncerImage, "--output-file", wsName+"-syncer.yaml")
	execCommand(syncerCmd)
	deploySyncer := exec.Command("kubectl", "--kubeconfig", wsName+"-kubeconfig", "apply", "-f", wsName+"-syncer.yaml")
	execCommand(deploySyncer)
	waitForSyncer := exec.Command("kubectl", "wait", "--for=condition=Ready", "synctarget/"+wsName)
	execCommand(waitForSyncer)
	bindCompute := exec.Command(kcpPath+"/bin/kubectl-kcp", "bind", "compute", currentWs)
	execCommand(bindCompute)
	fmt.Printf("Syncer successfully deployed on cluster:%v\n", wsName)
}

func main() {
	// defer func() {
	// 	cmd := exec.Command("rm", "-rf", kcpPath+"/.kcp")
	// 	execCommand(cmd)
	// }()
	numberOfClusters := 2
	clusterPrefix := "kcp-test"
	orgName := "perf-scale-test"
	kcpPath := "/Users/vishnuchalla/kcp"
	syncerImage := "ghcr.io/kcp-dev/kcp/syncer:v0.11.0"
	// var wg sync.WaitGroup
	// wg.Add(numberOfClusters)
	initializeKcp(kcpPath)
	createOrg(kcpPath, orgName)

	fmt.Printf("Bringing up clusters and their namespaces\n")
	for idx := 0; idx < numberOfClusters; idx++ {
		clusterName := clusterPrefix + "-" + strconv.Itoa(idx)
		createWs(kcpPath, clusterName)
		createKindCluster(clusterName)
	}

	fmt.Printf("Deploying syncer in each cluster\n")
	for idx := 0; idx < numberOfClusters; idx++ {
		wsName := clusterPrefix + "-" + strconv.Itoa(idx)
		setupSyncer(kcpPath, orgName, wsName, syncerImage)
	}
}
