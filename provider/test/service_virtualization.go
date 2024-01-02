package test

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/yunarta/terraform-api-transport/transport"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

type ServiceVirtualization struct {
	router *mux.Router
}

var _ transport.PayloadTransport = &ServiceVirtualization{}

func (service *ServiceVirtualization) SendWithExpectedStatus(request *transport.PayloadRequest, expectedStatus ...int) (*transport.PayloadResponse, error) {
	return service.Send(request)
}

func (service *ServiceVirtualization) Send(request *transport.PayloadRequest) (*transport.PayloadResponse, error) {
	var reader io.Reader
	if request.Payload != nil {
		reader = bytes.NewReader(request.Payload.ContentMust())
	} else {
		reader = nil
	}

	muxRequest, err := http.NewRequest(
		request.Method,
		request.Url,
		reader,
	)
	if err != nil {
		return nil, err
	}

	muxResponse := httptest.NewRecorder()
	service.router.ServeHTTP(muxResponse, muxRequest)
	return &transport.PayloadResponse{
		StatusCode: muxResponse.Code,
		Body:       muxResponse.Body.String(),
	}, nil
}

type BambooRouter struct {
	deployments map[string]bamboo.Deployment
}

func (router *BambooRouter) deploymentSearchHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)

	deploymentItems := make([]bamboo.DeploymentItem, 0)
	searchTerm := request.URL.Query().Get("searchTerm")
	for _, value := range router.deployments {
		if strings.Contains(value.Name, searchTerm) {
			deploymentItems = append(deploymentItems, bamboo.DeploymentItem{
				Id:   strconv.Itoa(value.ID),
				Type: value.Name,
				SearchEntity: bamboo.DeploymentEntity{
					Id:          strconv.Itoa(value.ID),
					Key:         value.PlanKey.Key,
					ProjectName: value.Name,
					Description: value.Description,
				},
			})
		}
	}

	output, _ := json.Marshal(bamboo.DeploymentList{
		Start:     0,
		MaxResult: 100,
		Results:   deploymentItems,
	})
	_, _ = writer.Write(output)
}

func NewServiceVirtualization() *ServiceVirtualization {
	bambooRouter := &BambooRouter{
		deployments: make(map[string]bamboo.Deployment),
	}

	router := mux.NewRouter()
	router.HandleFunc("/rest/api/latest/search/deployment", bambooRouter.deploymentSearchHandler)
	return &ServiceVirtualization{
		router: router,
	}
}
