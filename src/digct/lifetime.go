package dijct

// LifetimeScope はインスタンスのライフタイムスコープです
type LifetimeScope int

const (
	// ContainerManaged の場合、そのコンテナ及び派生したコンテナでインスタンスは一意です
	ContainerManaged LifetimeScope = iota
	// InvokeManaged の場合、その呼び出し内でインスタンスは一意です
	InvokeManaged
)
