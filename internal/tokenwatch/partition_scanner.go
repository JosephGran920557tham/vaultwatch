package tokenwatch

import (
	"context"
	"log"

	"github.com/your-org/vaultwatch/internal/alert"
)

// PartitionScanner fans out alert scanning across partition buckets,
// allowing each bucket to be scanned independently.
type PartitionScanner struct {
	registry  *Registry
	partition *Partition
	lookup    func(ctx context.Context, tokenID string) (TokenInfo, error)
	detect    func(ctx context.Context, info TokenInfo) (*alert.Alert, error)
	logger    *log.Logger
}

// NewPartitionScanner constructs a PartitionScanner.
// Panics if required dependencies are nil.
func NewPartitionScanner(
	registry *Registry,
	partition *Partition,
	lookup func(ctx context.Context, tokenID string) (TokenInfo, error),
	detect func(ctx context.Context, info TokenInfo) (*alert.Alert, error),
	logger *log.Logger,
) *PartitionScanner {
	if registry == nil {
		panic("tokenwatch: PartitionScanner requires a non-nil registry")
	}
	if partition == nil {
		panic("tokenwatch: PartitionScanner requires a non-nil partition")
	}
	if lookup == nil {
		panic("tokenwatch: PartitionScanner requires a non-nil lookup func")
	}
	if detect == nil {
		panic("tokenwatch: PartitionScanner requires a non-nil detect func")
	}
	if logger == nil {
		logger = log.Default()
	}
	return &PartitionScanner{
		registry:  registry,
		partition: partition,
		lookup:    lookup,
		detect:    detect,
		logger:    logger,
	}
}

// Scan iterates all registered tokens, assigns them to buckets, then
// scans each bucket and collects alerts.
func (s *PartitionScanner) Scan(ctx context.Context) ([]*alert.Alert, error) {
	tokens := s.registry.List()
	for _, id := range tokens {
		s.partition.Assign(id)
	}

	var results []*alert.Alert
	for bucket := 0; bucket < s.partition.cfg.MaxBuckets; bucket++ {
		ids := s.partition.Tokens(bucket)
		for _, id := range ids {
			info, err := s.lookup(ctx, id)
			if err != nil {
				s.logger.Printf("partition_scanner: lookup error for token %s: %v", id, err)
				continue
			}
			a, err := s.detect(ctx, info)
			if err != nil {
				s.logger.Printf("partition_scanner: detect error for token %s: %v", id, err)
				continue
			}
			if a != nil {
				results = append(results, a)
			}
		}
	}
	return results, nil
}
