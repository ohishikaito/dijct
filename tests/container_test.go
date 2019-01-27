package dijcttest

import (
	"reflect"
	"testing"

	"github.com/tampopos/dijct"
)

func TestOnContainerResolve(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Register(NewUseCase)
	if err != nil {
		t.Fatal(err)
	}
	err = container.Register(NewNestedService)
	if err != nil {
		t.Fatal(err)
	}
	err = container.Register(NewService1)
	if err != nil {
		t.Fatal(err)
	}
	err = container.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged})
	if err != nil {
		t.Fatal(err)
	}
	ifs := []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}
	err = container.Register(NewService3(), dijct.RegisterOptions{Interfaces: ifs})
	if err != nil {
		t.Fatal(err)
	}
	service1ID := ""
	service2ID := ""
	service3ID := ""
	err = container.Invoke(func(
		useCase UseCase,
		nestedService NestedService,
		service1 Service1,
		service2 Service2,
		service3 Service3,
	) {
		if useCase.GetName() != "useCase" {
			t.FailNow()
		}
		if nestedService.GetName() != "nestedService" &&
			nestedService.GetID() != useCase.GetNestedService().GetID() {
			t.FailNow()
		}
		if service1.GetName() != "service1" &&
			service1.GetID() != useCase.GetService1().GetID() &&
			service1.GetID() != useCase.GetNestedService().GetService1().GetID() {
			t.FailNow()
		}
		if service2.GetName() != "service2" &&
			service2.GetID() != useCase.GetService2().GetID() &&
			service2.GetID() != useCase.GetNestedService().GetService2().GetID() {
			t.FailNow()
		}
		if service3.GetName() != "service3" &&
			service3.GetID() != useCase.GetService3().GetID() &&
			service3.GetID() != useCase.GetNestedService().GetService3().GetID() {
			t.FailNow()
		}
		service1ID = service1.GetID()
		service2ID = service2.GetID()
		service3ID = service3.GetID()
	})
	if err != nil {
		t.Fatal(err)
	}
	err = container.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
		if service1ID == service1.GetID() {
			t.FailNow()
		}
		if service2ID != service2.GetID() {
			t.FailNow()
		}
		if service3ID != service3.GetID() {
			t.FailNow()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestOnContainerLifetime(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Register(NewService1)
	if err != nil {
		t.Fatal(err)
	}
	err = container.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged})
	if err != nil {
		t.Fatal(err)
	}
	ifs := []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}
	err = container.Register(NewService3(), dijct.RegisterOptions{Interfaces: ifs})
	if err != nil {
		t.Fatal(err)
	}
	service1ID := ""
	service2ID := ""
	service3ID := ""
	err = container.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
		service1ID = service1.GetID()
		service2ID = service2.GetID()
		service3ID = service3.GetID()
	})
	if err != nil {
		t.Fatal(err)
	}
	err = container.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
		if service1ID == service1.GetID() {
			t.FailNow()
		}
		if service2ID != service2.GetID() {
			t.FailNow()
		}
		if service3ID != service3.GetID() {
			t.FailNow()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestOnChildContainer(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Register(NewService1)
	if err != nil {
		t.Fatal(err)
	}
	err = container.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged})
	if err != nil {
		t.Fatal(err)
	}
	var s1 Service1
	var s2 Service2
	err = container.Invoke(func(service1 Service1, service2 Service2) {
		s1 = service1
		s2 = service2
	})
	if err != nil {
		t.Fatal(err)
	}

	childContainer := container.CreateChildContainer()
	err = childContainer.Register(NewService3)
	if err != nil {
		t.Fatal(err)
	}

	err = childContainer.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
		if s1.GetID() == service1.GetID() {
			t.FailNow()
		}
		if s2.GetID() != service2.GetID() {
			t.FailNow()
		}
		if service3.GetName() != "service3" {
			t.FailNow()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestOnContainerResolved(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Invoke(func(
		currentContainer dijct.Container,
		ioCContainer dijct.IoCContainer,
		serviceLocator dijct.ServiceLocator,
	) {
		if container != currentContainer ||
			container != ioCContainer ||
			container != serviceLocator {
			t.FailNow()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestOnRegisterError1(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Register(func() (string, string) {
		return "", ""
	})
	if err == nil || err.Error() != "コンストラクタの戻り値は単一である必要があります" {
		t.FailNow()
	}
}
func TestOnRegisterError2(t *testing.T) {
	container := dijct.NewContainer()
	opt1 := dijct.RegisterOptions{}
	opt2 := dijct.RegisterOptions{}
	err := container.Register(func() string {
		return ""
	}, opt1, opt2)
	if err == nil || err.Error() != "オプションは単一である必要があります" {
		t.Fatal(err)
	}
}
func TestOnRegisterError3(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Register(NewService3())
	if err == nil || err.Error() != "ポインタを登録する場合は、インターフェイスを指定する必要があります" {
		t.Fatal(err)
	}
}
func TestOnInvokeError1(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Invoke("")
	if err == nil || err.Error() != "関数を指定してください" {
		t.Fatal(err)
	}
}
func TestOnInvokeError2(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Invoke(func() {})
	if err == nil || err.Error() != "解決するオブジェクトが存在しません" {
		t.Fatal(err)
	}
}
func TestOnInvokeError3(t *testing.T) {
	container := dijct.NewContainer()
	err := container.Invoke(func(service1 Service1) {})
	if err == nil || err.Error() != "指定されたタイプを解決できません。(dijcttest.Service1)" {
		t.Fatal(err)
	}
}
