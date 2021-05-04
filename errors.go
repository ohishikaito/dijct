package dijct

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrNoMultipleOption                  = fmt.Errorf("オプションは単一である必要があります")
	ErrNeedInterfaceOnPointerRegistering = fmt.Errorf("ポインタを登録する場合は、インターフェイスを指定する必要があります")
	ErrRequireFunction                   = fmt.Errorf("関数を指定してください")
	ErrNotFoundComponent                 = fmt.Errorf("解決するオブジェクトが存在しません")
	ErrNeedSingleResponseConstructor     = fmt.Errorf("コンストラクタの戻り値は単一である必要があります")
)

func newErrInvalidResolveComponent(t reflect.Type) error {
	return fmt.Errorf("指定されたタイプを解決できません。(%v)", t)
}
func IsErrInvalidResolveComponent(err error) bool {
	return strings.HasPrefix(err.Error(), "指定されたタイプを解決できません。")
}
