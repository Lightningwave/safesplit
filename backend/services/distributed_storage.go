package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type StoredShare struct {
	Index            int     `json:"index"`
	Value            string  `json:"value"`
	HolderType       string  `json:"holder_type"`
	EncryptionNonce  []byte  `json:"encryption_nonce"`
	MasterKeyVersion *int    `json:"master_key_version,omitempty"`
	ServerKeyID      *string `json:"server_key_id,omitempty"`
}

type DistributedStorageService struct {
	basePath  string
	nodePaths []string
}

func NewDistributedStorageService(basePath string, nodeCount int) (*DistributedStorageService, error) {
	log.Printf("Initializing distributed storage at %s with %d nodes", basePath, nodeCount)

	// Create base storage directory
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create nodes directory
	nodesDir := filepath.Join(basePath, "nodes")
	if err := os.MkdirAll(nodesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create nodes directory: %w", err)
	}

	nodePaths := make([]string, nodeCount)
	for i := 0; i < nodeCount; i++ {
		// Create node directory
		nodePath := filepath.Join(nodesDir, fmt.Sprintf("node_%d", i))
		if err := os.MkdirAll(nodePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create node directory %s: %w", nodePath, err)
		}

		// Create fragments and shards directories for each node
		if err := os.MkdirAll(filepath.Join(nodePath, "fragments"), 0755); err != nil {
			return nil, fmt.Errorf("failed to create fragments directory: %w", err)
		}
		if err := os.MkdirAll(filepath.Join(nodePath, "shards"), 0755); err != nil {
			return nil, fmt.Errorf("failed to create shards directory: %w", err)
		}

		nodePaths[i] = nodePath
		log.Printf("Created node directory: %s", nodePath)
	}

	return &DistributedStorageService{
		basePath:  basePath,
		nodePaths: nodePaths,
	}, nil
}

// NodeCount returns the number of available storage nodes
func (s *DistributedStorageService) NodeCount() int {
	return len(s.nodePaths)
}

// StoreShards distributes and stores file shards across nodes
func (s *DistributedStorageService) StoreShards(fileID uint, shards [][]byte) error {
	log.Printf("Storing %d shards for file %d", len(shards), fileID)

	fileDir := fmt.Sprintf("file_%d", fileID)

	for i, shard := range shards {
		nodeIndex := i % len(s.nodePaths)
		shardPath := filepath.Join(s.nodePaths[nodeIndex], "shards", fileDir)

		if err := os.MkdirAll(shardPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory in node %d: %w", nodeIndex, err)
		}

		// Store shard
		fullPath := filepath.Join(shardPath, fmt.Sprintf("shard_%d", i))
		if err := os.WriteFile(fullPath, shard, 0600); err != nil {
			return fmt.Errorf("failed to write shard %d to node %d: %w", i, nodeIndex, err)
		}

		log.Printf("Stored shard %d in node %d: %s", i, nodeIndex, fullPath)
	}

	return nil
}

// RetrieveShards collects shards for a file from nodes
func (s *DistributedStorageService) RetrieveShards(fileID uint, totalShards int) ([][]byte, error) {
	log.Printf("Retrieving %d shards for file %d", totalShards, fileID)

	shards := make([][]byte, totalShards)
	retrievedCount := 0
	dataShards := totalShards - 2
	fileDir := fmt.Sprintf("file_%d", fileID)

	for shardIndex := 0; shardIndex < totalShards; shardIndex++ {
		nodeIndex := shardIndex % len(s.nodePaths)
		fullPath := filepath.Join(s.nodePaths[nodeIndex], "shards", fileDir,
			fmt.Sprintf("shard_%d", shardIndex))

		data, err := os.ReadFile(fullPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("error reading shard %d: %w", shardIndex, err)
			}
			log.Printf("Shard %d missing from node %d", shardIndex, nodeIndex)
			continue
		}

		shards[shardIndex] = data
		retrievedCount++
		log.Printf("Retrieved shard %d from node %d: %s", shardIndex, nodeIndex, fullPath)
	}

	if retrievedCount < dataShards {
		return nil, fmt.Errorf("insufficient shards: found %d, need %d",
			retrievedCount, dataShards)
	}

	return shards, nil
}

// StoreFragment stores a single key fragment in a node
func (s *DistributedStorageService) StoreFragment(nodeIndex int, fragmentPath string, data []byte) error {
	if nodeIndex >= len(s.nodePaths) {
		return fmt.Errorf("invalid node index: %d", nodeIndex)
	}

	fullPath := filepath.Join(s.nodePaths[nodeIndex], "fragments", fragmentPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create fragment directory: %w", err)
	}

	if err := os.WriteFile(fullPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write fragment: %w", err)
	}

	log.Printf("Stored fragment in node %d: %s", nodeIndex, fullPath)
	return nil
}

// RetrieveFragment retrieves a single key fragment from a node
func (s *DistributedStorageService) RetrieveFragment(nodeIndex int, fragmentPath string) ([]byte, error) {
	if nodeIndex >= len(s.nodePaths) {
		return nil, fmt.Errorf("invalid node index: %d", nodeIndex)
	}

	fullPath := filepath.Join(s.nodePaths[nodeIndex], "fragments", fragmentPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fragment: %w", err)
	}

	log.Printf("Retrieved fragment from node %d: %s", nodeIndex, fullPath)
	return data, nil
}

// DeleteFragment removes a single key fragment from a node
func (s *DistributedStorageService) DeleteFragment(nodeIndex int, fragmentPath string) error {
	if nodeIndex >= len(s.nodePaths) {
		return fmt.Errorf("invalid node index: %d", nodeIndex)
	}

	fullPath := filepath.Join(s.nodePaths[nodeIndex], "fragments", fragmentPath)
	if err := os.Remove(fullPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete fragment: %w", err)
		}
	}

	log.Printf("Deleted fragment from node %d: %s", nodeIndex, fullPath)
	return nil
}

// DeleteShards removes all shards and fragments for a file
func (s *DistributedStorageService) DeleteShards(fileID uint) error {
	log.Printf("Deleting shards and fragments for file %d", fileID)
	fileDir := fmt.Sprintf("file_%d", fileID)

	for nodeIndex, nodePath := range s.nodePaths {
		// Delete shards
		shardDir := filepath.Join(nodePath, "shards", fileDir)
		if err := os.RemoveAll(shardDir); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: failed to delete shards from node %d: %v", nodeIndex, err)
			}
		}

		// Delete fragments
		fragmentDir := filepath.Join(nodePath, "fragments", fileDir)
		if err := os.RemoveAll(fragmentDir); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Warning: failed to delete fragments from node %d: %v", nodeIndex, err)
			}
		}

		log.Printf("Deleted data from node %d", nodeIndex)
	}

	return nil
}
