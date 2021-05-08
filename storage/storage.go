package storage

import (
	"errors"
	"fmt"
	"gistKV/gist"
	"gistKV/node"
	"strings"
	"sync"
)

const (
	JsonFormat = "json"
	TreeFormat = "tree"
)

var (
	NotForceDeletion = errors.New("can't remove repository")
)

type Storage interface {
	Create(key, value string) error
	Read(key string) (string, error)
	Update(key, value string) error
	Delete(key string, force bool) error
	List(format string) (string, error)
}

type storage struct {
	client gist.Client
	mutex  sync.RWMutex
}

func New(client gist.Client) Storage {
	return &storage{client: client}
}

func (s *storage) Create(key, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	root, err := s.client.Read()
	if err != nil {
		return err
	}

	err = create(root, key, value)
	if err != nil {
		return err
	}

	return s.client.Write(root)
}

func (s *storage) Read(key string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	errorPrefix := fmt.Sprintf("read[key=%s]", key)

	root, err := s.client.Read()
	if err != nil {
		return "", fmt.Errorf("%s, %w", errorPrefix, err)
	}

	foundNode, _, err := root.FindNode(parseKey(key))
	if err != nil {
		return "", fmt.Errorf("%s, can't find node, %w", errorPrefix, err)
	}

	if foundNode.NodeType != node.ValueType {
		return "", fmt.Errorf("%s, node is not a %s", errorPrefix, node.ValueType)
	}

	return foundNode.Value, nil
}

func (s *storage) Update(key, value string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	errorPrefix := fmt.Sprintf("update[key=%s]", key)

	root, err := s.client.Read()
	if err != nil {
		return fmt.Errorf("%s, %w", errorPrefix, err)
	}

	err = remove(root, key, false)
	if err != nil {
		return fmt.Errorf("%s, %w", errorPrefix, err)
	}

	err = create(root, key, value)
	if err != nil {
		return fmt.Errorf("%s, %w", errorPrefix, err)
	}

	err = s.client.Write(root)
	if err != nil {
		return fmt.Errorf("%s, %w", errorPrefix, err)
	}

	return nil
}

func (s *storage) Delete(key string, force bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	errorPrefix := fmt.Sprintf("delete[key=%s]", key)

	root, err := s.client.Read()
	if err != nil {
		return fmt.Errorf("%s, %w", errorPrefix, err)
	}

	err = remove(root, key, force)
	if err != nil {
		return fmt.Errorf("%s, %w", errorPrefix, err)
	}

	err = s.client.Write(root)
	if err != nil {
		return fmt.Errorf("%s, %w", errorPrefix, err)
	}

	return nil
}

func (s *storage) List(format string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	errorPrefix := fmt.Sprintf("list[%s]", format)

	root, err := s.client.Read()
	if err != nil {
		return "", fmt.Errorf("%s, %w", errorPrefix, err)
	}

	switch format {
	case JsonFormat:
		return root.JsonFormat()
	case TreeFormat:
		return root.TreeFormat()
	default:
		return root.JsonFormat()
	}
}

func parseKey(key string) []string {
	return strings.Split(key, "/")
}

func create(root *node.Node, key, value string) error {
	keys := parseKey(key)

	currNode, restKeys, err := root.FindNode(keys)
	if err == nil {
		return fmt.Errorf("node already exists")
	} else if !errors.Is(err, node.NotFound) {
		return fmt.Errorf("can't find node, %w", err)
	}

	if currNode.NodeType != node.DirectoryType {
		return fmt.Errorf("can't create new node inside not directory node")
	}

	for idx := range restKeys[:len(restKeys)-1] {
		currNode = currNode.CreateSubDir(restKeys[idx])
	}

	currNode.CreateValue(restKeys[len(restKeys)-1], value)

	return nil
}

func remove(root *node.Node, key string, force bool) error {
	keys := parseKey(key)

	foundNode, _, err := root.FindNode(keys)
	if err != nil {
		return fmt.Errorf("can't find node, %w", err)
	}

	if foundNode.NodeType == node.DirectoryType && !force {
		return NotForceDeletion
	}

	parentNode, _, err := root.FindNode(keys[:len(keys)-1])
	if err != nil && !errors.Is(err, node.EmptyKey) {
		return fmt.Errorf("can't find parent node, %w", err)
	}

	for idx, currNode := range parentNode.Nodes {
		if currNode == foundNode {
			parentNode.Nodes = append(parentNode.Nodes[:idx], parentNode.Nodes[idx+1:]...)
			return nil
		}
	}

	return fmt.Errorf("can't find node in parent node")
}
