package artemis

import (
	"reflect"
)

/**
 * The most raw entity system. It should not typically be used, but you can create your own
 * entity system handling by extending this. It is recommended that you use the other provided
 * entity system implementations.
 */
type BaseSystem struct {
	systemIndex uint
	actives     []*Entity
	passive     bool
	dummy       bool

	*Aspect

	BaseEntityObserver
	BaseWorldSaver
	BaseInitializer
}

// Creates an entity system that uses the specified aspect as a matcher against entities.
func NewBaseSystem(aspect *Aspect) (es *BaseSystem) {
	es = &BaseSystem{
		//actives = new Bag<Entity>();
		Aspect: aspect,
	}

	es.systemIndex = _SystemIndexManager.getIndexFor(es)
	// This system can't possibly be interested in any entity, so it must be "dummy"
	es.dummy = es.allSet.Len() == 0 && es.oneSet.Len() == 0
	return
}

func (es *BaseSystem) Process() {
	if es.checkProcessing() {
		es.begin()
		es.processEntities(es.actives)
		es.end()
	}
}

// Called before processing of entities begins.
func (es *BaseSystem) begin() {}

// Called after the processing of entities ends.
func (es *BaseSystem) end() {}

/**
 * Any implementing entity system must implement this method and the logic
 * to process the given entities of the system.
 *
 * @param entities the entities this system contains.
 */
//protected abstract void processEntities(ImmutableBag<Entity> entities);
func (es *BaseSystem) processEntities(entities []*Entity) bool { return false }

// return true if the system should be processed, false if not.
//protected abstract boolean checkProcessing();
func (es *BaseSystem) checkProcessing() bool { return false }

// Override to implement code that gets executed when systems are initialized.
//protected void initialize() {};

/**
 * Called if the system has received a entity it is interested in, e.g. created or a component was added to it.
 * @param e the entity that was added to this system.
 */
func (es *BaseSystem) Inserted(e *Entity) {}

/**
 * Called if a entity was removed from this system, e.g. deleted or had one of it's components removed.
 * @param e the entity that was removed from this system.
 */
func (es *BaseSystem) Removed(e *Entity) {}

// Will check if the entity is of interest to this system.
func (es *BaseSystem) check(e *Entity) {
	if es.dummy {
		return
	}

	contains := e.SystemBits().Test(es.systemIndex)
	interested := true // possibly interested, let's try to prove it wrong.

	componentBits := e.ComponentBits()

	// Check if the entity possesses ALL of the components defined in the aspect.
	if es.allSet.Len() > 0 {
		for i, ok := es.allSet.NextSet(0); ok; i, ok = es.allSet.NextSet(i + 1) {
			if !componentBits.Test(i) {
				interested = false
				break
			}
		}
	}

	// Check if the entity possesses ANY of the exclusion components, if it does then the system is not interested.
	if es.exclusionSet.Len() > 0 && interested {
		interested = !es.exclusionSet.Intersection(componentBits).Any()
	}

	// Check if the entity possesses ANY of the components in the oneSet. If so, the system is interested.
	if es.oneSet.Len() > 0 {
		interested = es.oneSet.Intersection(componentBits).Any()
	}

	if interested && !contains {
		es.insertToSystem(e)
	} else if !interested && contains {
		es.removeFromSystem(e)
	}
}

func (es *BaseSystem) removeFromSystem(e *Entity) {
	for i, other := range es.actives {
		if other == e {
			es.actives = append(es.actives[:i], es.actives[i+1:]...)
			break
		}
	}
	e.SystemBits().Clear(es.systemIndex)
	es.Removed(e)
}

func (es *BaseSystem) insertToSystem(e *Entity) {
	es.actives = append(es.actives, e)
	e.SystemBits().Set(es.systemIndex)
	es.Inserted(e)
}

func (es *BaseSystem) Added(e *Entity)   { es.check(e) }
func (es *BaseSystem) Changed(e *Entity) { es.check(e) }

func (es *BaseSystem) Deleted(e *Entity) {
	if e.SystemBits().Test(es.systemIndex) {
		es.removeFromSystem(e)
	}
}

func (es *BaseSystem) Disabled(e *Entity) {
	if e.SystemBits().Test(es.systemIndex) {
		es.removeFromSystem(e)
	}
}

func (es *BaseSystem) Enabled(e *Entity)       { es.check(e) }
func (es *BaseSystem) SetWorld(world *World)   { es.world = world }
func (es *BaseSystem) IsPassive() bool         { return es.passive }
func (es *BaseSystem) SetPassive(passive bool) { es.passive = passive }
func (es *BaseSystem) Actives() []*Entity      { return es.actives }

// Used to generate a unique bit for each system. Only used internally in BaseSystem.
var _SystemIndexManager = SystemIndexManager{
	INDEX:   0,
	indices: make(map[reflect.Type]uint),
}

type SystemIndexManager struct {
	INDEX   uint
	indices map[reflect.Type]uint
}

func (m *SystemIndexManager) getIndexFor(es System) uint {
	t := reflect.TypeOf(es)
	index, ok := m.indices[t]
	if !ok {
		m.INDEX++
		index = m.INDEX
		m.indices[t] = index
	}
	return index
}