package tools

type Tool struct {
	Name        string
	Description string
	Parameters  any //用于大模型生成json schema
	Execute     func(args string) (string, error)
}
