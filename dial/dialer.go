package dial

import (
	"net"
	"net/http"

	"github.com/ssimunic/gossm/logger"
)

// Dialer is used to test connections
type Dialer struct {
	semaphore chan struct{}
}

// Status saves information about connection
type Status struct {
	Ok  bool
	Err error
}

// NewDialer returns pointer to new Dialer
func NewDialer(concurrentConnections int) *Dialer {
	return &Dialer{
		semaphore: make(chan struct{}, concurrentConnections),
	}
}

// NewWorker is used to send address over NetAddressTimeout to make request and receive status over DialerStatus
// Blocks until slot in semaphore channel for concurrency is free
func (d *Dialer) NewWorker() (chan<- NetAddressTimeout, <-chan Status) {
	netAddressTimeoutCh := make(chan NetAddressTimeout)
	dialerStatusCh := make(chan Status)

	d.semaphore <- struct{}{}

	go func() {

		var err error
		var dialerStatus Status
		netAddressTimeout := <-netAddressTimeoutCh

		if netAddressTimeout.NetAddress.Network == "http" {
			err = func() error {
				client := http.Client{
					Timeout: netAddressTimeout.Timeout,
				}
				resp, err := client.Get(netAddressTimeout.Address)
				if err != nil {
					return err
				}
				logger.Logln(" -> ", resp.Status, resp.Request.URL, resp.Proto)
				return resp.Body.Close()
			}()
		} else {
			err = func() error {
				conn, err := net.DialTimeout(netAddressTimeout.Network, netAddressTimeout.Address, netAddressTimeout.Timeout)
				if err != nil {
					return err
				}
				conn.Close()
				return nil
			}()
		}

		if err != nil {
			dialerStatus.Ok = false
			dialerStatus.Err = err
		} else {
			dialerStatus.Ok = true
		}
		dialerStatusCh <- dialerStatus
		<-d.semaphore
	}()

	return netAddressTimeoutCh, dialerStatusCh
}
