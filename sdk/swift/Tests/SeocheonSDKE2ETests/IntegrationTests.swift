import XCTest
@testable import SeocheonSDK

/// E2E integration tests for the Seocheon Swift SDK.
///
/// Skip conditions:
///   - SEOCHEON_GRPC not set
///   - SEOCHEON_MNEMONIC not set
///
/// Run with:
///   swift test --filter E2E
final class IntegrationTests: XCTestCase {

    // MARK: - Environment

    private var grpc: String { ProcessInfo.processInfo.environment["SEOCHEON_GRPC"] ?? "" }
    private var mnemonic: String { ProcessInfo.processInfo.environment["SEOCHEON_MNEMONIC"] ?? "" }
    private var rpc: String { ProcessInfo.processInfo.environment["SEOCHEON_RPC"] ?? "http://localhost:26657" }
    private var chainID: String { ProcessInfo.processInfo.environment["SEOCHEON_CHAIN_ID"] ?? "seocheon-e2e" }

    // MARK: - Helpers

    private func skipIfMissing() throws {
        guard !grpc.isEmpty else {
            throw XCTSkip("E2E 스킵: SEOCHEON_GRPC 미설정")
        }
        guard !mnemonic.isEmpty else {
            throw XCTSkip("E2E 스킵: SEOCHEON_MNEMONIC 미설정")
        }
    }

    private func buildSDK() throws -> SeocheonSDK {
        let config = SDKConfig.defaultConfig(
            chainId: chainID,
            rpcEndpoint: rpc,
            grpcEndpoint: grpc,
            signing: SigningConfig(mode: .direct, mnemonic: mnemonic)
        )
        return try SeocheonSDK(config: config)
    }

    // MARK: - Tests

    func testConnect() async throws {
        try skipIfMissing()
        let sdk = try buildSDK()

        try await sdk.connect()
        defer { Task { await sdk.disconnect() } }

        XCTAssertTrue(sdk.isConnected(), "Connect 후 isConnected() = false")
    }

    func testGetLatestBlock() async throws {
        try skipIfMissing()
        let sdk = try buildSDK()

        try await sdk.connect()
        defer { Task { await sdk.disconnect() } }

        let block = try await sdk.cosmos.getBlockInfo()
        XCTAssertGreaterThan(block.blockHeight, 0, "블록 높이가 양수여야 함")
        print("최신 블록: height=\(block.blockHeight) chainId=\(block.chainId)")
    }

    func testQueryNodeModule() async throws {
        try skipIfMissing()
        let sdk = try buildSDK()

        try await sdk.connect()
        defer { Task { await sdk.disconnect() } }

        let resp = try await sdk.node.search(tag: "", status: "", limit: 10, orderBy: "asc")
        print("x/node 조회 성공: total=\(resp.total)")
        _ = resp // 결과 무시 없이 성공만 확인
    }

    func testQueryEpochInfo() async throws {
        try skipIfMissing()
        let sdk = try buildSDK()

        try await sdk.connect()
        defer { Task { await sdk.disconnect() } }

        let info = try await sdk.epoch.getInfo()
        XCTAssertGreaterThan(info.blockHeight, 0, "에포크 블록 높이가 양수여야 함")
        print("에포크: epoch=\(info.epochNumber) window=\(info.windowNumber) height=\(info.blockHeight)")
    }
}
