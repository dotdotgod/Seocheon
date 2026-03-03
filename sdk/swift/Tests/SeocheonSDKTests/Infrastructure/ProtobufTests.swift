import XCTest
@testable import SeocheonSDK

final class ProtobufTests: XCTestCase {

    func testEncodeVarintSmall() {
        let data = Protobuf.encodeVarint(1)
        XCTAssertEqual(data, Data([0x01]))
    }

    func testEncodeVarintZero() {
        let data = Protobuf.encodeVarint(0)
        XCTAssertEqual(data, Data([0x00]))
    }

    func testEncodeVarint150() {
        let data = Protobuf.encodeVarint(150)
        XCTAssertEqual(data, Data([0x96, 0x01]))
    }

    func testEncodeFieldVarint() {
        let data = Protobuf.encodeFieldVarint(1, value: 42)
        XCTAssertFalse(data.isEmpty)
        XCTAssertEqual(data[0], 0x08) // field 1, wire type 0
    }

    func testEncodeFieldBytes() {
        let payload = Data([0xDE, 0xAD])
        let data = Protobuf.encodeFieldBytes(2, data: payload)
        XCTAssertTrue(data.count > 2)
        XCTAssertEqual(data[0], 0x12) // field 2, wire type 2
    }

    func testEncodeFieldString() {
        let data = Protobuf.encodeFieldString(1, value: "hello")
        XCTAssertFalse(data.isEmpty)
    }
}
