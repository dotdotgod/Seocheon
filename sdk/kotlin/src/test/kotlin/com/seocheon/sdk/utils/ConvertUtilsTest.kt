package com.seocheon.sdk.utils

import com.seocheon.sdk.constants.ChainConstants
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import java.math.BigInteger

class ConvertUtilsTest {

    @Test
    fun `convertDenom same denom returns same amount`() {
        val amount = BigInteger.valueOf(100)
        assertEquals(amount, ConvertUtils.convertDenom(amount, "uppyeo", "uppyeo"))
    }

    @Test
    fun `convertDenom kkot to uppyeo`() {
        val result = ConvertUtils.convertDenom(BigInteger.ONE, "kkot", "uppyeo")
        assertEquals(BigInteger.valueOf(ChainConstants.UPPYEO_PER_KKOT), result)
    }

    @Test
    fun `convertDenom uppyeo to kkot`() {
        val result = ConvertUtils.convertDenom(
            BigInteger.valueOf(ChainConstants.UPPYEO_PER_KKOT), "uppyeo", "kkot"
        )
        assertEquals(BigInteger.ONE, result)
    }

    @Test
    fun `convertDenom sal to pi`() {
        // 1 sal = 100 uppyeo, 1 pi = 10000 uppyeo
        // 100 sal = 10000 uppyeo = 1 pi
        val result = ConvertUtils.convertDenom(BigInteger.valueOf(100), "sal", "pi")
        assertEquals(BigInteger.ONE, result)
    }

    @Test
    fun `convertDenom hon to sum`() {
        // 1 hon = 100000000 uppyeo, 1 sum = 1000000 uppyeo
        // 1 hon = 100 sum
        val result = ConvertUtils.convertDenom(BigInteger.ONE, "hon", "sum")
        assertEquals(BigInteger.valueOf(100), result)
    }

    @Test
    fun `convertDenom throws for unknown denom`() {
        assertThrows(IllegalArgumentException::class.java) {
            ConvertUtils.convertDenom(BigInteger.ONE, "unknown", "uppyeo")
        }
    }

    @Test
    fun `formatKkot formats correctly`() {
        assertEquals("1.0000000000", ConvertUtils.formatKkot(10000000000L))
        assertEquals("0.0000000001", ConvertUtils.formatKkot(1L))
        assertEquals("0.0000000000", ConvertUtils.formatKkot(0L))
    }

    @Test
    fun `parseKkot parses correctly`() {
        assertEquals(10000000000L, ConvertUtils.parseKkot("1.0000000000"))
        assertEquals(10000000000L, ConvertUtils.parseKkot("1"))
        assertEquals(1L, ConvertUtils.parseKkot("0.0000000001"))
    }

    @Test
    fun `parseKkot and formatKkot roundtrip`() {
        val original = 12345678901L
        val formatted = ConvertUtils.formatKkot(original)
        val parsed = ConvertUtils.parseKkot(formatted)
        assertEquals(original, parsed)
    }

    @Test
    fun `denomFactor returns correct factors`() {
        assertEquals(BigInteger.ONE, ConvertUtils.denomFactor("uppyeo"))
        assertEquals(BigInteger.valueOf(100), ConvertUtils.denomFactor("sal"))
        assertEquals(BigInteger.valueOf(10000), ConvertUtils.denomFactor("pi"))
        assertEquals(BigInteger.valueOf(1000000), ConvertUtils.denomFactor("sum"))
        assertEquals(BigInteger.valueOf(100000000), ConvertUtils.denomFactor("hon"))
        assertEquals(BigInteger.valueOf(10000000000L), ConvertUtils.denomFactor("kkot"))
    }
}
