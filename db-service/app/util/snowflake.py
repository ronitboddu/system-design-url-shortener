import threading
import time

class SnowflakeGenerator:
    def __init__(self, node_id: int, epoch: int = 1704067200000):
        # epoch is in milliseconds; default here is 2024-01-01 UTC
        self.node_id = node_id
        self.epoch = epoch

        self.node_id_bits = 10
        self.sequence_bits = 12

        self.max_node_id = (1 << self.node_id_bits) - 1
        self.max_sequence = (1 << self.sequence_bits) - 1

        if node_id < 0 or node_id > self.max_node_id:
            raise ValueError(f"node_id must be between 0 and {self.max_node_id}")

        self.sequence = 0
        self.last_timestamp = -1
        self.lock = threading.Lock()

        self.node_id_shift = self.sequence_bits
        self.timestamp_shift = self.sequence_bits + self.node_id_bits

    def _current_millis(self) -> int:
        return int(time.time() * 1000)

    def _wait_next_millis(self, last_timestamp: int) -> int:
        timestamp = self._current_millis()
        while timestamp <= last_timestamp:
            timestamp = self._current_millis()
        return timestamp

    def next_id(self) -> int:
        with self.lock:
            timestamp = self._current_millis()

            if timestamp < self.last_timestamp:
                raise RuntimeError("Clock moved backwards. Refusing to generate id.")

            if timestamp == self.last_timestamp:
                self.sequence = (self.sequence + 1) & self.max_sequence
                if self.sequence == 0:
                    timestamp = self._wait_next_millis(self.last_timestamp)
            else:
                self.sequence = 0

            self.last_timestamp = timestamp

            return (
                ((timestamp - self.epoch) << self.timestamp_shift)
                | (self.node_id << self.node_id_shift)
                | self.sequence
            )
