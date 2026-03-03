import Foundation

/// Signs transactions via an external vault server (production).
public final class VaultSigningService: SigningService, @unchecked Sendable {
    private let endpoint: String
    private let keyName: String
    private var _address: String = ""
    private var _pubKeyData: Data = Data()
    private let session: URLSession
    private var initialized = false

    public init(endpoint: String, keyName: String) {
        self.endpoint = endpoint.hasSuffix("/") ? String(endpoint.dropLast()) : endpoint
        self.keyName = keyName
        self.session = URLSession(configuration: .default)
    }

    private func ensureInitialized() async throws {
        guard !initialized else { return }

        // Fetch address
        let addrURL = URL(string: "\(endpoint)/v1/keys/\(keyName)/address")!
        let (addrData, _) = try await session.data(from: addrURL)
        let addrResult = try JSONDecoder().decode(VaultAddressResponse.self, from: addrData)
        _address = addrResult.address

        // Fetch public key
        let pubKeyURL = URL(string: "\(endpoint)/v1/keys/\(keyName)/pubkey")!
        let (pubKeyResponseData, _) = try await session.data(from: pubKeyURL)
        let pubKeyResult = try JSONDecoder().decode(VaultPubKeyResponse.self, from: pubKeyResponseData)
        guard let pubKeyBytes = hexToData(pubKeyResult.pubkey) else {
            throw SDKError.signingFailed("invalid vault pubkey hex")
        }
        _pubKeyData = pubKeyBytes
        initialized = true
    }

    public func sign(_ txBytes: Data) async throws -> Data {
        try await ensureInitialized()

        let url = URL(string: "\(endpoint)/v1/keys/\(keyName)/sign")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let body = VaultSignRequest(data: txBytes.map { String(format: "%02x", $0) }.joined())
        request.httpBody = try JSONEncoder().encode(body)

        let (responseData, response) = try await session.data(for: request)
        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw SDKError.signingFailed("vault sign failed: \(String(data: responseData, encoding: .utf8) ?? "unknown")")
        }

        let signResult = try JSONDecoder().decode(VaultSignResponse.self, from: responseData)
        guard let sigBytes = hexToData(signResult.signature) else {
            throw SDKError.signingFailed("invalid signature hex from vault")
        }
        return sigBytes
    }

    public func getAddress() -> String {
        return _address
    }

    public func getPubKey() -> Data {
        return _pubKeyData
    }

    // MARK: - Private types

    private struct VaultSignRequest: Codable {
        let data: String
    }

    private struct VaultSignResponse: Codable {
        let signature: String
    }

    private struct VaultAddressResponse: Codable {
        let address: String
    }

    private struct VaultPubKeyResponse: Codable {
        let pubkey: String
    }

    private func hexToData(_ hex: String) -> Data? {
        var data = Data()
        var index = hex.startIndex
        while index < hex.endIndex {
            let nextIndex = hex.index(index, offsetBy: 2, limitedBy: hex.endIndex) ?? hex.endIndex
            guard nextIndex != index else { return nil }
            let byteString = hex[index..<nextIndex]
            guard let byte = UInt8(byteString, radix: 16) else { return nil }
            data.append(byte)
            index = nextIndex
        }
        return data
    }
}
