package common

import (
	"bytes"
	"io"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	Agency        int
	ServerAddress string
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
			c.config.Agency,
			err,
		)
	}
	c.conn = conn
	return nil
}

func (c *Client) SendBet(bet Bet) {
	var buf bytes.Buffer

	c.createClientSocket()
	defer c.conn.Close()

	// Ignore errors as writing to buf can't fail
	bet.Marshal(&buf, c.config.Agency)

	_, err := io.Copy(c.conn, &buf)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | agency: %v | error: %v",
			c.config.Agency,
			err,
		)
		return
	}

	result, err := io.ReadAll(c.conn)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | agency: %v | error: %v",
			c.config.Agency,
			err,
		)
		return
	}
	log.Infof(
		"action: apuesta_enviada | result: %s | dni: %v | numero: %v",
		result,
		bet.Document,
		bet.Number,
	)
}

func (c *Client) Start(sig chan os.Signal, bet Bet) {
	done := make(chan bool)

	log.Infof(
		"action: enviar_apuesta | result: in_progress | dni: %v | numero: %v",
		bet.Document,
		bet.Number,
	)

	go func() {
		c.SendBet(bet)
		done <- true
	}()

	select {
	case <-sig:
		log.Infof("action: signal_received | result: exit")
		c.conn.Close()
		<-done
	case <-done:
	}
}
