package model

// KubeletPodsSpec holds the mapping of the /var/lib/kubelet/pods directories and subdirecotires.
type KubeletPodsSpec struct {
	Paths      []string
	Namespaces []string
	Pods       []string
	UUIDs      []string
}

// PodDestinationConfig holds kubernetes configuration nessesary to exploit the scenario.
type PodDestinationConfig struct {
	SaToken       string
	Namespace     string
	PodName       string
	ContainerName string
	UUID          string
	LogFilePath   string
	LogFile       string
	Path          string
	APIServerAddr string
}

// Result holds the tokens and all associated data discovered.
type Result struct {
	Pod       string
	Namespace string
	Suffix    string
	Token     string
	Path      string
}

// Config holds the values of the command-line flags.
type Config struct {
	ServiceAccountToken   string
	SuffixFilePath        string
	SuffixGeneratorFile   string
	HostMountPath         string
	KubernetesServiceHost string
	TargetNamespace       string
	OutputPath            string
}
