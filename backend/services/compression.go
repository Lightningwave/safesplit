package services

import (
	"log"
	"sync"

	"github.com/klauspost/compress/zstd"
)

type CompressionService struct {
	encoder *zstd.Encoder
	decoder *zstd.Decoder
	mu      sync.Mutex
}

func NewCompressionService() (*CompressionService, error) {
	encoder, err := zstd.NewWriter(nil,
		zstd.WithEncoderLevel(zstd.SpeedBestCompression),
		zstd.WithEncoderConcurrency(1),
	)
	if err != nil {
		return nil, err
	}

	decoder, err := zstd.NewReader(nil)
	if err != nil {
		encoder.Close()
		return nil, err
	}

	return &CompressionService{
		encoder: encoder,
		decoder: decoder,
	}, nil
}

// Compress compresses the input data and returns the compressed data along with the compression ratio
func (s *CompressionService) Compress(data []byte) ([]byte, float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	compressed := s.encoder.EncodeAll(data, make([]byte, 0, len(data)))
	ratio := float64(len(compressed)) / float64(len(data))

	log.Printf("Compression ratio: %.2f%%", ratio*100)
	return compressed, ratio, nil
}

// Decompress decompresses the input data
func (s *CompressionService) Decompress(data []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.decoder.DecodeAll(data, nil)
}

// Close releases resources used by the compression service
func (s *CompressionService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.encoder != nil {
		s.encoder.Close()
	}
	if s.decoder != nil {
		s.decoder.Close()
	}
}
