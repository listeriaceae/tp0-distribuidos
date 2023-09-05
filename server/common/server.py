import socket
import logging

import common.utils as utils


AGENCY_COUNT = 5


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.agencies = []

    def __enter__(self):
        return self

    def __exit__(self, exc_type, value, traceback):
        self.close()

        return exc_type is utils.SignalException

    def close(self):
        self._server_socket.close()
        for client_sock, _ in self.agencies:
            client_sock.close()

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client stops sending data,
        server starts to accept new connections again. When all clients
        are done, server finds winners.
        """
        send = lambda writer, block: writer.send(block)

        while len(self.agencies) < AGENCY_COUNT:
            client_sock = self.__accept_new_connection()
            try:
                self.agencies.append((client_sock, self.__handle_client_connection(client_sock)))
            except (OSError, PartialDataError):
                logging.info(f"action: connection | result: fail | error: {e}")
                return

        logging.info("action: sorteo | result: success")
        winners = [(bet.agency, bet.document) for bet in utils.load_bets() if utils.has_won(bet)]
        for client_sock, agency in self.agencies:
            msg = ''.join(bet[1]+'\n' for bet in winners if bet[0] == agency)
            msg = msg.encode('utf-8')
            if utils.try_write(send, client_sock, msg) != len(msg):
                logging.error(f"action: send_response | result: fail | agency: {agency} | error: short write")
            else:
                logging.info(f"action: send_response | result: success | agency: {agency}")


    def __handle_client_connection(self, client_sock):
        """
        Receive batches of Bets from an agency and store them accordingly
        """
        send = lambda writer, block: writer.send(block)
        recv = lambda reader, n: reader.recv(n)
        try:
            while (buffer := client_sock.recv(4096)) != b"":
                if len(buffer) < 2:
                    raise utils.PartialDataError("short read")
                batch_size = int.from_bytes(buffer[:2], "big")
                if batch_size == 0:
                    break
                consumed = 2
                buffer += utils.read_batch(recv, client_sock, batch_size - (len(buffer) - consumed))

                if len(buffer) != batch_size + consumed:
                    raise utils.PartialDataError(f"Expected {batch_size} bytes, got {len(buffer)-2}")

                bets = []
                while consumed < len(buffer):
                    bet, n = utils.Bet.from_bytes(buffer[consumed:])
                    bets.append(bet)
                    consumed += n

                utils.store_bets(bets)

                msg = b"success"
                if utils.try_write(send, client_sock, msg) != len(msg):
                    raise utils.PartialDataError("short write")

            if len(buffer) != 4:
                raise utils.PartialDataError("Expected agency encoded as 2 bytes")
        except (OSError, PartialDataError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            raise

        logging.info(f"action: apuesta_almacenada | result: success")
        return int.from_bytes(buffer[2:], "big")


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
