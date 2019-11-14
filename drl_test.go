package drl

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TykTechnologies/leakybucket"
	"github.com/TykTechnologies/leakybucket/memorycache"
)

func TestDRL(t *testing.T) {
	// defer goleak.VerifyNoLeaks(t)
	t.Run("3 servers", func(ts *testing.T) {
		// simulating 20rps in 2s which totals to 40 requests.
		testDRL(ts, 3, 20, 1, 20, 20)
	})
}

func testDRL(t *testing.T, numServers, rate, per, success, fail int) {
	var servers []Server
	for i := 0; i < numServers; i++ {
		istr := strconv.FormatInt(int64(i), 10)
		servers = append(servers, Server{
			HostName:   "host-" + istr,
			ID:         "id-" + istr,
			LoadPerSec: 1,
		})
	}
	collect := Collect{}
	var hosts []*Host
	for i := 0; i < numServers; i++ {
		hosts = append(hosts, &Host{
			id:      servers[i].ID + "|" + servers[i].HostName,
			server:  servers[i],
			send:    make(chan struct{}),
			stop:    make(chan struct{}),
			servers: servers,
			Collect: &collect,
			limit:   Limit{Rate: float64(rate), Per: per},
		})
	}
	for _, h := range hosts {
		go runDRL(t, h)
	}
	robin := &RoundRobin{hosts: hosts}
	generateLoad(t, robin, rate*2, time.Second)
	for _, h := range hosts {
		h.Close()
	}
	if collect.pass != int64(success) {
		t.Errorf("SUCCESS expected %d got %d", success, collect.pass)
	}
	if collect.fail != int64(fail) {
		t.Errorf("FAIL expected %d got %d", fail, collect.fail)
	}
	t.Logf("TOTA REQUESTS :%d", collect.pass+collect.fail)
}

func runDRL(t *testing.T, h *Host) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dd := &DRL{
		ThisServerID: h.id,
	}
	dd.Init(ctx)
	dd.AddOrUpdateServer(h.server)
	for _, sv := range h.servers {
		err := dd.AddOrUpdateServer(sv)
		if err != nil {
			t.Logf("error trying to add server %v", err)
		}
	}
	store := memorycache.New()
	key := "bucket-key"
	for {
		select {
		case <-h.stop:
			return
		case <-h.send:
			err := checkLimit(key, store, h.limit, dd)
			if err != nil {
				h.Collect.Fail()
				// t.Logf("failed check [%s] %v", h.id, err)
			} else {
				h.Collect.Pass()
			}
		}
	}
}

type Limit struct {
	Rate float64
	Per  int
}

type Collect struct {
	pass int64
	fail int64
}

func (c *Collect) Pass() {
	atomic.AddInt64(&c.pass, 1)
}

func (c *Collect) Fail() {
	atomic.AddInt64(&c.fail, 1)
}

func checkLimit(
	bucketKey string,
	store leakybucket.Storage,
	lm Limit,
	dd *DRL,
) error {
	rate := uint(lm.Rate * float64(dd.RequestTokenValue))
	if rate < uint(dd.CurrentTokenValue()) {
		rate = uint(dd.CurrentTokenValue())
	}
	userBucket, err := store.Create(bucketKey, rate, time.Duration(lm.Per)*time.Second)
	if err != nil {
		return err
	}
	_, err = userBucket.Add(uint(dd.CurrentTokenValue()))
	if err != nil {
		return err
	}
	return nil
}

type Host struct {
	id      string
	server  Server
	send    chan struct{}
	stop    chan struct{}
	limit   Limit
	servers []Server
	Collect *Collect
}

func (h *Host) Ping() {
	h.send <- struct{}{}
}

func (h *Host) Close() {
	h.stop <- struct{}{}
}

type RoundRobin struct {
	hosts []*Host
	mu    sync.Mutex
	next  int
}

func (r *RoundRobin) Next() *Host {
	r.mu.Lock()
	s := r.hosts[r.next]
	r.next = (r.next + 1) % len(r.hosts)
	r.mu.Unlock()
	return s
}

func generateLoad(t *testing.T, r *RoundRobin, rate int, duration time.Duration) {
	tick := time.NewTicker(time.Second)
	end := time.Now().Add(duration)
	defer tick.Stop()
	n := 0
	for {
		select {
		case ts := <-tick.C:
			if ts.After(end) {
				return
			}
			n = 0
		default:
			if n > rate-1 {
				continue
			}
			r.Next().Ping()
			n++
		}
	}
}
