# dijct

Simple DI Container

## Required

go(v1.11 latest)

## Command

#### Build

```sh
go build -o dist/dijct
```

#### Test

```sh
go test ./tests/ -test.v
```

#### Release

```sh
git tag 1.0.0
git push origin --tags
```

## Installation

```sh
go get github.com/tampopos/dijct
```

## Usage

### Basic

```go
package main
import (
	"fmt"
	"os"

	"github.com/tampopos/dijct"
)
type (
	service1 struct {
		id   string
		name string
	}
	Service1 interface {
		GetID() string
		GetName() string
	}
)
func NewService1() Service1 {
	return &service1{id: uuid.New().String(), name: "service1"}
}
func (service1 *service1) GetID() string {
	return service1.id
}
func (service1 *service1) GetName() string {
	return service1.name
}
func main() {
    // Create container
    container := dijct.NewContainer()

    // Register service
	err := container.Register(NewService1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

    // Invoke service
	err = container.Invoke(func(
		service1 Service1,
	) {
        // service1 is auto resolved by container.
		fmt.Printf("Invoke %v %v\n", service1.GetName(), service1.GetID())
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
```

### Advance

#### Register

```go
// Register
container.Register(NewService1)

// Register as singleton
container.Register(NewService1, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged})

// Register const value as singleton
ifs := []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}
container.Register(NewService3(), dijct.RegisterOptions{Interfaces: ifs})
```

#### Invoke

```go
container := dijct.NewContainer()
container.Register(NewService1)
container.Invoke(func(
    service1 Service1,
    currentContainer dijct.Container,
    ioCContainer dijct.IoCContainer,
    serviceLocator dijct.ServiceLocator,
) {
    // service1 is auto resolved by container.
    // dijct.Container, dijct.IoCContainer, dijct.ServiceLocator are auto resolved without register. These are equal to container.
})
container.Invoke(func(
    service1 Service1,
) {
    // LifetimeScope of service1 is InvokeManaged by default.
    // In this case service1 will be recreated by reinvoke.
})

```

#### ChildContainer

```go
container := dijct.NewContainer()
container.Register(NewService1)
childContainer := container.CreateChildContainer()
childContainer.Register(NewService2)
childContainer.Invoke(func(
    service1 Service1,
    service2 Service2,
    currentContainer dijct.Container,
    ioCContainer dijct.IoCContainer,
    serviceLocator dijct.ServiceLocator,
) {
    // service1 is auto resolved by container.
    // service2 is auto resolved by childContainer.
    // currentContainer, ioCContainer, serviceLocator are equal to childContainer.
})
```
