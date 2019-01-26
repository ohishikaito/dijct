package dijcttest

import "github.com/google/uuid"

type (
	useCase struct {
		id            string
		name          string
		nestedService NestedService
		service1      Service1
	}
	// UseCase is
	UseCase interface {
		GetID() string
		GetName() string
		GetNestedService() NestedService
		GetService1() Service1
	}
	nestedService struct {
		id       string
		name     string
		service1 Service1
		service2 Service2
	}
	// NestedService is
	NestedService interface {
		GetID() string
		GetName() string
		GetService1() Service1
		GetService2() Service2
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
)

// NewUseCase is
func NewUseCase(nestedService NestedService, service1 Service1) UseCase {
	return &useCase{id: uuid.New().String(), name: "useCase", nestedService: nestedService, service1: service1}
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

// NewNestedService is
func NewNestedService(service1 Service1, service2 Service2) NestedService {
	return &nestedService{id: uuid.New().String(), name: "nestedService", service1: service1, service2: service2}
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
