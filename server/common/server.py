import socket
import logging
from functools import partial

import common.utils as utils


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, value, traceback):
        logging.info(f'action: exit_requested | result: success | closing listener')
        self.close()

        return exc_type is utils.SignalException

    def close(self):
        self._server_socket.close()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while True:
            with self.__accept_new_connection() as client_sock:
                self.__handle_client_connection(client_sock)

    def __handle_client_connection(self, client_sock):
        """
        Receive a Bet from an agency and store it
        """
        send = lambda writer, block: writer.send(block)
        try:
            buffer = b""
            for block in iter(partial(client_sock.recv, 1024), b""):
                buffer += block
                if len(buffer) < utils.HEADER_SIZE:
                    continue
                try:
                    bet, _ = utils.Bet.from_bytes(buffer)
                except PartialDataError:
                    pass
                else:
                    utils.store_bets((bet,))
                    break

            msg = b"success"
            logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}")
            if utils.try_write(send, client_sock, msg) != len(msg):
                logging.error(f"action: send_response | result: fail | error: short_write")
        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
