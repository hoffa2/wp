package wp

import "testing"

type test struct{}

func TestInit(t *testing.T) {
	assertChan := make(chan int, 10000)
	tests := 10000
	p := NewPool(100, nil, func(args interface{}) error {
		s := args.(chan int)
		s <- 0
		return nil
	})
	p.Start()
	for i := 0; i < 10000; i++ {
		p.Add(assertChan)
	}
	p.Wait()
	receives := 0
	done := false
	for {
		select {
		case <-assertChan:
			receives++
		default:
			done = true
			break
		}
		if done {
			break
		}
	}
	if receives != tests {
		t.Errorf("Should have %d - got %d", tests, receives)
	}
	p.Quit()
}

func TestConcurrentAdd(t *testing.T) {
	assertChan := make(chan int, 10000)
	tests := 10000
	p := NewPool(100, nil, func(args interface{}) error {
		s := args.(chan int)
		s <- 0
		return nil
	})

	p.Start()

	for i := 0; i < 10000; i++ {
		go func(i int) {
			p.Add(assertChan)
		}(i)
	}
	p.Wait()

	receives := 0
	done := false
	for {
		select {
		case <-assertChan:
			receives++
		default:
			done = true
			break
		}
		if done {
			break
		}
	}
	if receives != tests {
		t.Errorf("Should have %d - got %d", tests, receives)
	}
	p.Quit()

}
