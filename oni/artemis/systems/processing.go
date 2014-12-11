package systems

import (
	. "oniproject/oni/artemis"
)

// A typical entity system. Use this when you need to process entities possessing the provided component types.
type EntityProcessingSystem struct {
	*BaseSystem
}

func NewEntityProcessingSystem(aspect *Aspect) (sys *EntityProcessingSystem) {
	sys = &EntityProcessingSystem{}
	sys.BaseSystem = NewBaseSystem(aspect, sys)
	return
}

// Process a entity this system is interested in.
func (sys *EntityProcessingSystem) Process(e Entity) {}

func (sys *EntityProcessingSystem) ProcessEntities(entities []Entity) {
	for _, e := range entities {
		sys.Process(e)
	}
}

func (sys *EntityProcessingSystem) CheckProcessing() bool { return true }

func (sys *EntityProcessingSystem) Inserted(e Entity) {}
func (sys *EntityProcessingSystem) Removed(e Entity)  {}
