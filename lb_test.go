package loadbalance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"loadbalance/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backend", func() {
	Specify("[static-hosts] + [http] ", func() {
		addCh := make(chan *types.LoadBalanceUnit, 16)
		updateCh := make(chan bool, 16)
		ctx := context.Background()
		b := New(ctx, addCh, updateCh)
		Expect(b).ShouldNot(BeNil())
		defer b.Stop()

		for i := 0; i < 50; i++ {
			addCh <- types.New(tranTypes.NewTransportStatusOptions(ctx, fmt.Sprintf("10.4.110.%2d", i)))
		}

		temp1 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.50")
		addCh <- types.New(temp1)
		temp2 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.51")
		addCh <- types.New(temp2)
		temp3 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.52")
		addCh <- types.New(temp3)
		temp4 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.53")
		addCh <- types.New(temp4)

		temp5 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.54")
		addCh <- types.New(temp5)
		temp6 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.55")
		addCh <- types.New(temp6)
		temp7 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.56")
		addCh <- types.New(temp7)
		temp8 := tranTypes.NewTransportStatusOptions(ctx, "10.4.110.57")
		addCh <- types.New(temp8)
		// ch <- types.New(tranTypes.NewTransportStatusOptions(ctx, "10.4.110.58"))
		// ch <- types.New(tranTypes.NewTransportStatusOptions(ctx, "10.4.110.59"))

		time.Sleep(3 * time.Second)

		b.Dump()

		var wg sync.WaitGroup
		var sub time.Duration
		var start time.Time
		var stop time.Time
		var nums []int = make([]int, 6)
		for j := 0; j < 100; j++ {
			wg.Add(1)
			go func(v int) {
				for i := 0; i < 10000; i++ {
					start = time.Now()
					addr, done, err := b.Get(ctx)
					if nil != err {
						fmt.Println(err)
						time.Sleep(time.Second)
						continue
					}
					if addr >= "10.4.110.54" {
						fmt.Println("error addr:", addr)
						wg.Done()
						return
					}
					done()
					stop = time.Now()
					sub = stop.Sub(start)
					if sub > 30*time.Microsecond {
						nums[5]++
						if nums[5] == 1 {
							temp6.SetStatus(tranTypes.TCPConnError)
						}
					} else if sub > 25*time.Microsecond {
						nums[4]++
						if nums[4] == 1 {
							temp7.SetStatus(tranTypes.TCPConnError)
						}
					} else if sub > 20*time.Microsecond {
						nums[3]++
						if nums[3] == 1 {
							temp4.SetStatus(tranTypes.TCPConnClosed)
						}
					} else if sub > 15*time.Microsecond {
						nums[2]++
						if nums[2] == 3 {
							temp1.SetStatus(tranTypes.TCPConnClosed)
						}
						if nums[2] == 5 {
							temp5.SetStatus(tranTypes.TCPConnError)
						}

					} else if sub > 10*time.Microsecond {
						nums[1]++
						if nums[1] == 5 {
							temp2.SetStatus(tranTypes.TCPConnClosed)
						}
						if nums[1] == 2 {
							temp8.SetStatus(tranTypes.TCPConnError)
						}
					} else {
						nums[0]++
						if nums[0] == 2 {
							temp3.SetStatus(tranTypes.TCPConnClosed)
						}
					}
					fmt.Println(fmt.Sprintf("dongcf---%3d---%s---%v---", v, addr, sub))
				}
				wg.Done()
			}(j)
		}
		time.Sleep(time.Second * 3)
		wg.Wait()
		fmt.Println("dongcf-------spend time distribution:", nums)
		time.Sleep(6 * time.Second)
		b.Dump()
	})
})
