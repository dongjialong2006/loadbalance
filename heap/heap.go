package heap

import (
	"context"
	"fmt"
	"time"

	"loadbalance/types"

	"github.com/dongjialong2006/log"
)

type LoadBalanceHeap struct {
	ctx  context.Context
	log  *log.Entry
	name string
	heap *types.Heap
	stop chan struct{}
}

func New(ctx context.Context, name string, max int) *LoadBalanceHeap {
	return &LoadBalanceHeap{
		ctx:  ctx,
		log:  log.New("loadbalance/heap"),
		name: name,
		heap: types.NewHeap(max),
		stop: make(chan struct{}),
	}
}

func (hp *LoadBalanceHeap) Start() error {
	hp.heap.HeapInit()
	return nil
}

func (hp *LoadBalanceHeap) Stop() {
	if nil != hp.stop {
		close(hp.stop)
		hp.stop = nil
	}

	if nil != hp.heap {
		hp.heap = nil
	}
}

func (hp *LoadBalanceHeap) Add(conns ...types.Conn) error {
	for _, conn := range conns {
		if nil == conn {
			continue
		}
		hp.heap.Push(conn)

		go func() {
			var err error = nil
			tick := time.Tick(time.Second * 3)
			for {
				select {
				case <-hp.ctx.Done():
					return
				case <-hp.stop:
					return
				case <-tick:
					if err = conn.CheckHealth(); nil != err {
						hp.log.Error(err)
					}
				}
			}
		}()
	}

	return nil
}

func (hp *LoadBalanceHeap) Get(ctx context.Context) (types.Conn, error) {
	tick := time.Tick(time.Second * 3)
	deadline, ok := ctx.Deadline()

	for {
		if ok {
			if deadline.Before(time.Now()) {
				break
			}
		}

		if conn, err := hp.get(); nil == err {
			return conn, nil
		}

		select {
		case <-hp.stop:
			return nil, nil
		case <-hp.ctx.Done():
			return nil, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-tick:
			return nil, fmt.Errorf("timeout.")
		default:
		}
	}

	return nil, fmt.Errorf("loadbalance get addr timeout.")
}

func (hp *LoadBalanceHeap) get() (types.Conn, error) {
	conn := hp.heap.Pop()
	if nil == conn {
		return nil, fmt.Errorf("no connection is available.")
	}

	return conn, nil
}
