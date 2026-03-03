"""Epoch and window computation utilities for the Seocheon SDK."""


def compute_epoch(block_height: int, epoch_length: int) -> int:
    """Calculate the epoch number for a given block height.

    epoch_number = (block_height - 1) / epoch_length
    """
    if block_height <= 0 or epoch_length <= 0:
        return 0
    return (block_height - 1) // epoch_length


def compute_window(block_height: int, epoch_length: int, windows_per_epoch: int) -> int:
    """Calculate the window index within an epoch for a given block height.

    window_index = ((block_height - 1) % epoch_length) / window_length
    """
    if epoch_length <= 0 or windows_per_epoch <= 0:
        return 0
    window_length = epoch_length // windows_per_epoch
    return ((block_height - 1) % epoch_length) // window_length


def epoch_start_block(epoch_number: int, epoch_length: int) -> int:
    """Return the starting block height for the given epoch number."""
    return epoch_number * epoch_length + 1


def epoch_end_block(epoch_number: int, epoch_length: int) -> int:
    """Return the ending block height for the given epoch number."""
    return (epoch_number + 1) * epoch_length


def window_start_block(epoch_start: int, window_index: int, window_length: int) -> int:
    """Return the starting block height for a given window within an epoch."""
    return epoch_start + window_index * window_length


def window_end_block(window_start: int, window_length: int) -> int:
    """Return the ending block height for a given window."""
    return window_start + window_length - 1
