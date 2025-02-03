package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Node struct {
	client     *s3.Client
	bucketName string
	region     string
}

type MultiS3StorageService struct {
	nodes []S3Node
}

func NewMultiS3StorageService(configs []struct {
	Region     string
	BucketName string
}) (*MultiS3StorageService, error) {
	nodes := make([]S3Node, len(configs))

	for i, cfg := range configs {
		// Load region-specific configuration
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           fmt.Sprintf("https://s3.%s.amazonaws.com", cfg.Region),
				SigningRegion: cfg.Region,
			}, nil
		})

		awsCfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.Region),
			config.WithEndpointResolverWithOptions(customResolver),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to load AWS config for region %s: %w", cfg.Region, err)
		}

		// Create S3 client for this region
		client := s3.NewFromConfig(awsCfg)

		nodes[i] = S3Node{
			client:     client,
			bucketName: cfg.BucketName,
			region:     cfg.Region,
		}

		log.Printf("Initialized S3 node in region %s with bucket %s", cfg.Region, cfg.BucketName)
	}

	return &MultiS3StorageService{
		nodes: nodes,
	}, nil
}

// Rest of the methods remain the same...

func (s *MultiS3StorageService) NodeCount() int {
	return len(s.nodes)
}

func (s *MultiS3StorageService) StoreShards(fileID uint, shards [][]byte) error {
	ctx := context.Background()

	for i, shard := range shards {
		nodeIndex := i % len(s.nodes)
		node := s.nodes[nodeIndex]

		key := fmt.Sprintf("shards/file_%d/shard_%d", fileID, i)

		_, err := node.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(node.bucketName),
			Key:    aws.String(key),
			Body:   bytes.NewReader(shard),
		})
		if err != nil {
			return fmt.Errorf("failed to store shard %d in region %s: %w",
				i, node.region, err)
		}

		log.Printf("Stored shard %d in region %s: s3://%s/%s",
			i, node.region, node.bucketName, key)
	}

	return nil
}

func (s *MultiS3StorageService) RetrieveShards(fileID uint, totalShards int) ([][]byte, error) {
	ctx := context.Background()
	shards := make([][]byte, totalShards)
	retrievedCount := 0
	dataShards := totalShards - 2

	for shardIndex := 0; shardIndex < totalShards; shardIndex++ {
		nodeIndex := shardIndex % len(s.nodes)
		node := s.nodes[nodeIndex]

		key := fmt.Sprintf("shards/file_%d/shard_%d", fileID, shardIndex)

		result, err := node.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(node.bucketName),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("Shard %d missing from region %s: %v", shardIndex, node.region, err)
			continue
		}

		data, err := io.ReadAll(result.Body)
		result.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("error reading shard %d data from region %s: %w",
				shardIndex, node.region, err)
		}

		shards[shardIndex] = data
		retrievedCount++
		log.Printf("Retrieved shard %d from region %s: s3://%s/%s",
			shardIndex, node.region, node.bucketName, key)
	}

	if retrievedCount < dataShards {
		return nil, fmt.Errorf("insufficient shards: found %d, need %d",
			retrievedCount, dataShards)
	}

	return shards, nil
}

func (s *MultiS3StorageService) StoreFragment(nodeIndex int, fragmentPath string, data []byte) error {
	if nodeIndex >= len(s.nodes) {
		return fmt.Errorf("invalid node index: %d", nodeIndex)
	}

	node := s.nodes[nodeIndex]
	key := fmt.Sprintf("fragments/%s", fragmentPath)

	_, err := node.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(node.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to store fragment in region %s: %w", node.region, err)
	}

	log.Printf("Stored fragment in region %s: s3://%s/%s",
		node.region, node.bucketName, key)
	return nil
}

func (s *MultiS3StorageService) RetrieveFragment(nodeIndex int, fragmentPath string) ([]byte, error) {
	if nodeIndex >= len(s.nodes) {
		return nil, fmt.Errorf("invalid node index: %d", nodeIndex)
	}

	node := s.nodes[nodeIndex]
	key := fmt.Sprintf("fragments/%s", fragmentPath)

	result, err := node.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(node.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve fragment from region %s: %w",
			node.region, err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading fragment data from region %s: %w",
			node.region, err)
	}

	log.Printf("Retrieved fragment from region %s: s3://%s/%s",
		node.region, node.bucketName, key)
	return data, nil
}

func (s *MultiS3StorageService) DeleteFragment(nodeIndex int, fragmentPath string) error {
	if nodeIndex >= len(s.nodes) {
		return fmt.Errorf("invalid node index: %d", nodeIndex)
	}

	node := s.nodes[nodeIndex]
	key := fmt.Sprintf("fragments/%s", fragmentPath)

	_, err := node.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(node.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete fragment from region %s: %w", node.region, err)
	}

	log.Printf("Deleted fragment from region %s: s3://%s/%s",
		node.region, node.bucketName, key)
	return nil
}

func (s *MultiS3StorageService) DeleteShards(fileID uint) error {
	ctx := context.Background()

	for i, node := range s.nodes {
		prefix := fmt.Sprintf("shards/file_%d/", fileID)

		input := &s3.ListObjectsV2Input{
			Bucket: aws.String(node.bucketName),
			Prefix: aws.String(prefix),
		}

		paginator := s3.NewListObjectsV2Paginator(node.client, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				log.Printf("Warning: failed to list objects in node %d (region %s): %v",
					i, node.region, err)
				continue
			}

			for _, obj := range page.Contents {
				_, err := node.client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String(node.bucketName),
					Key:    obj.Key,
				})
				if err != nil {
					log.Printf("Warning: failed to delete object %s from node %d (region %s): %v",
						*obj.Key, i, node.region, err)
				}
			}
		}

		log.Printf("Deleted shards from node %d (region %s)", i, node.region)
	}

	return nil
}
