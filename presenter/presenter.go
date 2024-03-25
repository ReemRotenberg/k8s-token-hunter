package presenter

import (
	"k8s-token-hunter/model"
	"os"

	"github.com/olekukonko/tablewriter"
)

type TableData interface {
	GetData() [][]string
	GetHeaders() []string
}

// KubeletPodsSpecPresenter embeds the model.KubeletPodsSpec to allow method attachment.
type KubeletPodsSpecPresenter struct {
	*model.KubeletPodsSpec
}

// PodDestinationConfigPresenter embeds the model.PodDestinationConfig to allow method attachment.
type PodDestinationConfigPresenter struct {
	*model.PodDestinationConfig
}

// ResultPresenter embeds the model.Result to allow method attachment.
type ResultPresenter struct {
	*model.Result // Assuming Result is a struct type defined somewhere in your package.
}

func (k KubeletPodsSpecPresenter) GetData() [][]string {
	data := make([][]string, len(k.UUIDs))
	for i, uuid := range k.UUIDs {
		data[i] = []string{k.Namespaces[i], k.Pods[i], uuid}
	}
	return data
}

func (k KubeletPodsSpecPresenter) GetHeaders() []string {
	return []string{"Namespace", "Pod Name", "UUID"}
}

func (p PodDestinationConfigPresenter) GetData() [][]string {
	return [][]string{{p.Namespace, p.PodName, p.ContainerName, p.LogFile}}
}

func (p PodDestinationConfigPresenter) GetHeaders() []string {
	return []string{"Namespace", "Pod Name", "Container Name", "Log File"}
}

func (r ResultPresenter) GetData() [][]string {
	return [][]string{{r.Pod, r.Namespace, r.Suffix, r.Token[55:65] + "..." + r.Token[len(r.Token)-10:], r.Path}}
}

func (r ResultPresenter) GetHeaders() []string {
	return []string{"Pod Name", "Namespace", "Suffix", "Token", "Path"}
}

func CreateAndRenderTable(data TableData) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(data.GetHeaders())
	for _, row := range data.GetData() {
		table.Append(row)
	}
	table.Render()
}
