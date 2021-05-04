package dijcttest

import (
	"errors"
	"reflect"
	"testing"

	"github.com/wakuwaku3/dijct"
)

func Test_container_Invoke(t *testing.T) {
	t.Run("最後尾の引数がエラーで nil でない場合", func(t *testing.T) {
		sut := dijct.NewContainer()
		if err := sut.Register(NewService1With2WithError); err != nil {
			t.Fatal()
		}
		if err := sut.Invoke(func(service1 Service1) {}); err == nil || err.Error() != "NewService1With2WithError Error" {
			t.Fatal()
		}
	})
	t.Skip()

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
				t.Fatal()
			}
			if nestedService.GetName() != "nestedService" &&
				nestedService.GetID() != useCase.GetNestedService().GetID() {
				t.Fatal()
			}
			if service1.GetName() != "service1" &&
				service1.GetID() != useCase.GetService1().GetID() &&
				service1.GetID() != useCase.GetNestedService().GetService1().GetID() {
				t.Fatal()
			}
			if service2.GetName() != "service2" &&
				service2.GetID() != useCase.GetService2().GetID() &&
				service2.GetID() != useCase.GetNestedService().GetService2().GetID() {
				t.Fatal()
			}
			if service3.GetName() != "service3" &&
				service3.GetID() != useCase.GetService3().GetID() &&
				service3.GetID() != useCase.GetNestedService().GetService3().GetID() {
				t.Fatal()
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("LifetimeScope が InvokeManaged の場合 Invoke 毎に異なるインスタンスが生成されること", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()

		if err := sut.Register(NewService1); err != nil {
			t.Fatal(err)
		}
		service1ID := ""
		if err := sut.Invoke(func(service1 Service1) {
			service1ID = service1.GetID()
		}); err != nil {
			t.Fatal(err)
		}
		if err := sut.Invoke(func(service1 Service1) {
			if service1ID == service1.GetID() {
				t.Fatal()
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("LifetimeScope が ContainerManaged の場合 Invoke 時に同一のインスタンスが生成されること", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()

		if err := sut.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
			t.Fatal(err)
		}
		ifs := []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}
		if err := sut.Register(NewService3(), dijct.RegisterOptions{Interfaces: ifs}); err != nil {
			t.Fatal(err)
		}
		service2ID := ""
		service3ID := ""
		if err := sut.Invoke(func(service2 Service2, service3 Service3) {
			service2ID = service2.GetID()
			service3ID = service3.GetID()
		}); err != nil {
			t.Fatal(err)
		}
		if err := sut.Invoke(func(service2 Service2, service3 Service3) {
			if service2ID != service2.GetID() {
				t.Fatal(service2ID, service2.GetID())
			}
			if service3ID != service3.GetID() {
				t.Fatal(service3ID, service3.GetID())
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
				t.Fatal()
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
		if err == nil || err != dijct.ErrRequireFunction {
			t.Fatal(err)
		}
	})
	t.Run("解決するオブジェクトが存在しない", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()
		err := sut.Invoke(func() {})
		if err == nil || err != dijct.ErrNotFoundComponent {
			t.Fatal(err)
		}
	})
	t.Run("指定されたタイプを解決できない", func(t *testing.T) {
		t.Parallel()
		sut := dijct.NewContainer()
		err := sut.Invoke(func(service1 Service1) {})
		if err == nil || !dijct.IsErrInvalidResolveComponent(err) {
			t.Fatal(err)
		}
	})
	t.Run("Invoke した関数の引数", func(t *testing.T) {
		t.Run("error が返ってきた場合", func(t *testing.T) {
			sut := dijct.NewContainer()
			if err := sut.Register(NewService1); err != nil {
				t.Fatal()
			}
			e := errors.New("invoked function returning error")
			if err := sut.Invoke(func(service1 Service1) error {
				return e
			}); err != e {
				t.Fatal()
			}
		})
	})
	t.Run("コンストラクタの戻り値が複数の場合", func(t *testing.T) {
		t.Run("先頭の引数が解決される", func(t *testing.T) {
			sut := dijct.NewContainer()
			if err := sut.Register(NewService1With2); err != nil {
				t.Fatal()
			}
			if err := sut.Invoke(func(service1 Service1) {}); err != nil {
				t.Fatal()
			}
			if err := sut.Invoke(func(service2 Service2) {}); !dijct.IsErrInvalidResolveComponent(err) {
				t.Fatal()
			}
		})
		t.Run("最後尾の引数がエラーで nil でない場合 ()", func(t *testing.T) {
			sut := dijct.NewContainer()
			if err := sut.Register(NewService1With2WithError); err != nil {
				t.Fatal()
			}
			if err := sut.Invoke(func(service1 Service1) {}); err.Error() != "NewService1With2WithError Error" {
				t.Fatal()
			}
		})
		t.Run("最後尾の引数がエラーで nil でない場合(ContainerManaged)", func(t *testing.T) {
			sut := dijct.NewContainer()
			if err := sut.Register(NewService1With2WithError, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
				t.Fatal()
			}
			if err := sut.Invoke(func(service1 Service1) {}); err.Error() != "NewService1With2WithError Error" {
				t.Fatal()
			}
		})
	})
}
func Test_container_CreateChildContainer(t *testing.T) {
	t.Run("コンポーネントの Invoke の状態を継承すること", func(t *testing.T) {
		type result struct {
			ns NestedService
			s1 Service1
			s2 Service2
			s3 Service3
		}
		setupParent := func(t *testing.T, container dijct.Container) *result {
			if err := container.Register(NewNestedService, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
				t.Fatal(err)
			}
			if err := container.Register(NewService1); err != nil {
				t.Fatal(err)
			}
			if err := container.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
				t.Fatal(err)
			}
			if err := container.Register(NewService3(), dijct.RegisterOptions{Interfaces: []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}}); err != nil {
				t.Fatal(err)
			}
			var ns NestedService
			var s1 Service1
			var s2 Service2
			var s3 Service3

			if err := container.Invoke(func(nestedService NestedService, service1 Service1, service2 Service2, service3 Service3) {
				ns = nestedService
				s1 = service1
				s2 = service2
				s3 = service3
			}); err != nil {
				t.Fatal(err)
			}
			return &result{ns, s1, s2, s3}
		}
		chk := func(t *testing.T, sut dijct.Container, r *result) {
			if err := sut.Invoke(func(nestedService NestedService, service1 Service1, service2 Service2, service3 Service3) {
				if r.ns.GetID() != nestedService.GetID() {
					t.Fatal(r.ns.GetID(), r.s1.GetID(), r.s2.GetID(), r.s3.GetID(), nestedService.GetID(), service1.GetID(), service2.GetID(), service3.GetID())
				}
				if r.s1.GetID() == service1.GetID() || r.ns.GetService1().GetID() == service1.GetID() || nestedService.GetService1().GetID() == service1.GetID() {
					t.Fatal(r.s1.GetID(), service1.GetID(), r.ns.GetService1().GetID(), nestedService.GetService1().GetID())
				}
				if r.s2.GetID() != service2.GetID() || r.ns.GetService2().GetID() != service2.GetID() || nestedService.GetService2().GetID() != service2.GetID() {
					t.Fatal(r.ns.GetID(), r.s1.GetID(), r.s2.GetID(), r.s3.GetID(), nestedService.GetID(), service1.GetID(), service2.GetID(), service3.GetID())
				}
				if r.s3.GetID() != service3.GetID() || r.ns.GetService3().GetID() != service3.GetID() || nestedService.GetService3().GetID() != service3.GetID() {
					t.Fatal(r.ns.GetID(), r.s1.GetID(), r.s2.GetID(), r.s3.GetID(), nestedService.GetID(), service1.GetID(), service2.GetID(), service3.GetID())
				}
			}); err != nil {
				t.Fatal(err)
			}
		}
		t.Run("親コンテナで登録したコンポーネントを子コンテナでインスタンス生成できること", func(t *testing.T) {
			t.Parallel()
			container := dijct.NewContainer()
			r := setupParent(t, container)
			sut := container.CreateChildContainer()
			chk(t, sut, r)
		})
		t.Run("親コンテナで登録したコンポーネントを孫コンテナでインスタンス生成できること", func(t *testing.T) {
			t.Parallel()
			container := dijct.NewContainer()
			r := setupParent(t, container)
			sut := container.CreateChildContainer().CreateChildContainer()
			chk(t, sut, r)
		})
		t.Run("子コンテナで登録したコンポーネントを孫コンテナでインスタンス生成できること", func(t *testing.T) {
			t.Parallel()
			container := dijct.NewContainer().CreateChildContainer()
			r := setupParent(t, container)
			sut := container.CreateChildContainer()
			chk(t, sut, r)
		})
	})
	t.Run("コンポーネントの登録状態を継承すること", func(t *testing.T) {
		setupParent := func(t *testing.T, container dijct.Container) {
			if err := container.Register(NewNestedService, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
				t.Fatal(err)
			}
			if err := container.Register(NewService1); err != nil {
				t.Fatal(err)
			}
			if err := container.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
				t.Fatal(err)
			}
			if err := container.Register(NewService3(), dijct.RegisterOptions{Interfaces: []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}}); err != nil {
				t.Fatal(err)
			}
		}
		chk := func(t *testing.T, sut dijct.Container) {
			var ns NestedService
			var s1 Service1
			var s2 Service2
			var s3 Service3

			if err := sut.Invoke(func(nestedService NestedService, service1 Service1, service2 Service2, service3 Service3) {
				ns = nestedService
				s1 = service1
				s2 = service2
				s3 = service3
			}); err != nil {
				t.Fatal(err)
			}
			if err := sut.Invoke(func(nestedService NestedService, service1 Service1, service2 Service2, service3 Service3) {
				if ns.GetID() != nestedService.GetID() {
					t.Fatal(ns.GetID(), s1.GetID(), s2.GetID(), s3.GetID(), nestedService.GetID(), service1.GetID(), service2.GetID(), service3.GetID())
				}
				if s1.GetID() == service1.GetID() || ns.GetService1().GetID() == service1.GetID() || nestedService.GetService1().GetID() == service1.GetID() {
					t.Fatal(s1.GetID(), service1.GetID(), ns.GetService1().GetID(), nestedService.GetService1().GetID())
				}
				if s2.GetID() != service2.GetID() || ns.GetService2().GetID() != service2.GetID() || nestedService.GetService2().GetID() != service2.GetID() {
					t.Fatal(ns.GetID(), s1.GetID(), s2.GetID(), s3.GetID(), nestedService.GetID(), service1.GetID(), service2.GetID(), service3.GetID())
				}
				if s3.GetID() != service3.GetID() || ns.GetService3().GetID() != service3.GetID() || nestedService.GetService3().GetID() != service3.GetID() {
					t.Fatal()
				}
			}); err != nil {
				t.Fatal(err)
			}
			if err := sut.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
				if s1.GetID() == service1.GetID() {
					t.Fatal()
				}
				if s2.GetID() != service2.GetID() {
					t.Fatal()
				}
				if s3.GetID() != service3.GetID() {
					t.Fatal()
				}
			}); err != nil {
				t.Fatal(err)
			}
		}
		t.Run("親コンテナで登録したコンポーネントを子コンテナでインスタンス生成できること", func(t *testing.T) {
			t.Parallel()
			container := dijct.NewContainer()
			setupParent(t, container)
			sut := container.CreateChildContainer()
			chk(t, sut)
		})
		t.Run("親コンテナで登録したコンポーネントを孫コンテナでインスタンス生成できること", func(t *testing.T) {
			t.Parallel()
			container := dijct.NewContainer()
			setupParent(t, container)
			sut := container.CreateChildContainer().CreateChildContainer()
			chk(t, sut)
		})
		t.Run("子コンテナで登録したコンポーネントを孫コンテナでインスタンス生成できること", func(t *testing.T) {
			t.Parallel()
			container := dijct.NewContainer().CreateChildContainer()
			setupParent(t, container)
			sut := container.CreateChildContainer()
			chk(t, sut)
		})
	})
	t.Run("子コンテナで登録したコンポーネントを子コンテナでインスタンス生成できること", func(t *testing.T) {
		t.Parallel()
		container := dijct.NewContainer()

		sut := container.CreateChildContainer()
		if err := sut.Register(NewService1); err != nil {
			t.Fatal(err)
		}
		if err := sut.Register(NewService2, dijct.RegisterOptions{LifetimeScope: dijct.ContainerManaged}); err != nil {
			t.Fatal(err)
		}
		if err := sut.Register(NewService3(), dijct.RegisterOptions{Interfaces: []reflect.Type{reflect.TypeOf((*Service3)(nil)).Elem()}}); err != nil {
			t.Fatal(err)
		}

		var s1 Service1
		var s2 Service2
		var s3 Service3
		if err := sut.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
			s1 = service1
			s2 = service2
			s3 = service3
		}); err != nil {
			t.Fatal(err)
		}
		if err := sut.Invoke(func(service1 Service1, service2 Service2, service3 Service3) {
			if s1.GetID() == service1.GetID() {
				t.Fatal()
			}
			if s2.GetID() != service2.GetID() {
				t.Fatal()
			}
			if s3.GetID() != service3.GetID() {
				t.Fatal()
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("親コンテナで登録した内容を子コンテナで上書きできること", func(t *testing.T) {
		t.Parallel()
		container := dijct.NewContainer()
		if err := container.Register(NewService1(), dijct.RegisterOptions{Interfaces: []reflect.Type{reflect.TypeOf((*Service1)(nil)).Elem()}}); err != nil {
			t.Fatal(err)
		}

		sut := container.CreateChildContainer()
		if err := sut.Register(NewService2(), dijct.RegisterOptions{Interfaces: []reflect.Type{reflect.TypeOf((*Service1)(nil)).Elem()}}); err != nil {
			t.Fatal(err)
		}

		if err := sut.Invoke(func(service1 Service1) {
			if service1.GetName() != "service2" {
				t.Fatal()
			}
		}); err != nil {
			t.Fatal(err)
		}
	})
}
func Test_container_Register(t *testing.T) {
	t.Run("返り値がない関数を登録しようとした場合", func(t *testing.T) {
		sut := dijct.NewContainer()
		err := sut.Register(func() {
		})
		if err == nil || err != dijct.ErrRequireResponse {
			t.Fatal(err)
		}
	})
	t.Run("オプションは単一である必要があること", func(t *testing.T) {
		sut := dijct.NewContainer()
		opt1 := dijct.RegisterOptions{}
		opt2 := dijct.RegisterOptions{}
		err := sut.Register(func() string {
			return ""
		}, opt1, opt2)
		if err == nil || err != dijct.ErrNoMultipleOption {
			t.Fatal(err)
		}
	})
	t.Run("ポインタを登録する場合は、インターフェイスを指定する必要があること", func(t *testing.T) {
		sut := dijct.NewContainer()
		err := sut.Register(NewService3())
		if err == nil || err != dijct.ErrNeedInterfaceOnPointerRegistering {
			t.Fatal(err)
		}
	})
}
func Test_container_Verify(t *testing.T) {
	t.Run("Verify できること1", func(t *testing.T) {
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
		if err := sut.Verify(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Verify できること2", func(t *testing.T) {
		sut := dijct.NewContainer()
		if err := sut.Verify(); err == nil || err != dijct.ErrNotFoundComponent {
			t.Fatal(err)
		}
	})
	t.Run("Verify できること3", func(t *testing.T) {
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

		if err := sut.Verify(); err == nil || !dijct.IsErrInvalidResolveComponent(err) {
			t.Fatal(err)
		}
	})
}
