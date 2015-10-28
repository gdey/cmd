// Copyright 2015 Gautam Dey. All right reserved.
// Use of this source is governed by a BSD-style license that can be found in the
// LICENSE file.

// Package cmd contains the Context type that can be used to cleanly terminate an
// application upon receiving a Termination signal. This package wrapps the basic
// pattern I have observed to enable it to be simpler to use.
package cmd

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/net/context"
)

type Context struct {
	// The net/context context; can be useful to create sub contexts.
	// however, it would be better to have a seperate net/context tree.
	ctx    context.Context
	c      chan os.Signal
	cancel context.CancelFunc
	wg     sync.WaitGroup
	l      sync.RWMutex
	signal os.Signal
}

var ctx *Context

// Complete is a blocking call that should be the last call in your main function.
// The purpose of this function is to wait for the cancel go routine to cleanup
// corretly.
func (c *Context) Complete() {
	if c == nil {
		return
	}
	c.cancel()
	c.wg.Wait()
}

// Complete is a blocking call that should be the last call in your main function.
// The purpose of this function is to wait for the cancel go routine to cleanup
// corretly.
func Complete() {
	ctx.Complete()
}

// Cancelled is provided for use in select statments. It can be used to determine
// if a termination signal has been sent.
func (c *Context) Cancelled() <-chan struct{} {
	if c == nil {
		// If we are nil, we should ctr-c can not be trapped, so we should
		// always block.
		return nil
	}
	return c.ctx.Done()
}

// Cancelled is provided for use in select statments. It can be used to determine
// if a termination signal has been sent.
func Cancelled() <-chan struct{} {
	return ctx.Cancelled()
}

// IsCancelled is provided for use in if and for blocks, this can be used to check
// to see if a termination signal has been send, and to the excuate appropriate logic
// as needed.
func (c *Context) IsCancelled() bool {
	if c == nil {
		// If we are nil ctr-c can not be trapped, so we can not be in a
		// cancelled stated.
		return false
	}
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

// IsCancelled is provided for use in if and for blocks, this can be used to check
// to see if a termination signal has been send, and to the excuate appropriate logic
// as needed.
func IsCancelled() bool {
	return ctx.IsCancelled()
}

func (c *Context) signalHandler() {
	if c == nil {
		return
	}
	select {
	case s := <-c.c:
		c.l.Lock()
		c.signal = s
		c.l.Unlock()
		c.cancel()
	case <-c.ctx.Done():
	}
	c.wg.Done()
}

// Signal provides one the ability to introspect which signal was actually send.
func (c *Context) Signal() os.Signal {
	c.l.RLock()
	s := c.signal
	c.l.RUnlock()
	return s
}

// Signal provides one the ability to introspect which signal was actually send.
func Signal() os.Signal {
	return ctx.Signal()
}

// NewContext initilizes and setups up the context. An explicate list of signals
// can be passed in as well, if no list is passed os.Interrupt, and syscall.SIGTERM is
// assumed.
func NewContext(signals ...os.Signal) *Context {
	ch := make(chan os.Signal)
	if len(signals) == 0 {
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	} else {
		signal.Notify(ch, signals...)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := Context{
		ctx:    ctx,
		cancel: cancel,
		c:      ch,
		wg:     sync.WaitGroup{},
	}
	c.wg.Add(1)
	go c.signalHandler()
	return &c
}

// New initilizes and setups up the global context. An explicate list of signals
// can be passed in as well, if no list is passed os.Interrupt, and syscall.SIGTERM is
// assumed.
func New(signals ...os.Signal) *Context {
	if ctx == nil {
		ctx = NewContext(signals...)
	}
	return ctx
}

/*
func main(){
	cmd.New()
	defer cmd.Complete()
	for i := 0; i < 100; i++ {
		fmt.Println("Going to nap for a second.")
		select {
			case <-time.After(1 * time.Second):
				fmt.Println("Ahhh that was a good nap!")
			case <-cmd.Cancelled():
				fmt.Println("Ho! I got Ctr-C")
		}
		if cmd.IsCancelled() {
			fmt.Println("Ctr-C got called.")
			break
		}
		// do some chunk off work.
	}
}
func main(){
	ctx := cmd.NewContext()
	defer ctx.Complete()
	// Main code here
	for i := 0; i < 100; i++ {
		fmt.Println("Going to nap for a second.")
		select {
			case <-time.After(1 * time.Second):
				fmt.Println("Ahhh that was a good nap!")
			case <-ctx.Cancelled():
				fmt.Println("Ho! I got Ctr-C")
		}
		if ctx.IsCancelled() {
			fmt.Println("Ctr-C got called.")
			break
		}
		// do some chunk off work.
	}
}
*/
