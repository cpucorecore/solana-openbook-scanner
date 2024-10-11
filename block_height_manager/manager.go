package block_height_manager

import (
	"sync/atomic"
)

type BlockHeightManager interface {
	CanCommit(height uint64) bool
	Commit(height uint64) bool
	Get() uint64
}

type blockHeightManager struct {
	finishedHeight uint64
}

func NewBlockHeightManager(startHeight uint64) BlockHeightManager {
	return &blockHeightManager{
		finishedHeight: startHeight - 1,
	}
}

func (b *blockHeightManager) CanCommit(height uint64) bool {
	finishedHeight := atomic.LoadUint64(&b.finishedHeight)
	return height == finishedHeight+1
}

func (b *blockHeightManager) Commit(height uint64) bool {
	old := atomic.LoadUint64(&b.finishedHeight)
	return atomic.CompareAndSwapUint64(&b.finishedHeight, old, height)
}

func (b *blockHeightManager) Get() uint64 {
	return atomic.LoadUint64(&b.finishedHeight)
}
