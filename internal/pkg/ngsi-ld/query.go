package ngsi

import (
	"net/http"
	"strings"
)

//Query is an interface to be used when passing queries to context registries and sources
type Query interface {
	HasDeviceReference() bool
	Device() string

	EntityAttributes() []string
	EntityTypes() []string

	Request() *http.Request
}

func newQueryFromParameters(req *http.Request, types []string, attributes []string, q string) Query {

	const refDevicePrefix string = "refDevice==\""

	qw := &queryWrapper{request: req, types: types, attributes: attributes}

	if strings.HasPrefix(q, refDevicePrefix) {
		splitElems := strings.Split(q, "\"")
		qw.device = &splitElems[1]
	}

	return qw
}

type queryWrapper struct {
	request    *http.Request
	types      []string
	attributes []string
	device     *string
}

func (q *queryWrapper) HasDeviceReference() bool {
	return q.device != nil
}

func (q *queryWrapper) Device() string {
	return *q.device
}

func (q *queryWrapper) EntityAttributes() []string {
	return q.attributes
}

func (q *queryWrapper) EntityTypes() []string {
	return q.types
}

func (q *queryWrapper) Request() *http.Request {
	return q.request
}
