package mock

import (
	"testing"
)

func TestMockRegister(t *testing.T) {
	Mock(func(i int) {
	})
}
