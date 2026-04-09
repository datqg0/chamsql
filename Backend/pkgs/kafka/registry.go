package kafka

import "sync"

type Registry struct {
	mu     sync.RWMutex
	topics map[string]TopicDefinition
}

func NewRegistry() *Registry {
	return &Registry{
		topics: make(map[string]TopicDefinition),
	}
}

func (r *Registry) Register(td TopicDefinition) {
	if !td.Validate() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.topics[td.Name] = td
}

func (r *Registry) Get(name string) (TopicDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	td, ok := r.topics[name]
	return td, ok
}

func (r *Registry) All() []TopicDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]TopicDefinition, 0, len(r.topics))
	for _, td := range r.topics {
		result = append(result, td)
	}
	return result
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.topics))
	for name := range r.topics {
		names = append(names, name)
	}
	return names
}

func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.topics)
}
