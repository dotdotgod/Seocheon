"""Tests for epoch utilities."""

from seocheon.utils.epoch import (
    compute_epoch,
    compute_window,
    epoch_end_block,
    epoch_start_block,
    window_end_block,
    window_start_block,
)


def test_compute_epoch_block_1():
    assert compute_epoch(1, 17280) == 0


def test_compute_epoch_block_17280():
    assert compute_epoch(17280, 17280) == 0


def test_compute_epoch_block_17281():
    assert compute_epoch(17281, 17280) == 1


def test_compute_epoch_edge_zero():
    assert compute_epoch(0, 17280) == 0
    assert compute_epoch(100, 0) == 0
    assert compute_window(100, 0, 12) == 0
    assert compute_window(100, 17280, 0) == 0


def test_compute_window_first_window():
    assert compute_window(1, 17280, 12) == 0


def test_compute_window_second_window():
    # Window length = 1440, first window: blocks 1-1440, second: 1441-2880
    assert compute_window(1441, 17280, 12) == 1


def test_compute_window_last_window():
    # Last window of epoch 0: blocks 15841-17280
    assert compute_window(17280, 17280, 12) == 11


def test_epoch_start_block():
    assert epoch_start_block(0, 17280) == 1
    assert epoch_start_block(1, 17280) == 17281


def test_epoch_end_block():
    assert epoch_end_block(0, 17280) == 17280
    assert epoch_end_block(1, 17280) == 34560


def test_window_start_end_block():
    assert window_start_block(1, 0, 1440) == 1
    assert window_start_block(1, 1, 1440) == 1441
    assert window_end_block(1, 1440) == 1440
    assert window_end_block(1441, 1440) == 2880


def test_compute_window_resets_new_epoch():
    # First block of epoch 1 → window 0
    assert compute_window(17281, 17280, 12) == 0
