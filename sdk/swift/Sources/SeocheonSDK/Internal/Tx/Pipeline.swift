import Foundation

/// Configuration for the TX pipeline.
internal struct PipelineConfig {
    let chainID: String
    var gasPrice: UInt64 = Gas.defaultGasPrice
    var confirmTimeoutSeconds: TimeInterval = 30
    var pollIntervalSeconds: TimeInterval = 1

    static func defaultConfig(chainID: String) -> PipelineConfig {
        return PipelineConfig(chainID: chainID)
    }
}

/// Executes the full 4-phase TX pipeline:
///   Phase 1 - Assembly: query account, build TxBody + AuthInfo + SignDoc
///   Phase 2 - Signing: sign the SignDoc
///   Phase 3 - Broadcast: encode TxRaw and broadcast
///   Phase 4 - Confirmation: poll for TX inclusion
internal enum TxPipeline {
    static func executeTx(
        querier: ChainQuerier,
        signer: TxSigner,
        config: PipelineConfig,
        request: TxRequest
    ) async throws -> TxResult {
        // Phase 1: Assembly
        let address = signer.getAddress()
        let pubKey = signer.getPubKey()

        let accountInfo = try await querier.getAccountInfo(address: address)

        // Determine gas limit
        var gasLimit = request.gasLimit
        if gasLimit == 0 {
            gasLimit = Gas.defaultGasForMessage(request.message.typeURL)
        }

        // Determine fee
        var feeAmount = request.feeAmount
        if feeAmount == 0 {
            let gasPrice = config.gasPrice > 0 ? config.gasPrice : Gas.defaultGasPrice
            feeAmount = Gas.calculateFee(gasLimit: gasLimit, gasPrice: gasPrice)
        }

        let feeDenom = request.feeDenom.isEmpty ? Gas.defaultFeeDenom : request.feeDenom

        // Encode TxBody
        let bodyBytes = Envelope.encodeTxBody(messages: [request.message], memo: request.memo, timeoutHeight: request.timeoutHeight)

        // Encode AuthInfo
        let feeCoins = [Coin(denom: feeDenom, amount: "\(feeAmount)")]
        let authInfoBytes = Envelope.encodeAuthInfo(pubKey: pubKey, sequence: accountInfo.sequence, feeCoins: feeCoins, gasLimit: gasLimit)

        // Encode SignDoc
        let signDocBytes = Envelope.encodeSignDoc(bodyBytes: bodyBytes, authInfoBytes: authInfoBytes, chainID: config.chainID, accountNumber: accountInfo.accountNumber)

        // Phase 2: Signing
        let signature = try await signer.sign(signDocBytes)

        // Phase 3: Broadcast
        let txRawBytes = Envelope.encodeTxRaw(bodyBytes: bodyBytes, authInfoBytes: authInfoBytes, signatures: [signature])

        let broadcastResult = try await querier.broadcastTxSync(txBytes: txRawBytes)

        // Check broadcast result code
        if broadcastResult.code != 0 {
            return TxResult(
                txHash: broadcastResult.txHash,
                code: broadcastResult.code,
                rawLog: broadcastResult.rawLog
            )
        }

        // Phase 4: Confirmation
        let result = try await TxConfirm.pollTxConfirmation(
            querier: querier,
            txHash: broadcastResult.txHash,
            timeout: config.confirmTimeoutSeconds,
            pollInterval: config.pollIntervalSeconds
        )

        return result
    }
}
