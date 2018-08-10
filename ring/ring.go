package ring

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"loadbalance/types"

	"github.com/dongjialong2006/log"
)

type LoadBalanceRoundRobinRing struct {
	sync.RWMutex
	ctx  context.Context
	log  *log.Entry
	name string
	conn *ring
	stop chan struct{}
}

func New(ctx context.Context, name string) *LoadBalanceRoundRobinRing {
	return &LoadBalanceRoundRobinRing{
		ctx:  ctx,
		log:  log.New("loadbalance/heap"),
		name: name,
		stop: make(chan struct{}),
	}
}

func (robin *LoadBalanceRoundRobinRing) next() *ring {
	robin.Lock()
	conn := robin.conn
	if nil == conn || (nil == conn.Conn && nil == conn.next) {
		robin.Unlock()
		return nil
	}

	if nil == conn.Conn || !conn.Conn.Enable() {
		robin.conn = conn.next
		robin.Unlock()
		return nil
	}

	robin.conn = conn.next
	robin.Unlock()

	return conn
}

func (robin *LoadBalanceRoundRobinRing) Get(ctx context.Context) (types.Conn, error) {
	tick := time.Tick(time.Second * 3)
	deadline, ok := ctx.Deadline()
	for {
		if ok {
			if deadline.Before(time.Now()) {
				break
			}
		}

		if conn := robin.next(); nil != conn {
			return conn.Conn, nil
		}
		select {
		case <-robin.ctx.Done():
			return nil, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-tick:
			return nil, fmt.Errorf("timeout.")
		default:
		}
	}

	return nil, errors.New("loadbalance get addr timeout.")
}

func (robin *LoadBalanceRoundRobinRing) Stop() {
	if nil != robin.stop {
		close(robin.stop)
		robin.stop = nil
	}
}

func (robin *LoadBalanceRoundRobinRing) Start() error {
	return nil
}

func (robin *LoadBalanceRoundRobinRing) Add(conns ...types.Conn) error {
	for _, conn := range conns {
		if nil == conn {
			continue
		}

		robin.Lock()
		if nil == robin.conn {
			robin.conn = new(conn)
		} else {
			robin.conn.Link(new(conn))
		}
		robin.Unlock()

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
