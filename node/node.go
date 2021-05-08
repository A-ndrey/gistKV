package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	ValueType     = "value"
	DirectoryType = "directory"
)

const RootKey = "root"

var (
	NotFound = errors.New("node not found")
	EmptyKey = errors.New("empty key")
)

type Node struct {
	Key      string  `json:"key"`
	NodeType string  `json:"type"`
	Value    string  `json:"value,omitempty"`
	Nodes    []*Node `json:"nodes,omitempty"`
}

func NewRoot() *Node {
	return &Node{
		Key:      RootKey,
		NodeType: DirectoryType,
	}
}

func (n *Node) FindNode(keys []string) (*Node, []string, error) {
	if len(keys) == 0 {
		return n, nil, EmptyKey
	}

	for idxNode, currNode := range n.Nodes {
		if currNode.Key != keys[0] {
			continue
		}

		if len(keys) == 1 {
			return n.Nodes[idxNode], nil, nil
		}

		if currNode.NodeType != DirectoryType {
			return nil, nil, fmt.Errorf("%s is not a %s", currNode.Key, DirectoryType)
		}

		return n.Nodes[idxNode].FindNode(keys[1:])
	}

	return n, keys, NotFound
}

func (n *Node) IsRoot() bool {
	return n.Key == RootKey && n.NodeType == DirectoryType
}

func (n *Node) CreateSubDir(key string) *Node {
	newNode := &Node{
		Key:      key,
		NodeType: DirectoryType,
	}

	n.Nodes = append(n.Nodes, newNode)

	return newNode
}

func (n *Node) CreateValue(key, value string) *Node {
	newNode := &Node{
		Key:      key,
		NodeType: ValueType,
		Value:    value,
	}

	n.Nodes = append(n.Nodes, newNode)

	return newNode
}

func (n *Node) JsonFormat() (string, error) {
	bytes, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return "", fmt.Errorf("can't marshall to json, %w", err)
	}

	return string(bytes), nil
}

func (n *Node) TreeFormat() (string, error) {
	if n.NodeType == ValueType {
		return fmt.Sprintf("- %s [%s]\n", n.Key, n.Value), nil
	}

	builder := strings.Builder{}
	_, err := builder.WriteString(fmt.Sprintf("+ %s\n", n.Key))
	if err != nil {
		return "", err
	}

	for i := range n.Nodes {
		tree, err := n.Nodes[i].TreeFormat()
		if err != nil {
			return "", err
		}

		treeSplit := strings.Split(strings.TrimRight(tree, "\n"), "\n")
		for i := range treeSplit {
			_, err = builder.WriteString(fmt.Sprintf("| %s\n", treeSplit[i]))
			if err != nil {
				return "", err
			}
		}
	}

	return builder.String(), nil
}
