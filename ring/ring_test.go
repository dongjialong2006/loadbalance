package ring

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRing(t *testing.T) {
	var cancel context.CancelFunc
	updateCh := make(chan types.Status, 100)
	ctx := context.Background()
	var temp = make([]*types.ConnStatus, 100)
	lb := New(ctx, updateCh)
	for i := 0; i < 100; i++ {
		temp[i] = types.NewConnStatus(fmt.Sprintf("10.10.121.%2d", i))
		updateCh <- temp[i]
	}
	time.Sleep(time.Second * 5)
	t.Parallel()

	_, ok := ctx.Deadline()
	if !ok {
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second*60))
		if nil != cancel {
			defer cancel()
		}
	}

	var wg sync.WaitGroup
	for j := 0; j < 100; j++ {
		go func() {
			wg.Add(1)
			defer wg.Done()
			for i := 0; i < 100000; i++ {
				start := time.Now()
				addr, _, err := lb.Get(ctx)
				stop := time.Now().Sub(start)
				if nil != err {
					fmt.Println("dongcf------", i, "--------", err)
					return
				}
				fmt.Println(fmt.Sprintf("dongcf-------%5d---%s---%v", i, addr, stop))
			}
		}()
	}

	go func() {
		for i := 15; i < 50; i++ {
			switch i % 3 {
			case 0:
				temp[i].SetStatus(types.TCPConnError)
			case 1:
				temp[i].SetStatus(types.TCPConnClosed)
			case 2:
				temp[i].SetStatus(types.TCPConnUnknownError)
			}
			temp[i].SetOpType(types.TransportUpdate)
			updateCh <- temp[i]
		}
	}()

	wg.Wait()
	lb.Dump()
	return
}
