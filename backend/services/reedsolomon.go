package services

import (
    "encoding/binary"
    "fmt"
    "log"

    "github.com/klauspost/reedsolomon"
)

type ReedSolomonService struct {
    storage *DistributedStorageService
}

type FileShards struct {
    Shards       [][]byte
    OriginalSize uint64
}

// Updated constructor to accept an existing storage service
func NewReedSolomonService(storage *DistributedStorageService) (*ReedSolomonService, error) {
    if storage == nil {
        return nil, fmt.Errorf("storage service cannot be nil")
    }

    return &ReedSolomonService{
        storage: storage,
    }, nil
}

func (s *ReedSolomonService) SplitFile(data []byte, dataShards, parityShards int) (*FileShards, error) {
    // Store original data size
    originalSize := uint64(len(data))

    // Create size prefix (8 bytes for uint64)
    sizeBytes := make([]byte, 8)
    binary.LittleEndian.PutUint64(sizeBytes, originalSize)

    // Prepend size to data
    dataWithSize := append(sizeBytes, data...)

    // Create encoder with auto-goroutines based on data size
    enc, err := reedsolomon.New(dataShards, parityShards,
        reedsolomon.WithAutoGoroutines(len(dataWithSize)))
    if err != nil {
        return nil, fmt.Errorf("failed to create encoder: %w", err)
    }

    // Calculate chunk size (64KB) and shard size
    chunkSize := 64 * 1024
    shardSize := (len(dataWithSize) + dataShards - 1) / dataShards

    // Round up to nearest chunk size
    if shardSize%chunkSize != 0 {
        shardSize = ((shardSize + chunkSize - 1) / chunkSize) * chunkSize
    }

    log.Printf("Splitting file - Original size: %d, Data with size header: %d, Shard size: %d, Total shards: %d",
        originalSize, len(dataWithSize), shardSize, dataShards+parityShards)

    // Create shards
    shards := make([][]byte, dataShards+parityShards)
    for i := range shards {
        shards[i] = make([]byte, shardSize)
    }

    // Copy data into data shards
    remaining := len(dataWithSize)
    for i := 0; i < dataShards && remaining > 0; i++ {
        chunk := shardSize
        if remaining < chunk {
            chunk = remaining
        }
        copy(shards[i], dataWithSize[i*shardSize:i*shardSize+chunk])
        remaining -= chunk
    }

    // Encode parity shards
    if err = enc.Encode(shards); err != nil {
        return nil, fmt.Errorf("failed to encode shards: %w", err)
    }

    log.Printf("Created %d shards", len(shards))

    return &FileShards{
        Shards:       shards,
        OriginalSize: originalSize,
    }, nil
}

func (s *ReedSolomonService) ReconstructFile(shards [][]byte, dataShards, parityShards int) ([]byte, error) {
    log.Printf("Starting reconstruction - Shard size: %d bytes", len(shards[0]))

    enc, err := reedsolomon.New(dataShards, parityShards)
    if err != nil {
        return nil, fmt.Errorf("failed to create RS encoder: %w", err)
    }

    // Reconstruct any missing shards
    if err := enc.Reconstruct(shards); err != nil {
        return nil, fmt.Errorf("failed to reconstruct shards: %w", err)
    }

    // Join the data shards
    joined := make([]byte, 0, len(shards[0])*dataShards)
    for i := 0; i < dataShards; i++ {
        joined = append(joined, shards[i]...)
    }

    // First 8 bytes contain the original size
    if len(joined) < 8 {
        return nil, fmt.Errorf("reconstructed data too short")
    }

    originalSize := binary.LittleEndian.Uint64(joined[:8])
    log.Printf("Calculated data size: %d bytes", originalSize)

    // Verify size and extract data
    if uint64(len(joined)-8) < originalSize {
        return nil, fmt.Errorf("reconstructed data shorter than original size")
    }

    // Extract the actual data
    data := joined[8 : 8+originalSize]
    log.Printf("Successfully reconstructed data: %d bytes", len(data))

    return data, nil
}

func (s *ReedSolomonService) StoreShards(fileID uint, fileShards *FileShards) error {
    log.Printf("Storing %d shards for file %d", len(fileShards.Shards), fileID)
    return s.storage.StoreShards(fileID, fileShards.Shards)
}

func (s *ReedSolomonService) RetrieveShards(fileID uint, totalShards int) (*FileShards, error) {
    log.Printf("Retrieving %d shards for file %d", totalShards, fileID)
    shards, err := s.storage.RetrieveShards(fileID, totalShards)
    if err != nil {
        return nil, err
    }

    // Get original size from first shard
    if len(shards) == 0 || len(shards[0]) < 8 {
        return nil, fmt.Errorf("invalid shard data")
    }

    originalSize := binary.LittleEndian.Uint64(shards[0][:8])

    return &FileShards{
        Shards:       shards,
        OriginalSize: originalSize,
    }, nil
}

func (s *ReedSolomonService) ValidateShards(shards [][]byte, dataShards int) bool {
    validShards := 0
    shardSize := -1

    for i, shard := range shards {
        if shard == nil {
            log.Printf("Shard %d: Missing", i)
            continue
        }

        // Verify shard size consistency
        if shardSize == -1 {
            shardSize = len(shard)
        } else if len(shard) != shardSize {
            log.Printf("Shard %d has inconsistent size: got %d, want %d",
                i, len(shard), shardSize)
            return false
        }

        validShards++
        log.Printf("Shard %d is valid: %d bytes", i, len(shard))
    }

    result := validShards >= dataShards
    log.Printf("Shard validation result: %v (have %d valid shards, need %d)",
        result, validShards, dataShards)

    return result
}

func (s *ReedSolomonService) DeleteShards(fileID uint) error {
    log.Printf("Deleting shards for file %d", fileID)
    return s.storage.DeleteShards(fileID)
}