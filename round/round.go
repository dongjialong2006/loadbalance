package round

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"loadbalance/types"

	"github.com/dongjialong2006/log"
)

type LoadBalanceRoundRobin struct {
	sync.RWMutex
	ctx   context.Context
	log   *log.Entry
	name  string
	count int64
	point int64
	stop  chan struct{}
	conns []types.Conn
}

func New(ctx context.Context, name string) *LoadBalanceRoundRobin {
	return &LoadBalanceRoundRobin{
		ctx:  ctx,
		log:  log.New("loadbalance/roundrobin"),
		name: name,
		stop: make(chan struct{}),
	}
}

func (l *LoadBalanceRoundRobin) get() (types.Conn, error) {
	if 0 == len(l.conns) {
		return nil, fmt.Errorf("conns is empty.")
	}

	l.RLock()
	defer l.RUnlock()
	n := atomic.AddInt64(&l.point, 1) % l.count
	if l.conns[n].Enable() {
		return l.conns[n], nil
	}

	return nil, fmt.Errorf("no connection is available.")
}

func (l *LoadBalanceRoundRobin) Start() error {
	return nil
}

func (robin *LoadBalanceRoundRobin) Add(conns ...types.Conn) error {
	for _, conn := range conns {
		if nil == conn || !conn.Enable() {
			continue
		}

		robin.Lock()
		robin.conns = append(robin.conns, conn)
		robin.Unlock()
		atomic.AddInt64(&robin.count, 1)

		go func() {
			var err error = nil
			tick := time.Tick(time.Second * 3)
			for {
				select {
				case <-robin.ctx.Done():
					return
				case <-robin.stop:
					return
				case <-tick:
					if err = conn.CheckHealth(); nil != err {
						robin.log.Error(err)
					}
				}
			}
		}()
	}
	return nil
}

func (l *LoadBalanceRoundRobin) Get(ctx context.Context) (types.Conn, error) {
	deadline, ok := ctx.Deadline()
	for {
		if ok {
			if deadline.Before(time.Now()) {
				break
			}
		}
		unit, err := l.get()
		if nil == err {
			return unit, nil
		}

		select {
		case <-l.ctx.Done():
			return nil, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}

	return nil, fmt.Errorf("loadbalance get addr timeout")
}

func (l *LoadBalanceRoundRobin) Stop() {
	l.Lock()
	defer l.Unlock()
	if nil != l.stop {
		close(l.stop)
		l.stop = nil
	}
}
