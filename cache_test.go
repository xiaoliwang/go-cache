package cache

import "testing"

func TestCache(t *testing.T) {
    tc := New(DefaultExpiration, 0)
    a, found := tc.Get("a")
    if found || a != nil {
        t.Error("Getting A found value that shouldn't exist:", a)
    }

    tc.Set("a", 1, DefaultExpiration)

    x, found := tc.Get("a")
    if !found {
        t.Error("a was not found while getting a2")
    }
    if x == nil {
        t.Error("x for a is nil")
    } else if a2 := x.(int); a2+2 != 3 {
        t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
    }
}
