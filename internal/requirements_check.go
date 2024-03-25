package internal

import (
	"log"
	"os/exec"
	"os/user"
	"strings"
)

func checkPermissions(token string, namespace string) bool {
	log.Printf("Checking permissions to read logs on '%s' namespace\n", namespace)
	cmdName := "/usr/local/bin/kubectl"
	cmdArgs := []string{"auth", "can-i", "get", "pods", "--subresource=log", "--namespace=" + namespace, "--token=" + token}

	cmd := exec.Command(cmdName, cmdArgs...)
	output, err := cmd.CombinedOutput() // Use CombinedOutput to get both STDOUT and STDERR

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Printf("Service account didn't have permissions to read logs on namespace '%s': %s\n", namespace, exitError.String())
		} else {
			log.Printf("An error occurred while checking permissions: %v\n", err)
		}
		return false
	}

	// Check the output when the command exits with status 0.
	if strings.TrimSpace(string(output)) == "yes" {
		log.Printf("Service account has permissions to read logs on namespace '%s'\n", namespace)
		return true
	}

	log.Printf("Service account doesn't have permissions to read logs on namespace '%s'\n", namespace)
	return false
}

// checkPrerequisites checks if the current user is root and if kubectl is installed.
func CheckPrerequisites() bool {
	if isRoot, err := checkIfRoot(); err != nil {
		log.Printf("Error checking root status: %v\n", err)
		return false
	} else if !isRoot {
		log.Println("User is not root. Please run as root.")
		return false
	}

	if isKubectlInstalled, err := checkIfKubectlInstalled(); err != nil {
		log.Printf("Error checking kubectl installation: %v\n", err)
		return false
	} else if !isKubectlInstalled {
		log.Println("kubectl is not installed. Please install kubectl.")
		return false
	}

	return true
}

// checkIfRoot checks if the current user is root.
func checkIfRoot() (bool, error) {
	// Assume root user has UID 0.
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}
	return currentUser.Uid == "0", nil
}

// checkIfKubectlInstalled checks if kubectl is installed.
func checkIfKubectlInstalled() (bool, error) {
	_, err := exec.LookPath("kubectl")
	if err != nil {
		return false, err
	}
	return true, nil
}
