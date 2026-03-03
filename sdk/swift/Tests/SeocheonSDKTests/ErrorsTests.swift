import XCTest
@testable import SeocheonSDK

final class ErrorsTests: XCTestCase {

    func testSDKErrorCodes() {
        XCTAssertEqual(SDKError.notConnected.code, 9000)
        XCTAssertEqual(SDKError.broadcastFailed("").code, 9001)
        XCTAssertEqual(SDKError.txTimeout.code, 9002)
        XCTAssertEqual(SDKError.txNotFound.code, 9003)
        XCTAssertEqual(SDKError.signingFailed("").code, 9004)
        XCTAssertEqual(SDKError.invalidConfig("").code, 9005)
        XCTAssertEqual(SDKError.queryFailed("").code, 9006)
        XCTAssertEqual(SDKError.invalidAddress.code, 9007)
    }

    func testNodeErrorCodes() {
        XCTAssertEqual(SDKError.nodeNotFound.code, 1101)
        XCTAssertEqual(SDKError.nodeAlreadyExists.code, 1102)
        XCTAssertEqual(SDKError.unauthorizedOperator.code, 1108)
        XCTAssertEqual(SDKError.unauthorizedAgentMsg.code, 1109)
    }

    func testActivityErrorCodes() {
        XCTAssertEqual(SDKError.submitterNotRegistered.code, 1200)
        XCTAssertEqual(SDKError.nodeNotEligible.code, 1201)
        XCTAssertEqual(SDKError.duplicateActivityHash.code, 1202)
        XCTAssertEqual(SDKError.quotaExceeded.code, 1203)
        XCTAssertEqual(SDKError.invalidActivityHash.code, 1204)
        XCTAssertEqual(SDKError.invalidContentURI.code, 1205)
    }

    func testFromABCICode() {
        XCTAssertEqual(SDKError.fromABCICode(1101), SDKError.nodeNotFound)
        XCTAssertEqual(SDKError.fromABCICode(1202), SDKError.duplicateActivityHash)
        XCTAssertEqual(SDKError.fromABCICode(1203), SDKError.quotaExceeded)
    }

    func testFromABCICodeUnknown() {
        let err = SDKError.fromABCICode(9999)
        XCTAssertEqual(err.code, 9999)
    }

    func testErrorDescription() {
        let desc = SDKError.nodeNotFound.errorDescription
        XCTAssertNotNil(desc)
        XCTAssertTrue(desc!.contains("1101"))
        XCTAssertTrue(desc!.contains("node not found"))
    }
}
