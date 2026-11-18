package workflow

import (
	"testing"
)

func TestExecLocker_Local(t *testing.T) {
	l := NewExecLocker(nil)
	u1, ok1 := l.TryLock("e1")
	if !ok1 {
		t.Fatal("first lock should succeed")
	}
	_, ok2 := l.TryLock("e1")
	if ok2 {
		t.Fatal("second lock should fail")
	}
	u1()
	u3, ok3 := l.TryLock("e1")
	if !ok3 {
		t.Fatal("lock after unlock should succeed")
	}
	u3()
}
