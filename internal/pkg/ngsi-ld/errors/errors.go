package errors

import (
	"encoding/json"
	"net/http"
)

//ProblemDetails stores details about a certain problem according to RFC7807
//See https://tools.ietf.org/html/rfc7807
type ProblemDetails interface {
	ContentType() string
	Type() string
	Title() string
	Detail() string
	MarshalJSON() ([]byte, error)
	WriteResponse(w http.ResponseWriter)
}

type ProblemDetailsImpl struct {
	typ    string
	title  string
	detail string
}

var problemContentType = "application/problem+json"

//InvalidRequest reports that the request associated to the operation is syntactically
//invalid or includes wrong content
type InvalidRequest struct {
	ProblemDetailsImpl
}

//NewInvalidRequest creates and returns a new instance of an InvalidRequest with the supplied problem detail
func NewInvalidRequest(detail string) *InvalidRequest {
	return &InvalidRequest{
		ProblemDetailsImpl: ProblemDetailsImpl{
			typ:    "https://uri.etsi.org/ngsi-ld/errors/InvalidRequest",
			title:  "Invalid Request",
			detail: detail,
		},
	}
}

//ReportNewInvalidRequest creates an InvalidRequest instance and sends it to the supplied http.ResponseWriter
func ReportNewInvalidRequest(w http.ResponseWriter, detail string) {
	ir := NewInvalidRequest(detail)
	ir.WriteResponse(w)
}

//InternalError reports that there has been an error during the operation execution
type InternalError struct {
	ProblemDetailsImpl
}

//NewInternalError creates and returns a new instance of an InternalError with the supplied problem detail
func NewInternalError(detail string) *InternalError {
	return &InternalError{
		ProblemDetailsImpl: ProblemDetailsImpl{
			typ:    "https://uri.etsi.org/ngsi-ld/errors/InternalError",
			title:  "Internal Error",
			detail: detail,
		},
	}
}

//ReportNewInternalError creates an InternalError instance and sends it to the supplied http.ResponseWriter
func ReportNewInternalError(w http.ResponseWriter, detail string) {
	ie := NewInternalError(detail)
	ie.WriteResponse(w)
}

//ContentType returns the ContentType to be used when returning this problem
func (p *ProblemDetailsImpl) ContentType() string {
	return problemContentType
}

func (p *ProblemDetailsImpl) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	}{
		Type:   p.typ,
		Title:  p.title,
		Detail: p.detail,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

//ResponseCode returns the HTTP response code to be used when returning a specific problem
func (p *ProblemDetailsImpl) ResponseCode() int {
	return http.StatusBadRequest
}

func (p *ProblemDetailsImpl) WriteResponse(w http.ResponseWriter) {
	w.WriteHeader(p.ResponseCode())
	w.Header().Add("Content-Type", p.ContentType())
	w.Header().Add("Content-Language", "en")

	pdbytes, err := json.MarshalIndent(p, "", "  ")
	if err == nil {
		w.Write(pdbytes)
	}
	// else write a 500 error ...
}
