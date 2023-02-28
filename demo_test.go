package cache

import (
	"fmt"
	"testing"
)

func TestDemo(t *testing.T) {
	dc := demoCache[interface{}]{make(map[string]interface{})}
	dc.Set("a", 1)
	expect := dc.Get("a")
	fmt.Println(expect)
	if 1 != expect {
		t.Error("fuck you")
	}
}
