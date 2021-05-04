package dijcttest

import (
	"errors"

	"github.com/google/uuid"
)

type (
	useCase struct {
		id            string
		name          string
		nestedService NestedService
		service1      Service1
		service2      Service2
		service3      Service3
	}
	// UseCase is
	UseCase interface {
		GetID() string
		GetName() string
		GetNestedService() NestedService
		GetService1() Service1
		GetService2() Service2
		GetService3() Service3
	}
	nestedService struct {
		id       string
		name     string
		service1 Service1
		service2 Service2
		service3 Service3
	}
	// NestedService is
	NestedService interface {
		GetID() string
		GetName() string
		GetService1() Service1
		GetService2() Service2
		GetService3() Service3
	}
	service1 struct {
		id   string
		name string
	}
	// Service1 is
	Service1 interface {
		GetID() string
		GetName() string
	}
	service2 struct {
		id   string
		name string
	}
	// Service2 is
	Service2 interface {
		GetID() string
		GetName() string
	}
	service3 struct {
		id   string
		name string
	}
	// Service3 is
	Service3 interface {
		GetID() string
		GetName() string
	}
)

// NewUseCase is
func NewUseCase(nestedService NestedService, service1 Service1, service2 Service2, service3 Service3) UseCase {
	return &useCase{
		id:            uuid.New().String(),
		name:          "useCase",
		nestedService: nestedService,
		service1:      service1,
		service2:      service2,
		service3:      service3,
	}
}

// GetID is
func (useCase *useCase) GetID() string {
	return useCase.id
}

// GetName is
func (useCase *useCase) GetName() string {
	return useCase.name
}

// GetNestedService is
func (useCase *useCase) GetNestedService() NestedService {
	return useCase.nestedService
}

// GetService1 is
func (useCase *useCase) GetService1() Service1 {
	return useCase.service1
}

// GetService2 is
func (useCase *useCase) GetService2() Service2 {
	return useCase.service2
}

// GetService3 is
func (useCase *useCase) GetService3() Service3 {
	return useCase.service3
}

// NewNestedService is
func NewNestedService(service1 Service1, service2 Service2, service3 Service3) NestedService {
	return &nestedService{
		id:       uuid.New().String(),
		name:     "nestedService",
		service1: service1,
		service2: service2,
		service3: service3,
	}
}

// GetID is
func (nestedService *nestedService) GetID() string {
	return nestedService.id
}

// GetName is
func (nestedService *nestedService) GetName() string {
	return nestedService.name
}

// GetService1 is
func (nestedService *nestedService) GetService1() Service1 {
	return nestedService.service1
}

// GetService2 is
func (nestedService *nestedService) GetService2() Service2 {
	return nestedService.service2
}

// GetService3 is
func (nestedService *nestedService) GetService3() Service3 {
	return nestedService.service3
}

// NewService1 is
func NewService1() Service1 {
	return &service1{id: uuid.New().String(), name: "service1"}
}

// GetID is
func (service1 *service1) GetID() string {
	return service1.id
}

// GetName is
func (service1 *service1) GetName() string {
	return service1.name
}

// NewService2 is
func NewService2() Service2 {
	return &service2{id: uuid.New().String(), name: "service2"}
}

// GetID is
func (service2 *service2) GetID() string {
	return service2.id
}

// GetName is
func (service2 *service2) GetName() string {
	return service2.name
}

// NewService3 is
func NewService3() Service3 {
	return &service3{id: uuid.New().String(), name: "service3"}
}

// GetID is
func (service3 *service3) GetID() string {
	return service3.id
}

// GetName is
func (service3 *service3) GetName() string {
	return service3.name
}

func NewService1With2() (Service1, Service2) {
	return &service1{id: uuid.New().String(), name: "service1"}, &service2{id: uuid.New().String(), name: "service2"}
}
func NewService1With2WithError() (Service1, Service2, error) {
	return &service1{id: uuid.New().String(), name: "service1"}, &service2{id: uuid.New().String(), name: "service2"}, errors.New("NewService1With2WithError Error")
}
