package controller

type KubeletPodsSpec struct {
	Paths      []string
	Namespaces []string
	Pods       []string
	UUIDs      []string
}

func (s KubeletPodsSpec) GetData() [][]string {
	data := make([][]string, len(s.UUIDs))
	for i := range s.UUIDs {
		data[i] = []string{s.Namespaces[i], s.Pods[i], s.UUIDs[i]}
	}
	return data
}
