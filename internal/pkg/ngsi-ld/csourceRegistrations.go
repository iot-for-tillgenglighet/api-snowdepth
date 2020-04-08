package ngsi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/google/uuid"
	"github.com/iot-for-tillgenglighet/api-snowdepth/internal/pkg/ngsi-ld/errors"
)

//CsourceRegistration is a wrapper for information about a registered context source
type CsourceRegistration interface {
	Endpoint() string
	ProvidesAttribute(attributeName string) bool
	ProvidesType(typeName string) bool
}

//NewRegisterContextSourceHandler handles POST requests for csource registrations
func NewRegisterContextSourceHandler(ctxReg ContextRegistry) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		reg, err := NewCsourceRegistrationFromJSON(body)

		if err != nil {
			errors.ReportNewBadRequestData(
				w,
				"Failed to create registration from payload: "+err.Error(),
			)
			return
		}

		remoteCtxSrc, err := NewRemoteContextSource(reg)

		ctxReg.Register(remoteCtxSrc)

		jsonBytes, err := json.Marshal(remoteCtxSrc)

		w.WriteHeader(http.StatusCreated)
		w.Header().Add("Content-Type", "application/json")
		w.Write(jsonBytes)
	})
}

type remoteResponse struct {
	responseCode int
	headers      http.Header
	bytes        []byte
}

func (rr *remoteResponse) Header() http.Header {
	if rr.headers == nil {
		rr.headers = make(http.Header)
	}
	return rr.headers
}

func (rr *remoteResponse) Write(b []byte) (int, error) {
	fmt.Println("Write: " + string(b))
	rr.bytes = append(rr.bytes, b...)
	return len(b), nil
}

func (rr *remoteResponse) WriteHeader(responseCode int) {
	fmt.Println("Response Code: ", responseCode)
	rr.responseCode = responseCode
}

//NewRemoteContextSource creates an instance of a ContextSource by wrapping a CsourceRegistration
func NewRemoteContextSource(registration CsourceRegistration) (ContextSource, error) {
	return &remoteContextSource{ID: uuid.New().String(), registration: registration}, nil
}

type remoteContextSource struct {
	ID           string `json:"id"`
	registration CsourceRegistration
}

func (rcs *remoteContextSource) GetEntities(query Query, callback QueryEntitiesCallback) error {
	u, err := url.Parse(rcs.registration.Endpoint())
	req := query.Request()

	req.URL.Host = u.Host
	req.URL.Scheme = u.Scheme

	forwardedHost := req.Header.Get("Host")
	if forwardedHost != "" {
		req.Header.Set("X-Forwarded-Host", forwardedHost)
	}
	req.Host = u.Host

	// Change the User-Agent header to something more appropriate
	req.Header.Add("User-Agent", "ngsi-context-broker/0.1")

	// We do not want to propagate the Accept-Encoding header to prevent compression
	req.Header.Del("Accept-Encoding")

	fmt.Println("Forwarding request to: " + req.URL.String())
	b, _ := httputil.DumpRequestOut(req, false)
	fmt.Println(string(b))

	response := &remoteResponse{}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(response, req)

	var unmarshaledResponse []interface{}

	// If the response code is 200 we can just unmarshal the payload
	// and iterate over the array to pass the individual entitites
	// to the supplied callback. This will of course fail if the response
	// payload is not an array, but that is an assignment for another day...
	if response.responseCode == 200 {
		err = json.Unmarshal(response.bytes, &unmarshaledResponse)
		if err == nil {
			for _, e := range unmarshaledResponse {
				callback(e)
			}
		}
	}

	return err
}

func (rcs *remoteContextSource) ProvidesAttribute(attributeName string) bool {
	return rcs.registration.ProvidesAttribute(attributeName)
}

func (rcs *remoteContextSource) ProvidesType(typeName string) bool {
	return rcs.registration.ProvidesType(typeName)
}

type ctxSrcReg struct {
	Type        string          `json:"type"`
	Information []ctxSrcRegInfo `json:"information"`
	Endpt       string          `json:"endpoint"`
}

func (csr *ctxSrcReg) Endpoint() string {
	return csr.Endpt
}

func (csr *ctxSrcReg) ProvidesAttribute(attributeName string) bool {
	for _, reginfo := range csr.Information {
		for _, attr := range reginfo.Properties {
			if attr == attributeName {
				return true
			}
		}
	}
	return false
}

func (csr *ctxSrcReg) ProvidesType(typeName string) bool {
	for _, reginfo := range csr.Information {
		for _, entity := range reginfo.Entities {
			if entity.Type == typeName {
				return true
			}
		}
	}
	return false
}

type entityInfo struct {
	Type string `json:"type"`
}

type ctxSrcRegInfo struct {
	Entities   []entityInfo `json:"entities"`
	Properties []string     `json:"properties"`
}

//NewCsourceRegistration creates and returns a concrete implementation of the CsourceRegistration interface
func NewCsourceRegistration(entityTypeName string, attributeNames []string, endpoint string) CsourceRegistration {
	regInfo := ctxSrcRegInfo{Entities: []entityInfo{}, Properties: attributeNames}
	regInfo.Entities = append(regInfo.Entities, entityInfo{Type: entityTypeName})

	reg := &ctxSrcReg{Type: "ContextSourceRegistration", Endpt: endpoint}
	reg.Information = []ctxSrcRegInfo{regInfo}

	return reg
}

//NewCsourceRegistrationFromJSON unpacks a byte buffer into a CsourceRegistration and validates its contents
func NewCsourceRegistrationFromJSON(jsonBytes []byte) (CsourceRegistration, error) {
	registration := &ctxSrcReg{}
	err := json.Unmarshal(jsonBytes, registration)

	if err != nil {
		return nil, err
	}

	// TODO: Validation ...

	return registration, nil
}
