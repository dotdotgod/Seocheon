import XCTest
@testable import SeocheonSDK

final class DenomTests: XCTestCase {

    func testFormatKkotZero() {
        XCTAssertEqual(ConvertUtils.formatKkot(0), "0.0000000000")
    }

    func testFormatKkotOneKkot() {
        XCTAssertEqual(ConvertUtils.formatKkot(10_000_000_000), "1.0000000000")
    }

    func testFormatKkotFractional() {
        XCTAssertEqual(ConvertUtils.formatKkot(5_000_000_000), "0.5000000000")
    }

    func testFormatKkotLargeAmount() {
        XCTAssertEqual(ConvertUtils.formatKkot(123_456_789_012_345), "12345.6789012345")
    }

    func testParseKkotWhole() throws {
        let result = try ConvertUtils.parseKkot("1.0000000000")
        XCTAssertEqual(result, 10_000_000_000)
    }

    func testParseKkotFractional() throws {
        let result = try ConvertUtils.parseKkot("0.5000000000")
        XCTAssertEqual(result, 5_000_000_000)
    }

    func testParseKkotNoDecimal() throws {
        let result = try ConvertUtils.parseKkot("5")
        XCTAssertEqual(result, 50_000_000_000)
    }

    func testConvertDenomSameUnit() throws {
        let result = try ConvertUtils.convertDenom(amount: 100, from: "uppyeo", to: "uppyeo")
        XCTAssertEqual(result, 100)
    }

    func testConvertDenomUppyeoToSal() throws {
        let result = try ConvertUtils.convertDenom(amount: 200, from: "uppyeo", to: "sal")
        XCTAssertEqual(result, 2)
    }

    func testConvertDenomUnknown() {
        XCTAssertThrowsError(try ConvertUtils.convertDenom(amount: 1, from: "invalid", to: "uppyeo"))
    }

    func testDenomNames() {
        XCTAssertEqual(ConvertUtils.denomNames.count, 6)
        XCTAssertEqual(ConvertUtils.denomNames.first, "uppyeo")
        XCTAssertEqual(ConvertUtils.denomNames.last, "kkot")
    }
}
