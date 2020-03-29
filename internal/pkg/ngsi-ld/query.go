package ngsi

import (
	"strings"
)

type Query interface {
	HasDeviceReference() bool
	Device() string
}

func newQueryFromParameters(types []string, attributes []string, q string) Query {

	const refDevicePrefix string = "refDevice==\""

	qw := &queryWrapper{types: types, attributes: attributes}

	if strings.HasPrefix(q, refDevicePrefix) {
		splitElems := strings.Split(q, "\"")
		qw.device = &splitElems[1]
	}

	return qw
}

type queryWrapper struct {
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
