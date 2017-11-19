package proto

type NodeInfo struct {
	ID        string
	Advertise string
	Version   string
	API       string
}

type NodesInfo []NodeInfo

func (c NodesInfo) Len() int {
	return len(c)
}

func (c NodesInfo) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}

func (c NodesInfo) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
