import XCTest
@testable import SeocheonSDK

final class MessagesTests: XCTestCase {

    func testMsgSubmitActivityTypeURL() {
        let msg = MsgSubmitActivity(submitter: "seocheon1abc", activityHash: String(repeating: "a", count: 64), contentURI: "ipfs://test")
        XCTAssertEqual(msg.typeURL, "/seocheon.activity.v1.MsgSubmitActivity")
    }

    func testMsgSubmitActivityEncode() {
        let msg = MsgSubmitActivity(submitter: "seocheon1abc", activityHash: String(repeating: "a", count: 64), contentURI: "ipfs://test")
        let data = msg.encode()
        XCTAssertFalse(data.isEmpty)
    }

    func testMsgSendTypeURL() {
        let msg = MsgSend(fromAddress: "seocheon1from", toAddress: "seocheon1to", amount: [Coin(denom: "uppyeo", amount: "1000")])
        XCTAssertEqual(msg.typeURL, "/cosmos.bank.v1beta1.MsgSend")
    }

    func testMsgSendEncode() {
        let msg = MsgSend(fromAddress: "seocheon1from", toAddress: "seocheon1to", amount: [Coin(denom: "uppyeo", amount: "1000")])
        let data = msg.encode()
        XCTAssertFalse(data.isEmpty)
    }

    func testMsgWithdrawNodeCommissionTypeURL() {
        let msg = MsgWithdrawNodeCommission(operator: "seocheon1op")
        XCTAssertEqual(msg.typeURL, "/seocheon.node.v1.MsgWithdrawNodeCommission")
    }

    func testMsgWithdrawNodeCommissionEncode() {
        let msg = MsgWithdrawNodeCommission(operator: "seocheon1op")
        let data = msg.encode()
        XCTAssertFalse(data.isEmpty)
    }
}
