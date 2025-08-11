package kernel

import (
	"github.com/elliotchance/orderedmap"
)

type Liveness struct {
	defs           *orderedmap.OrderedMap
	uses           *orderedmap.OrderedMap
	global_symbols *orderedmap.OrderedMap
}

func (this *Liveness) Init() {
	this.defs = orderedmap.NewOrderedMap()
	this.uses = orderedmap.NewOrderedMap()
	this.global_symbols = orderedmap.NewOrderedMap()
}

func (this *Liveness) Defs() *orderedmap.OrderedMap {
	return this.defs
}

func (this *Liveness) AddDef(def string) {
	this.defs.Set(def, true)
}

func (this *Liveness) Uses() *orderedmap.OrderedMap {
	return this.uses
}

func (this *Liveness) AddUse(use string) {
	this.uses.Set(use, true)
}

func (this *Liveness) GlobalSymbols() *orderedmap.OrderedMap {
	return this.global_symbols
}

func (this *Liveness) AddGlobalSymbol(global_symbol string) {
	this.global_symbols.Set(global_symbol, true)
}

func (this *Liveness) LocalSymbols() *orderedmap.OrderedMap {
	local_symbols := orderedmap.NewOrderedMap()
	for el := this.defs.Front(); el != nil; el = el.Next() {
		key := el.Key.(string)
		if _, ok := this.global_symbols.Get(key); !ok {
			local_symbols.Set(key, true)
		}
	}
	return local_symbols
}

func (this *Liveness) UnresolvedSymbols() *orderedmap.OrderedMap {
	unresolved_symbols := orderedmap.NewOrderedMap()
	for el := this.uses.Front(); el != nil; el = el.Next() {
		key := el.Key.(string)
		if _, ok := this.defs.Get(key); !ok {
			unresolved_symbols.Set(key, true)
		}
	}
	return unresolved_symbols
}
