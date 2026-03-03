package com.seocheon.sdk.utils

import com.seocheon.sdk.constants.ChainConstants
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class EpochUtilsTest {

    @Test
    fun `computeEpoch returns 0 for block 1`() {
        assertEquals(0, EpochUtils.computeEpoch(1))
    }

    @Test
    fun `computeEpoch returns 0 for last block of epoch 0`() {
        assertEquals(0, EpochUtils.computeEpoch(ChainConstants.EPOCH_LENGTH))
    }

    @Test
    fun `computeEpoch returns 1 for first block of epoch 1`() {
        assertEquals(1, EpochUtils.computeEpoch(ChainConstants.EPOCH_LENGTH + 1))
    }

    @Test
    fun `computeEpoch returns 0 for zero or negative height`() {
        assertEquals(0, EpochUtils.computeEpoch(0))
        assertEquals(0, EpochUtils.computeEpoch(-1))
    }

    @Test
    fun `computeWindow returns 0 for block 1`() {
        assertEquals(0, EpochUtils.computeWindow(1))
    }

    @Test
    fun `computeWindow returns correct window index`() {
        // Window 1 starts at block 1441
        assertEquals(1, EpochUtils.computeWindow(1441))
        // Last block of window 0 is 1440
        assertEquals(0, EpochUtils.computeWindow(1440))
    }

    @Test
    fun `computeWindow returns last window index`() {
        // Window 11 (0-indexed) is the last window in epoch 0
        val lastWindowStart = 11 * ChainConstants.WINDOW_LENGTH + 1
        assertEquals(11, EpochUtils.computeWindow(lastWindowStart))
    }

    @Test
    fun `epochStartBlock returns correct start`() {
        assertEquals(1, EpochUtils.epochStartBlock(0))
        assertEquals(ChainConstants.EPOCH_LENGTH + 1, EpochUtils.epochStartBlock(1))
    }

    @Test
    fun `epochEndBlock returns correct end`() {
        assertEquals(ChainConstants.EPOCH_LENGTH, EpochUtils.epochEndBlock(0))
        assertEquals(ChainConstants.EPOCH_LENGTH * 2, EpochUtils.epochEndBlock(1))
    }

    @Test
    fun `windowStartBlock returns correct start`() {
        val epochStart = EpochUtils.epochStartBlock(0)
        assertEquals(1, EpochUtils.windowStartBlock(epochStart, 0))
        assertEquals(1441, EpochUtils.windowStartBlock(epochStart, 1))
    }

    @Test
    fun `windowEndBlock returns correct end`() {
        assertEquals(1440, EpochUtils.windowEndBlock(1))
        assertEquals(2880, EpochUtils.windowEndBlock(1441))
    }
}
