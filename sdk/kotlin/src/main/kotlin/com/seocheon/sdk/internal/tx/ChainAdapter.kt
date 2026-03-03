package com.seocheon.sdk.internal.tx

import com.seocheon.sdk.infrastructure.ChainClient

/**
 * Adapter that bridges ChainClient to ChainQuerier interface.
 * ChainClient already implements ChainQuerier, so this provides
 * a convenience factory for explicit typing.
 */
object ChainAdapter {

    /**
     * Returns the ChainClient cast as a ChainQuerier.
     */
    fun asQuerier(client: ChainClient): ChainQuerier = client
}
