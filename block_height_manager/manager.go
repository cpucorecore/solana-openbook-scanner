package block_height_manager

import (
	"sync/atomic"
)

type BlockHeightManager interface {
	Init(int64)
	CanCommit(int64) bool
	Commit(int64) bool
	Get() int64
}

type blockHeightManager struct {
	curHeight int64
}

func NewBlockHeightManager() BlockHeightManager {
	return &blockHeightManager{}
}

func (b *blockHeightManager) Init(height int64) {
	b.curHeight = height
}

func (b *blockHeightManager) CanCommit(height int64) bool {
	curHeight := atomic.LoadInt64(&b.curHeight)
	return height == curHeight+1
}

func (b *blockHeightManager) Commit(height int64) bool {
	old := atomic.LoadInt64(&b.curHeight)
	return atomic.CompareAndSwapInt64(&b.curHeight, old, height)
}

func (b *blockHeightManager) Get() int64 {
	return atomic.LoadInt64(&b.curHeight)
}
