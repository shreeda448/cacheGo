package core

// Evicts the first key it found while iterating map
// TODO: Make it efficient by doing through sampling
func evictFirst() {
	for k := range store {
		delete(store, k)
		return
	}
}

// TODO: Make the eviction strategy configuration driven
// TODO: Support multiple eviction strategies
func evict() {
	evictFirst()
}
