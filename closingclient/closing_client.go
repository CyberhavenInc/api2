package closingclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
	CloseIdleConnections()
}

type ClosingClient struct {
	impl HttpClient

	mu            sync.Mutex
	closing       bool
	cancels       map[uint64]func()
	lastCancelKey uint64

	wg sync.WaitGroup
}

func New(impl HttpClient) (*ClosingClient, error) {
	return &ClosingClient{
		impl:    impl,
		cancels: make(map[uint64]func()),
	}, nil
}

func (c *ClosingClient) Do(req *http.Request) (*http.Response, error) {
	c.wg.Add(1)
	defer c.wg.Done()

	ctx, cancel := context.WithCancel(req.Context())

	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return nil, fmt.Errorf("api2 client is closing")
	}
	key := c.lastCancelKey
	c.lastCancelKey++
	c.cancels[key] = cancel
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.cancels, key)
	}()

	return c.impl.Do(req.Clone(ctx))
}

func (c *ClosingClient) CloseIdleConnections() {
	c.impl.CloseIdleConnections()
}

func (c *ClosingClient) Close() error {
	c.mu.Lock()
	if !c.closing {
		c.closing = true
		// Close active connections.
		for _, cancel := range c.cancels {
			cancel()
		}
		c.cancels = nil
	}
	c.mu.Unlock()

	c.impl.CloseIdleConnections()

	c.wg.Wait()

	if closer, ok := c.impl.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	return nil
}
