import socket
import logging

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
        Receive batches of Bets from an agency and store them accordingly
        """
        send = lambda writer, block: writer.send(block)
        recv = lambda reader, n: reader.recv(n)
        try:
            while (buffer := client_sock.recv(4096)) != b"":
                if len(buffer) < 2:
                    raise utils.PartialDataError
                batch_size = int.from_bytes(buffer[:2], "big")
                consumed = 2
                buffer += utils.read_batch(recv, client_sock, batch_size - (len(buffer) - consumed))

                if len(buffer) != batch_size + consumed:
                    raise utils.PartialDataError

                bets = []
                while consumed < len(buffer):
                    bet, n = utils.Bet.from_bytes(buffer[consumed:])
                    bets.append(bet)
                    consumed += n

                utils.store_bets(bets)

                msg = b"success"
                logging.info(f"action: apuesta_almacenada | result: success")
                if utils.try_write(send, client_sock, msg) != len(msg):
                    logging.error(f"action: send_response | result: fail | error: short_write")
        except (OSError, PartialDataError) as e:
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
