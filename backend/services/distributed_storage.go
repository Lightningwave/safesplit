package services

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
)

// DistributedStorageService handles storage across multiple local "nodes"
type DistributedStorageService struct {
    nodePaths []string
}

// NewDistributedStorageService creates a new storage service with the specified number of nodes
func NewDistributedStorageService(basePath string, nodeCount int) (*DistributedStorageService, error) {
    nodePaths := make([]string, nodeCount)
    
    // Create directory for each node
    for i := 0; i < nodeCount; i++ {
        nodePath := filepath.Join(basePath, fmt.Sprintf("node_%d", i))
        if err := os.MkdirAll(nodePath, 0755); err != nil {
            return nil, fmt.Errorf("failed to create node directory %s: %w", nodePath, err)
        }
        nodePaths[i] = nodePath
    }

    return &DistributedStorageService{
        nodePaths: nodePaths,
    }, nil
}

// StoreShards distributes and stores file shards across nodes
func (s *DistributedStorageService) StoreShards(fileID uint, shards [][]byte) error {
    log.Printf("Storing %d shards for file %d", len(shards), fileID)
    
    // Distribute shards across nodes using round-robin
    for i, shard := range shards {
        nodeIndex := i % len(s.nodePaths)
        nodePath := s.nodePaths[nodeIndex]
        
        // Create file directory in node
        fileDir := filepath.Join(nodePath, fmt.Sprintf("file_%d", fileID))
        if err := os.MkdirAll(fileDir, 0755); err != nil {
            return fmt.Errorf("failed to create directory in node %d: %w", nodeIndex, err)
        }
        
        // Store shard
        shardPath := filepath.Join(fileDir, fmt.Sprintf("shard_%d", i))
        if err := os.WriteFile(shardPath, shard, 0600); err != nil {
            return fmt.Errorf("failed to write shard %d to node %d: %w", i, nodeIndex, err)
        }
        
        log.Printf("Stored shard %d in node %d", i, nodeIndex)
    }
    
    return nil
}

// RetrieveShards collects all shards for a file from the nodes
func (s *DistributedStorageService) RetrieveShards(fileID uint, totalShards int) ([][]byte, error) {
    log.Printf("Retrieving %d shards for file %d", totalShards, fileID)
    
    shards := make([][]byte, totalShards)
    retrievedCount := 0
    
    // Look for each shard in the nodes
    for shardIndex := 0; shardIndex < totalShards; shardIndex++ {
        nodeIndex := shardIndex % len(s.nodePaths)
        nodePath := s.nodePaths[nodeIndex]
        
        shardPath := filepath.Join(nodePath, fmt.Sprintf("file_%d", fileID), fmt.Sprintf("shard_%d", shardIndex))
        data, err := os.ReadFile(shardPath)
        if err != nil {
            if !os.IsNotExist(err) {
                return nil, fmt.Errorf("error reading shard %d: %w", shardIndex, err)
            }
            continue
        }
        
        shards[shardIndex] = data
        retrievedCount++
        log.Printf("Retrieved shard %d from node %d", shardIndex, nodeIndex)
    }
    
    if retrievedCount < totalShards {
        return nil, fmt.Errorf("only found %d shards out of %d", retrievedCount, totalShards)
    }
    
    return shards, nil
}

// DeleteShards removes all shards for a file from all nodes
func (s *DistributedStorageService) DeleteShards(fileID uint) error {
    log.Printf("Deleting shards for file %d", fileID)
    
    for nodeIndex, nodePath := range s.nodePaths {
        fileDir := filepath.Join(nodePath, fmt.Sprintf("file_%d", fileID))
        
        if err := os.RemoveAll(fileDir); err != nil {
            if !os.IsNotExist(err) {
                return fmt.Errorf("failed to delete shards from node %d: %w", nodeIndex, err)
            }
        }
        
        log.Printf("Deleted shards from node %d", nodeIndex)
    }
    
    return nil
}