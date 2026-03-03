import XCTest
@testable import SeocheonSDK

final class GasTests: XCTestCase {

    func testDefaultSubmitGas() {
        XCTAssertEqual(Gas.defaultGasSubmitActivity, 200_000)
    }

    func testDefaultWithdrawGas() {
        XCTAssertEqual(Gas.defaultGasWithdrawNodeCommission, 300_000)
    }

    func testDefaultSendGas() {
        XCTAssertEqual(Gas.defaultGasSend, 100_000)
    }

    func testDefaultFallbackGas() {
        XCTAssertEqual(Gas.defaultGasFallback, 200_000)
    }

    func testDefaultGasPrice() {
        XCTAssertEqual(Gas.defaultGasPrice, 250)
    }

    func testCalculateFee() {
        let fee = Gas.calculateFee(gasLimit: 200_000, gasPrice: 250)
        XCTAssertEqual(fee, 50_000_000)
    }
}
