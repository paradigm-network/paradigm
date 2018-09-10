package timer

import (
	"math/rand"
	"time"
)

type TimerFactory func() <-chan time.Time

type ControlTimer struct {
	TimerFactory TimerFactory
	TickCh       chan struct{} //sends a signal to listening process
	ResetCh      chan struct{} //receives instruction to reset the heartbeatTimer
	StopCh       chan struct{} //receives instruction to stop the heartbeatTimer
	ShutdownCh   chan struct{} //receives instruction to exit Run loop
	Set          bool
}

func NewControlTimer(timerFactory TimerFactory) *ControlTimer {
	return &ControlTimer{
		TimerFactory: timerFactory,
		TickCh:       make(chan struct{}),
		ResetCh:      make(chan struct{}),
		StopCh:       make(chan struct{}),
		ShutdownCh:   make(chan struct{}),
	}
}

func NewRandomControlTimer(base time.Duration) *ControlTimer {

	randomTimeout := func() <-chan time.Time {
		minVal := base
		if minVal == 0 {
			return nil
		}
		extra := (time.Duration(rand.Int63()) % minVal)
		return time.After(minVal + extra)
	}
	return NewControlTimer(randomTimeout)
}

func (c *ControlTimer) Run() {

	setTimer := func() <-chan time.Time {
		c.Set = true
		return c.TimerFactory()
	}

	timer := setTimer()
	for {
		select {
		case <-timer:
			c.TickCh <- struct{}{}
			c.Set = false
		case <-c.ResetCh:
			timer = setTimer()
		case <-c.StopCh:
			timer = nil
			c.Set = false
		case <-c.ShutdownCh:
			c.Set = false
			return
		}
	}
}

func (c *ControlTimer) Shutdown() {
	close(c.ShutdownCh)
}
