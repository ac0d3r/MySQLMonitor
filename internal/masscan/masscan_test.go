package masscan

import (
	"sync"
	"testing"
)

func TestMasscanStart(t *testing.T) {
	m := New("masscan", "80,443", "10.10.40.191")
	out, err := m.Start()
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for o := range out {
			t.Log(o)
		}
		wg.Done()
	}()
	if err := m.Wait(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
