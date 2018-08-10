package loadbalance

import (
	"context"

	"loadbalance/heap"
	"loadbalance/ring"
	"loadbalance/round"
	"loadbalance/types"
)

// LoadBalance 负责负载均衡实现
type LoadBalance interface {
	Start() error

	Stop()

	Add(conns ...types.Conn) error

	Get(ctx context.Context) (types.Conn, error)
}

func NewHeap(ctx context.Context, name string, max int) LoadBalance {
	return heap.New(ctx, name, max)
}

func NewRoundRobin(ctx context.Context, name string) LoadBalance {
	return round.New(ctx, name)
}

func NewRoundRobinRing(ctx context.Context, name string) LoadBalance {
	return ring.New(ctx, name)
}
