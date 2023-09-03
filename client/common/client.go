package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os"

	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) communicateWithServer(msgID int) (string, error) {
	// Create the connection the server in every loop iteration. Send an
	c.createClientSocket()
	defer c.conn.Close()

	// TODO: Modify the send to avoid short-write
	fmt.Fprintf(
		c.conn,
		"[CLIENT %v] Message N°%v\n",
		c.config.ID,
		msgID,
	)

	return bufio.NewReader(c.conn).ReadString('\n')
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(sig chan os.Signal) {
loop:
	// autoincremental msgID to identify every message sent
	// Send messages if the loopLapse threshold has not been surpassed
	for msgID, timeout := 1, time.After(c.config.LoopLapse); ; msgID++ {
		msg, err := c.communicateWithServer(msgID)

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		select {
		case <-timeout:
			log.Infof("action: timeout_detected | result: success | client_id: %v",
				c.config.ID,
			)
			break loop
		case <-sig:
			log.Infof("action: signal_received | result: exiting | client_id: %v",
				c.config.ID,
			)
			return
		case <-time.After(c.config.LoopPeriod):
		}
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
