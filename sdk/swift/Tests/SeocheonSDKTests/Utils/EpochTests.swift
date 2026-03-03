import XCTest
@testable import SeocheonSDK

final class EpochTests: XCTestCase {

    func testComputeEpochBlock1() {
        XCTAssertEqual(EpochUtils.computeEpoch(blockHeight: 1), 0)
    }

    func testComputeEpochFirstBlockSecondEpoch() {
        XCTAssertEqual(EpochUtils.computeEpoch(blockHeight: 17281), 1)
    }

    func testComputeEpochLastBlockFirstEpoch() {
        XCTAssertEqual(EpochUtils.computeEpoch(blockHeight: 17280), 0)
    }

    func testComputeEpochLargeBlock() {
        XCTAssertEqual(EpochUtils.computeEpoch(blockHeight: 34561), 2)
    }

    func testComputeEpochZeroBlock() {
        XCTAssertEqual(EpochUtils.computeEpoch(blockHeight: 0), 0)
    }

    func testComputeWindowBlock1() {
        XCTAssertEqual(EpochUtils.computeWindow(blockHeight: 1), 0)
    }

    func testComputeWindowSecondWindow() {
        XCTAssertEqual(EpochUtils.computeWindow(blockHeight: 1441), 1)
    }

    func testComputeWindowLastBlockFirstWindow() {
        XCTAssertEqual(EpochUtils.computeWindow(blockHeight: 1440), 0)
    }

    func testComputeWindowLastWindow() {
        XCTAssertEqual(EpochUtils.computeWindow(blockHeight: 17280), 11)
    }

    func testEpochStartBlock() {
        XCTAssertEqual(EpochUtils.epochStartBlock(epochNumber: 0), 1)
        XCTAssertEqual(EpochUtils.epochStartBlock(epochNumber: 1), 17281)
        XCTAssertEqual(EpochUtils.epochStartBlock(epochNumber: 2), 34561)
    }

    func testEpochEndBlock() {
        XCTAssertEqual(EpochUtils.epochEndBlock(epochNumber: 0), 17280)
        XCTAssertEqual(EpochUtils.epochEndBlock(epochNumber: 1), 34560)
    }

    func testWindowStartBlock() {
        let epochStart: Int64 = 1
        XCTAssertEqual(EpochUtils.windowStartBlock(epochStartBlock: epochStart, windowIndex: 0), 1)
        XCTAssertEqual(EpochUtils.windowStartBlock(epochStartBlock: epochStart, windowIndex: 1), 1441)
    }

    func testWindowEndBlock() {
        XCTAssertEqual(EpochUtils.windowEndBlock(windowStart: 1), 1440)
        XCTAssertEqual(EpochUtils.windowEndBlock(windowStart: 1441), 2880)
    }
}
