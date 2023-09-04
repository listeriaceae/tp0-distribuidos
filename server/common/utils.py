import csv
import datetime
import time


""" Bets storage location. """
STORAGE_FILEPATH = "./bets.csv"
""" Simulated winner number in the lottery contest. """
LOTTERY_WINNER_NUMBER = 7574
""" Start of attributes. """
HEADER_SIZE = 12


""" A lottery bet registry. """
class Bet:
    def __init__(self, agency: str, first_name: str, last_name: str, document: str, birthdate: str, number: str):
        """
        agency must be passed with integer format.
        birthdate must be passed with format: 'YYYY-MM-DD'.
        number must be passed with integer format.
        """
        self.agency = int(agency)
        self.first_name = first_name
        self.last_name = last_name
        self.document = document
        self.birthdate = datetime.date.fromisoformat(birthdate)
        self.number = int(number)

    @classmethod
    def from_bytes(cls, stream: bytes):
        """
        stream must have at least HEADER_SIZE bytes.
        stream should have at least one Bet,
        otherwise PartialDataError is raised.
        returns the new Bet and any excess bytes.
        """
        agency_end     = int.from_bytes(stream[ 0: 2], "big")
        first_name_end = int.from_bytes(stream[ 2: 4], "big")
        last_name_end  = int.from_bytes(stream[ 4: 6], "big")
        document_end   = int.from_bytes(stream[ 6: 8], "big")
        birthdate_end  = int.from_bytes(stream[ 8:10], "big")
        number_end     = int.from_bytes(stream[10:12], "big")

        if len(stream) < number_end:
            raise PartialDataError(number_end - len(stream))

        agency     = stream[HEADER_SIZE    :     agency_end].decode('utf-8')
        first_name = stream[agency_end     : first_name_end].decode('utf-8')
        last_name  = stream[first_name_end :  last_name_end].decode('utf-8')
        document   = stream[last_name_end  :   document_end].decode('utf-8')
        birthdate  = stream[document_end   :  birthdate_end].decode('utf-8')
        number     = stream[birthdate_end  :     number_end].decode('utf-8')

        return cls(agency, first_name, last_name, document, birthdate, number), stream[number_end:]

""" Checks whether a bet won the prize or not. """
def has_won(bet: Bet) -> bool:
    return bet.number == LOTTERY_WINNER_NUMBER

"""
Persist the information of each bet in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def store_bets(bets: list[Bet]) -> None:
    with open(STORAGE_FILEPATH, 'a+') as file:
        writer = csv.writer(file, quoting=csv.QUOTE_MINIMAL)
        for bet in bets:
            writer.writerow([bet.agency, bet.first_name, bet.last_name,
                             bet.document, bet.birthdate, bet.number])

"""
Loads the information all the bets in the STORAGE_FILEPATH file.
Not thread-safe/process-safe.
"""
def load_bets() -> list[Bet]:
    with open(STORAGE_FILEPATH, 'r') as file:
        reader = csv.reader(file, quoting=csv.QUOTE_MINIMAL)
        for row in reader:
            yield Bet(row[0], row[1], row[2], row[3], row[4], row[5])


def try_write(fn, writer, block, max_remaining=0):
    """
    fn(writer, block) must return the number of bytes written or -1.
    call fn(writer, block) until max_remaining >= len(block) - bytes_written
    or an error occurs. return the number of bytes written.
    """
    bytes_written = 0
    while len(block) - bytes_written > max_remaining:
        ret = fn(writer, block[bytes_written:])
        if ret < 0:
            break
        bytes_written += ret

    return bytes_written


class PartialDataError(ValueError):
    pass


class SignalException(Exception):
    # do nothing
    pass


def handler(signum, frame):
    raise SignalException()
