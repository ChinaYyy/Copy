package pool

import (
	"fmt"
	"testing"
)

func TestGoPool(t *testing.T) {
	pool := NewGoPool(WithMaxLimit(3))
	defer pool.Wait()

	for i := 0; i <= 10; i++ {
		pool.Submit(func(args ...interface{}) {
			num := args[0].(int)
			fmt.Println(num, num*2)
		}, i)
	}
}
