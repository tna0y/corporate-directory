package transport

import (
	"context"
	"corporate-directory/pkg/service"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
	"time"
)

//
// Transport structures and functions responsible for request / response serialization
//

type setupRequest struct {
	Employees []*service.Employee `json:"employees"`
}

type setupResponse struct {
	Error string `json:"error,omitempty"`
}

type commonManagerRequest struct {
	First  int `json:"first"`
	Second int `json:"second"`
}

type commonManagerResponse struct {
	Common *service.Employee `json:"common,omitempty"`
	Error  string            `json:"error,omitempty"`
}

type getEmployeeRequest struct {
	Id int `json:"first"`
}

type getEmployeeResponse struct {
	Employee *service.Employee `json:"employee,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type getEmployeesResponse struct {
	Employees []*service.Employee `json:"employees"`
	Error     string              `json:"error,omitempty"`
}

func makeSetupEndpoint(svc service.CorporateDirectory) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(setupRequest)
		err := svc.Setup(req.Employees)
		if err != nil {
			return setupResponse{err.Error()}, nil
		}
		return setupResponse{""}, nil
	}
}

func makeCommonManagerEndpoint(svc service.CorporateDirectory) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(commonManagerRequest)
		res, err := svc.GetCommonManager(req.First, req.Second)
		if err != nil {
			return commonManagerResponse{Common: nil, Error: err.Error()}, nil
		}
		return commonManagerResponse{Common: res, Error: ""}, nil
	}
}

func makeGetEmployeeEndpoint(svc service.CorporateDirectory) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(getEmployeeRequest)
		res, err := svc.GetEmployee(req.Id)
		if err != nil {
			return getEmployeeResponse{Employee: nil, Error: err.Error()}, nil
		}
		return getEmployeeResponse{Employee: res, Error: ""}, nil
	}
}

func makeGetEmployeesEndpoint(svc service.CorporateDirectory) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		res, err := svc.GetEmployees()
		if err != nil {
			return getEmployeesResponse{Employees: nil, Error: err.Error()}, nil
		}
		return getEmployeesResponse{Employees: res, Error: ""}, nil
	}
}

func decodeSetupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request setupRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCommonManagerRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request commonManagerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeGetEmployeeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request getEmployeeRequest
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, err
	}
	request.Id = id
	return request, nil
}

func decodeGetEmployeesRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

// Function to set up all endpoints, encoders, router and HTTP server to serve requests.
func SetupHttpTransport(svc service.CorporateDirectory) *http.Server {
	setup := makeSetupEndpoint(svc)
	setupHandler := httptransport.NewServer(setup, decodeSetupRequest, encodeResponse)

	common := makeCommonManagerEndpoint(svc)
	commonHandler := httptransport.NewServer(common, decodeCommonManagerRequest, encodeResponse)

	one := makeGetEmployeeEndpoint(svc)
	oneHandler := httptransport.NewServer(one, decodeGetEmployeeRequest, encodeResponse)

	all := makeGetEmployeesEndpoint(svc)
	allHandler := httptransport.NewServer(all, decodeGetEmployeesRequest, encodeResponse)

	router := httprouter.New()
	router.Handler("POST", "/setup", setupHandler)
	router.Handler("GET", "/common", commonHandler)
	router.Handler("GET", "/employees/:id", oneHandler)
	router.Handler("GET", "/employees", allHandler)
	return &http.Server{
		Addr:           ":80",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
