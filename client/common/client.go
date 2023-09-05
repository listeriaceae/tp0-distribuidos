package common

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	Agency        int
	BatchSize     int
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

func (c *Client) sendBatch(r io.Reader, length int) error {
	result, err := sendBatch(c.conn, r, length)
	if err != nil {
		log.Errorf(
			"action: apuestas_enviadas | result: fail | agency: %v | error: %v",
			c.config.Agency,
			err,
		)
	} else {
		log.Infof(
			"action: apuestas_enviadas | result: %s | agency: %v",
			result,
			c.config.Agency,
		)
	}
	return err
}

func (c *Client) SendBets(r *csv.Reader) error {
	var buf bytes.Buffer

	for i := 0; ; i++ {
		if i == c.config.BatchSize {
			if err := c.sendBatch(&buf, buf.Len()); err != nil {
				return err
			}
			i = 0
			buf.Reset()
		}

		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf(
				"action: apuestas_enviadas | result: fail | agency: %v | error: %v",
				c.config.Agency,
				err,
			)
			return err
		}

		// Ignore errors as writing to buf can't fail
		BetMarshal(&buf, c.config.Agency, record)
	}

	if buf.Len() > 0 {
		if err := c.sendBatch(&buf, buf.Len()); err != nil {
			return err
		}
	}
	log.Infof("action: apuestas_enviadas | result: complete | agency: %v",
		c.config.Agency,
	)
	return nil
}

func (c *Client) RequestWinners() (int, error) {
	err := RequestWinners(c.conn, c.config.Agency)
	if err != nil {
		return 0, err
	}

	cant_ganadores := 0
	scanner := bufio.NewScanner(c.conn)
	for ; scanner.Scan(); cant_ganadores++ {
	}

	return cant_ganadores, scanner.Err()
}

func (c *Client) Start(sig chan os.Signal, r io.Reader) {
	done := make(chan error)

	c.createClientSocket()

	log.Infof(
		"action: enviar_apuestas | result: in_progress | agency: %v",
		c.config.Agency,
	)

	go func() {
		done <- c.SendBets(csv.NewReader(r))
	}()

	select {
	case <-sig:
		log.Infof(
			"action: signal_received | result: exit | agency: %v",
			c.config.Agency,
		)
		c.conn.Close()
		<-done
		return
	case err := <-done:
		if err != nil {
			c.conn.Close()
			return
		}
	}

	go func() {
		cant, err := c.RequestWinners()
		if err != nil {
			log.Errorf(
				"action: consulta_ganadores | result: fail | error: %v",
				err,
			)
		} else {
			log.Infof(
				"action: consulta_ganadores | result: success | cant_ganadores: %s",
				cant,
			)
		}
		done <- err
	}()

	select {
	case <-sig:
		log.Infof(
			"action: signal_received | result: exit | agency: %v",
			c.config.Agency,
		)
		c.conn.Close()
		<-done
	case <-done:
		c.conn.Close()
	}
}
