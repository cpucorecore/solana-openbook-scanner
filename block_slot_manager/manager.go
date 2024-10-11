package block_slot_manager

import (
	"sync/atomic"
)

type BlockSlotManager interface {
	CanCommit(uint64) bool
	Commit(uint64) bool
	Get() uint64
}

type blockSlotManager struct {
	curSlot uint64
}

func NewBlockSlotManager(startSlot uint64) BlockSlotManager {
	return &blockSlotManager{
		curSlot: startSlot - 1,
	}
}

func (b *blockSlotManager) CanCommit(parentSlot uint64) bool {
	curSlot := atomic.LoadUint64(&b.curSlot)
	return parentSlot == curSlot
}

func (b *blockSlotManager) Commit(slot uint64) bool {
	old := atomic.LoadUint64(&b.curSlot)
	return atomic.CompareAndSwapUint64(&b.curSlot, old, slot)
}

func (b *blockSlotManager) Get() uint64 {
	return atomic.LoadUint64(&b.curSlot)
}
