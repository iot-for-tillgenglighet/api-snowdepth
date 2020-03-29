package ngsi

//ContextRegistry is where Context Sources register the information that they can provide
type ContextRegistry interface {
	GetContextSourcesForQuery(query Query) []ContextSource

	Register(source ContextSource)
}

func NewContextRegistry() ContextRegistry {
	return &registry{}
}

type registry struct {
	sources []ContextSource
}

func (r *registry) GetContextSourcesForQuery(query Query) []ContextSource {
	// TODO: Fix potential race issue
	return r.sources
}

func (r *registry) Register(source ContextSource) {
	// TODO: Fix potential race issue
	r.sources = append(r.sources, source)
}

//ContextSource provides query and subscription support for a set of entities
type ContextSource interface {
	ProvidesAttribute(attributeName string) bool
	ProvidesType(typeName string) bool

	GetEntities(query Query, callback QueryEntitiesCallback) error
}
