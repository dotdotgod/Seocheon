import XCTest
@testable import SeocheonSDK

final class EnvelopeTests: XCTestCase {

    func testEncodeTxBody() {
        let msg = MsgSubmitActivity(submitter: "seocheon1abc", activityHash: String(repeating: "a", count: 64), contentURI: "ipfs://test")
        let body = Envelope.encodeTxBody(messages: [msg], memo: "", timeoutHeight: 0)
        XCTAssertFalse(body.isEmpty)
    }

    func testEncodeAuthInfo() {
        let pubKey = Data(repeating: 0x02, count: 33)
        let coins = [Coin(denom: "uppyeo", amount: "50000")]
        let authInfo = Envelope.encodeAuthInfo(pubKey: pubKey, sequence: 0, feeCoins: coins, gasLimit: 200000)
        XCTAssertFalse(authInfo.isEmpty)
    }

    func testEncodeSignDoc() {
        let bodyBytes = Data([0x01, 0x02])
        let authInfoBytes = Data([0x03, 0x04])
        let signDoc = Envelope.encodeSignDoc(bodyBytes: bodyBytes, authInfoBytes: authInfoBytes, chainID: "seocheon-1", accountNumber: 1)
        XCTAssertFalse(signDoc.isEmpty)
    }

    func testEncodeTxRaw() {
        let bodyBytes = Data([0x01, 0x02])
        let authInfoBytes = Data([0x03, 0x04])
        let signature = Data(repeating: 0xAA, count: 64)
        let txRaw = Envelope.encodeTxRaw(bodyBytes: bodyBytes, authInfoBytes: authInfoBytes, signatures: [signature])
        XCTAssertFalse(txRaw.isEmpty)
    }

    func testEncodeAny() {
        let data = Data([0x01])
        let anyData = Envelope.encodeAny(typeURL: "/test.Type", value: data)
        XCTAssertFalse(anyData.isEmpty)
    }

    func testEncodePubKeyAny() {
        let pubKey = Data(repeating: 0x02, count: 33)
        let pubKeyAny = Envelope.encodePubKeyAny(pubKey)
        XCTAssertFalse(pubKeyAny.isEmpty)
    }
}
