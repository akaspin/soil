package bus

import (
	"context"
	"reflect"
	"fmt"
)

type TestingConsumer struct {
	ctx         context.Context
	data        []Message
	messageChan chan Message
	assertChan  chan struct {
		expect  []Message
		resChan chan error
	}
	assertLastChan chan struct {
		expect  Message
		resChan chan error
	}
}

func NewTestingConsumer(ctx context.Context) (c *TestingConsumer) {
	c = &TestingConsumer{
		ctx: ctx,
		messageChan: make(chan Message),
		assertChan: make(chan struct {
			expect  []Message
			resChan chan error
		}),
		assertLastChan: make(chan struct {
			expect  Message
			resChan chan error
		}),
	}
	go c.loop()
	return
}

func (c *TestingConsumer) ConsumeMessage(message Message) {
	select {
	case <-c.ctx.Done():
	case c.messageChan <- message:
	}
}

func (c *TestingConsumer) ExpectMessagesFn(expect ...Message) (fn func() error) {
	fn = func() (err error) {
		resChan := make(chan error)
		select {
		case <-c.ctx.Done():
			err = c.ctx.Err()
			return
		case c.assertChan <- struct {
			expect  []Message
			resChan chan error
		}{
			expect:  expect,
			resChan: resChan,
		}:
		}

		select {
		case <-c.ctx.Done():
			err = c.ctx.Err()
			return
		case err = <-resChan:
		}

		return
	}
	return
}

func (c *TestingConsumer) ExpectLastMessageFn(message Message) (fn func() error) {
	fn = func() (err error) {
		resChan := make(chan error)
		select {
		case <-c.ctx.Done():
			err = c.ctx.Err()
			return
		case c.assertLastChan <- struct {
			expect  Message
			resChan chan error
		}{
			expect:  message,
			resChan: resChan,
		}:
		}

		select {
		case <-c.ctx.Done():
			err = c.ctx.Err()
			return
		case err = <-resChan:
		}
		return
	}
	return
}

func (c *TestingConsumer) loop() {
LOOP:
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.messageChan:
			c.data = append(c.data, msg)
		case assertReq := <-c.assertChan:
			var err error
			if !reflect.DeepEqual(assertReq.expect, c.data) {
				err = fmt.Errorf("not equal (expected)%s != (actual)%s", assertReq.expect, c.data)
			}
			assertReq.resChan <- err
		case assertReq := <-c.assertLastChan:
			var err error
			if len(c.data) == 0 {
				assertReq.resChan <- fmt.Errorf(`no messages found`)
				continue LOOP
			}
			if !reflect.DeepEqual(assertReq.expect, c.data[len(c.data)-1]) {
				err = fmt.Errorf("not equal (expected)%s != (actual)%s", assertReq.expect, c.data[len(c.data)-1])
			}
			assertReq.resChan <- err
		}
	}
}
