package com.seocheon.sdk.infrastructure.signing

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class VaultServiceTest {

    @Test
    fun `VaultService create rejects blank endpoint`() {
        assertThrows(IllegalArgumentException::class.java) {
            kotlinx.coroutines.test.runTest {
                VaultService.create("", "key1")
            }
        }
    }

    @Test
    fun `VaultService create rejects blank key name`() {
        assertThrows(IllegalArgumentException::class.java) {
            kotlinx.coroutines.test.runTest {
                VaultService.create("http://vault:8200", "")
            }
        }
    }

    @Test
    fun `KeystoreService create rejects blank path`() {
        assertThrows(IllegalArgumentException::class.java) {
            KeystoreService.create("", "passphrase")
        }
    }
}
