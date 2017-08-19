package gossm

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ssimunic/gossm/logger"
	"github.com/ssimunic/gossm/tracker"
)

type Monitor struct {
	// Holds settings and servers
	config *Config

	// Channel used to schedule checks for servers
	checkerCh chan *Server

	// Notification methods used to send messages when server can't be reached
	notifiers Notifiers

	// Channel used for receive servers that couldn't be reached
	notifierCh chan *Server

	// To reduce notification spam, tracker is used to delay notifications
	notificationTracker map[*Server]*tracker.TimeTracker

	// Used to regulate number of concurrent connections
	semaphore chan struct{}

	// Sending to stop channel makes program exit
	stop chan struct{}
}

func NewMonitor(c *Config) *Monitor {
	m := &Monitor{
		config:              c,
		checkerCh:           make(chan *Server),
		notifiers:           c.Settings.Notifications.GetNotifiers(),
		notifierCh:          make(chan *Server),
		notificationTracker: make(map[*Server]*tracker.TimeTracker),
		semaphore:           make(chan struct{}, c.Settings.Monitor.MaxConnections),
		stop:                make(chan struct{}),
	}

	m.initialize()

	return m
}

func (m *Monitor) initialize() {
	// Initialize notification methods to reduce overhead
	for _, notifier := range m.notifiers {
		if initializer, ok := notifier.(Initializer); ok {
			initializer.Initialize()
		}
	}

	// Initialize notificationTracker
	for _, server := range m.config.Servers {
		m.notificationTracker[server] =
			tracker.NewTimeTracker(tracker.NewExpBackoff(m.config.Settings.Monitor.ExponentialBackoffSeconds))
	}

	// Set default CheckInterval and Timeout for servers who miss them
	for _, server := range m.config.Servers {
		switch {
		case server.CheckInterval <= 0:
			server.CheckInterval = m.config.Settings.Monitor.CheckInterval
		case server.Timeout <= 0:
			server.Timeout = m.config.Settings.Monitor.Timeout
		}
	}
}

// Run runs monitor infinitely
func (m *Monitor) Run() {
	m.RunForSeconds(0)
}

// RunForSeconds runs monitor for runningSeconds seconds or infinitely if 0 is passed as an argument
func (m *Monitor) RunForSeconds(runningSeconds int) {
	if runningSeconds != 0 {
		go func() {
			runningSecondsTime := time.Duration(runningSeconds) * time.Second
			<-time.After(runningSecondsTime)
			m.stop <- struct{}{}
		}()
	}

	for _, server := range m.config.Servers {
		go m.scheduleServer(server)
	}

	logger.Logln("Starting monitor.")
	m.monitor()
}

func (m *Monitor) scheduleServer(s *Server) {
	tickerSeconds := time.NewTicker(time.Duration(s.CheckInterval) * time.Second)

	for range tickerSeconds.C {
		m.checkerCh <- s
	}
}

func (m *Monitor) monitor() {
	go m.listenForChecks()
	go m.listenForNotifications()

	// Wait for termination signal then exit monitor
	<-m.stop
	logger.Logln("Terminating.")
	os.Exit(0)
}

func (m *Monitor) listenForChecks() {
	for server := range m.checkerCh {
		go func(server *Server) {
			m.semaphore <- struct{}{}
			m.checkServerStatus(server)
			<-m.semaphore
		}(server)
	}
}

func (m *Monitor) listenForNotifications() {
	for server := range m.notifierCh {
		timeTracker := m.notificationTracker[server]
		if timeTracker.IsReady() {
			nextDelay, nextTime := timeTracker.SetNext()
			go m.notifiers.NotifyAll(server.String())
			logger.Logln("Next available notification for", server.String(), "in", nextDelay, "at", nextTime)
		}
	}
}

func (m *Monitor) checkServerStatus(server *Server) {
	logger.Logln("Checking", server)
	formattedAddress := fmt.Sprintf("%s:%d", server.IPAddress, server.Port)
	timeoutSeconds := time.Duration(server.Timeout) * time.Second
	conn, err := net.DialTimeout(server.Protocol, formattedAddress, timeoutSeconds)
	if err != nil {
		logger.Logln(err)
		logger.Logln("ERROR", server)
		go func() {
			m.notifierCh <- server
		}()
		return
	}
	defer conn.Close()
	logger.Logln("OK", server)
}
