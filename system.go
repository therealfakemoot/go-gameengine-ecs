package ecs

import "unsafe"

// System is the interface that controls a type of components assigned to entities
//
// New returns a unsafe.Pointer to a given id
// If the id already has an allocation, the same reference is returned instead
//
// Recycle recycles the block used by id
//
// Get returns the unsafe.Pointer for the id or nil if the id don't have a block in this system
//
// Set sets the contents of interface to id's component data
// The copy is limited to size of objRef, calculated when the system was created
//
// Zero fills the contents of id's component data with zeros
//
// Reset discards all memory used and sets the system to its initial state
type System interface {
	New(id ID) unsafe.Pointer
	Recycle(id ID)
	Get(id ID) unsafe.Pointer
	Set(id ID, i interface{})
	Zero(id ID)
	Reset()
}

type system struct {
	indices SparseArray
	factory MemoryPool
}

// NewSystem returns a new system identified by id, for allocations of type objRef
// The forceAlignment may be used to guarantee that allocations will start on multiples of this value in bytes
// (for example, when using some SIMD instructions that require alignment of 128, 256 or 512 bits)
// The forceAlignment argument is ignored if it's value is negative or zero
func NewSystem(id ID, objRef interface{}, forceAlignment int) System {
	return &system{
		NewSparseArray(),
		NewMemoryPool(id, objRef, forceAlignment),
	}
}

func (s *system) New(id ID) unsafe.Pointer {
	index := s.indices.Get(Index(id))
	if index != InvalidIndex {
		// var t *T
		// log.Printf("[System[%s].New] trying to instantiate already used ID! (id: %d)\n", reflect.TypeOf(t), id)
		return s.factory.Get(Index(index))
	}
	t, compId := s.factory.Alloc()
	s.indices.Set(Index(id), Index(compId))
	return t
}

func (s *system) Recycle(id ID) {
	index := s.indices.Get(Index(id))
	if index != InvalidIndex {
		s.indices.Invalidate(Index(id))
		s.factory.Free(Index(index))
	}
}

func (s *system) Get(id ID) unsafe.Pointer {
	index := s.indices.Get(Index(id))
	if index == InvalidIndex {
		// var t *T
		// log.Printf("[System[%s].New] trying to get not alloc'ed ID (id: %d)", reflect.TypeOf(t), id)
		return nil
	}
	return s.factory.Get(Index(index))
}

func (s *system) Set(id ID, t interface{}) {
	index := s.indices.Get(Index(id))
	if index != InvalidIndex {
		s.factory.Set(index, t)
	}
}

func (s *system) Zero(id ID) {
	index := s.indices.Get(Index(id))
	if index != InvalidIndex {
		s.factory.Zero(index)
	}
}

func (s *system) Reset() {
	s.indices.Reset()
	s.factory.Reset()
}
