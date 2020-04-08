package ngsi

//ContextRegistry is where Context Sources register the information that they can provide
type ContextRegistry interface {
	GetContextSourcesForQuery(query Query) []ContextSource

	Register(source ContextSource)
}

//NewContextRegistry initializes and returns a new default context registry without
//any registered context sources
func NewContextRegistry() ContextRegistry {
	return &registry{}
}

type registry struct {
	sources []ContextSource
}

func (r *registry) GetContextSourcesForQuery(query Query) []ContextSource {
	matchingSources := []ContextSource{}

	entityTypeNames := query.EntityTypes()
	entityAttributeNames := query.EntityAttributes()

	// TODO: Fix potential race issue
	for _, src := range r.sources {
		for _, typeName := range entityTypeNames {
			if typeName == "" || src.ProvidesType(typeName) {
				for _, attributeName := range entityAttributeNames {
					if attributeName == "" || src.ProvidesAttribute(attributeName) {
						matchingSources = append(matchingSources, src)
						break
					}
				}
			}
		}
	}

	return matchingSources
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
