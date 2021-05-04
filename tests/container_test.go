package dijcttest

import (
	"reflect"
	"testing"

	"github.com/wakuwaku3/dijct"
)

func Test_container_Invoke(t *testing.T) {
	t.Run("1回の Invoke で生成されるオブジェクトは登録された型ごとに一意であること", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()
		if err := sut.Register(NewUseCase); err != nil {
			t.Fatal(err)
		}
		if err := sut.Register(NewNestedService); err != nil {
			t.Fatal(err)
		}
		if err := sut.Register(NewService1); err != nil {
			t.Fatal(err)
		}
		if err := sut.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
			t.Fatal(err)
		}

		ifs := []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}
		if err := sut.Register(NewService3(), dijct.RegisterOptions{Interfaces: ifs}); err != nil {
			t.Fatal(err)
		}
		if err := sut.Invoke(func(
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
		}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("LifetimeScope が ContainerManaged の場合 Invoke 毎に異なるインスタンスが生成されること", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()

		if err := sut.Register(NewService1); err != nil {
			t.Fatal(err)
		}
		if err := sut.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
			t.Fatal(err)
		}
		ifs := []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}
		if err := sut.Register(NewService3(), dijct.RegisterOptions{Interfaces: ifs}); err != nil {
			t.Fatal(err)
		}
		service1ID := ""
		service2ID := ""
		service3ID := ""
		if err := sut.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
			service1ID = service1.GetID()
			service2ID = service2.GetID()
			service3ID = service3.GetID()
		}); err != nil {
			t.Fatal(err)
		}
		if err := sut.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
			if service1ID == service1.GetID() {
				t.FailNow()
			}
			if service2ID != service2.GetID() {
				t.FailNow()
			}
			if service3ID != service3.GetID() {
				t.FailNow()
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("コンテナインスタンス自身を自己解決できること", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()
		err := sut.Invoke(func(
			currentContainer dijct.Container,
			ioCContainer dijct.IoCContainer,
			serviceLocator dijct.ServiceLocator,
		) {
			if sut != currentContainer ||
				sut != ioCContainer ||
				sut != serviceLocator {
				t.FailNow()
			}
		})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("関数以外が指定された", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()
		err := sut.Invoke("")
		if err == nil || err.Error() != "関数を指定してください" {
			t.Fatal(err)
		}
	})
	t.Run("解決するオブジェクトが存在しない", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()
		err := sut.Invoke(func() {})
		if err == nil || err.Error() != "解決するオブジェクトが存在しません" {
			t.Fatal(err)
		}
	})
	t.Run("指定されたタイプを解決できない", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()
		err := sut.Invoke(func(service1 Service1) {})
		if err == nil || err.Error() != "指定されたタイプを解決できません。(dijcttest.Service1)" {
			t.Fatal(err)
		}
	})
}

func Test_container_CreateChildContainer(t *testing.T) {
	t.Run("親コンテナで登録したコンポーネントを子コンテナでインスタンス生成できること", func(t *testing.T) {
		t.Parallel()
		container := dijct.NewContainer()

		if err := container.Register(NewService1); err != nil {
			t.Fatal(err)
		}

		if err := container.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
			t.Fatal(err)
		}
		var s1 Service1
		var s2 Service2

		if err := container.Invoke(func(service1 Service1, service2 Service2) {
			s1 = service1
			s2 = service2
		}); err != nil {
			t.Fatal(err)
		}

		sut := container.CreateChildContainer()

		if err := sut.Invoke(func(service1 Service1, service2 Service2) {
			if s1.GetID() == service1.GetID() {
				t.FailNow()
			}
			if s2.GetID() != service2.GetID() {
				t.FailNow()
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("子コンテナで登録したコンポーネントを子コンテナでインスタンス生成できること", func(t *testing.T) {
		t.Parallel()
		container := dijct.NewContainer()

		sut := container.CreateChildContainer()

		if err := sut.Register(NewService3); err != nil {
			t.Fatal(err)
		}

		if err := sut.Invoke(func(service3 Service3) {
			if service3.GetName() != "service3" {
				t.FailNow()
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
}
func Test_container_Register(t *testing.T) {
	t.Run("コンストラクタの戻り値は単一である必要があること", func(t *testing.T) {
		sut := dijct.NewContainer()
		err := sut.Register(func() (string, string) {
			return "", ""
		})
		if err == nil || err.Error() != "コンストラクタの戻り値は単一である必要があります" {
			t.FailNow()
		}
	})
	t.Run("オプションは単一である必要があること", func(t *testing.T) {
		sut := dijct.NewContainer()
		opt1 := dijct.RegisterOptions{}
		opt2 := dijct.RegisterOptions{}
		err := sut.Register(func() string {
			return ""
		}, opt1, opt2)
		if err == nil || err.Error() != "オプションは単一である必要があります" {
			t.Fatal(err)
		}
	})
	t.Run("ポインタを登録する場合は、インターフェイスを指定する必要があること", func(t *testing.T) {
		sut := dijct.NewContainer()
		err := sut.Register(NewService3())
		if err == nil || err.Error() != "ポインタを登録する場合は、インターフェイスを指定する必要があります" {
			t.Fatal(err)
		}
	})
}
