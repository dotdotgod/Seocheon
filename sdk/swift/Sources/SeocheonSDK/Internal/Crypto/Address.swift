import Foundation
import CryptoKit

/// Bech32 address encoding for Cosmos-compatible addresses.
internal enum Address {
    /// Address HRP prefix.
    static let bech32Prefix = "seocheon"

    /// Derives a bech32 address from a compressed 33-byte public key.
    static func fromPubKey(_ pubKey: Data) throws -> String {
        guard pubKey.count == 33 else {
            throw SDKError.invalidAddress
        }

        // SHA-256 hash
        let sha = SHA256.hash(data: pubKey)

        // RIPEMD-160 hash
        let addrBytes = ripemd160(Data(sha))

        // Bech32 encode with "seocheon" HRP
        return try bech32Encode(hrp: bech32Prefix, data: addrBytes)
    }

    // MARK: - RIPEMD-160

    /// Pure Swift RIPEMD-160 implementation.
    static func ripemd160(_ message: Data) -> Data {
        var h0: UInt32 = 0x67452301
        var h1: UInt32 = 0xEFCDAB89
        var h2: UInt32 = 0x98BADCFE
        var h3: UInt32 = 0x10325476
        var h4: UInt32 = 0xC3D2E1F0

        // Pre-processing: adding padding bits
        var msg = Array(message)
        let msgLen = msg.count
        msg.append(0x80)
        while msg.count % 64 != 56 {
            msg.append(0x00)
        }
        // Length in bits as 64-bit little-endian
        let bitLen = UInt64(msgLen) * 8
        for i in 0..<8 {
            msg.append(UInt8((bitLen >> (i * 8)) & 0xFF))
        }

        let r: [Int] = [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,
                        7,4,13,1,10,6,15,3,12,0,9,5,2,14,11,8,
                        3,10,14,4,9,15,8,1,2,7,0,6,13,11,5,12,
                        1,9,11,10,0,8,12,4,13,3,7,15,14,5,6,2,
                        4,0,5,9,7,12,2,10,14,1,3,8,11,6,15,13]
        let rp: [Int] = [5,14,7,0,9,2,11,4,13,6,15,8,1,10,3,12,
                         6,11,3,7,0,13,5,10,14,15,8,12,4,9,1,2,
                         15,5,1,3,7,14,6,9,11,8,12,2,10,0,4,13,
                         8,6,4,1,3,11,15,0,5,12,2,13,9,7,10,14,
                         12,15,10,4,1,5,8,7,6,2,13,14,0,3,9,11]
        let s: [Int] = [11,14,15,12,5,8,7,9,11,13,14,15,6,7,9,8,
                        7,6,8,13,11,9,7,15,7,12,15,9,11,7,13,12,
                        11,13,6,7,14,9,13,15,14,8,13,6,5,12,7,5,
                        11,12,14,15,14,15,9,8,9,14,5,6,8,6,5,12,
                        9,15,5,11,6,8,13,12,5,12,13,14,11,8,5,6]
        let sp: [Int] = [8,9,9,11,13,15,15,5,7,7,8,11,14,14,12,6,
                         9,13,15,7,12,8,9,11,7,7,12,7,6,15,13,11,
                         9,7,15,11,8,6,6,14,12,13,5,14,13,13,7,5,
                         15,5,8,11,14,14,6,14,6,9,12,9,12,5,15,8,
                         8,5,12,9,12,5,14,6,8,13,6,5,15,13,11,11]

        func f(_ j: Int, _ x: UInt32, _ y: UInt32, _ z: UInt32) -> UInt32 {
            if j < 16 { return x ^ y ^ z }
            if j < 32 { return (x & y) | (~x & z) }
            if j < 48 { return (x | ~y) ^ z }
            if j < 64 { return (x & z) | (y & ~z) }
            return x ^ (y | ~z)
        }

        let k: [UInt32] = [0x00000000, 0x5A827999, 0x6ED9EBA1, 0x8F1BBCDC, 0xA953FD4E]
        let kp: [UInt32] = [0x50A28BE6, 0x5C4DD124, 0x6D703EF3, 0x7A6D76E9, 0x00000000]

        // Process each 64-byte block
        for blockStart in stride(from: 0, to: msg.count, by: 64) {
            var x = [UInt32](repeating: 0, count: 16)
            for i in 0..<16 {
                let offset = blockStart + i * 4
                x[i] = UInt32(msg[offset]) | (UInt32(msg[offset+1]) << 8) |
                       (UInt32(msg[offset+2]) << 16) | (UInt32(msg[offset+3]) << 24)
            }

            var al = h0, bl = h1, cl = h2, dl = h3, el = h4
            var ar = h0, br = h1, cr = h2, dr = h3, er = h4

            for j in 0..<80 {
                let round = j / 16
                let tl = al &+ f(j, bl, cl, dl) &+ x[r[j]] &+ k[round]
                let rotated = rotateLeft(tl, by: s[j]) &+ el
                al = el; el = dl; dl = rotateLeft(cl, by: 10); cl = bl; bl = rotated

                let tr = ar &+ f(79 - j, br, cr, dr) &+ x[rp[j]] &+ kp[round]
                let rotatedR = rotateLeft(tr, by: sp[j]) &+ er
                ar = er; er = dr; dr = rotateLeft(cr, by: 10); cr = br; br = rotatedR
            }

            let t = h1 &+ cl &+ dr
            h1 = h2 &+ dl &+ er
            h2 = h3 &+ el &+ ar
            h3 = h4 &+ al &+ br
            h4 = h0 &+ bl &+ cr
            h0 = t
        }

        var result = Data(count: 20)
        for (i, val) in [h0, h1, h2, h3, h4].enumerated() {
            result[i*4] = UInt8(val & 0xFF)
            result[i*4+1] = UInt8((val >> 8) & 0xFF)
            result[i*4+2] = UInt8((val >> 16) & 0xFF)
            result[i*4+3] = UInt8((val >> 24) & 0xFF)
        }
        return result
    }

    private static func rotateLeft(_ x: UInt32, by n: Int) -> UInt32 {
        return (x << n) | (x >> (32 - n))
    }

    // MARK: - Bech32

    static func bech32Encode(hrp: String, data: Data) throws -> String {
        let converted = try convertBits(data: Array(data), fromBits: 8, toBits: 5, pad: true)

        var values = converted + [0, 0, 0, 0, 0, 0]
        let polymod = bech32Polymod(expandHRP(hrp) + values) ^ 1
        for i in 0..<6 {
            values[converted.count + i] = UInt8((polymod >> (5 * (5 - i))) & 31)
        }

        let charset: [Character] = Array("qpzry9x8gf2tvdw0s3jn54khce6mua7l")
        var result = hrp + "1"
        for v in values {
            result.append(charset[Int(v)])
        }
        return result
    }

    private static func convertBits(data: [UInt8], fromBits: UInt, toBits: UInt, pad: Bool) throws -> [UInt8] {
        var acc: UInt32 = 0
        var bits: UInt = 0
        var result: [UInt8] = []
        let maxV: UInt32 = (1 << toBits) - 1

        for b in data {
            acc = (acc << fromBits) | UInt32(b)
            bits += fromBits
            while bits >= toBits {
                bits -= toBits
                result.append(UInt8((acc >> bits) & maxV))
            }
        }

        if pad {
            if bits > 0 {
                result.append(UInt8((acc << (toBits - bits)) & maxV))
            }
        }

        return result
    }

    private static func expandHRP(_ hrp: String) -> [UInt8] {
        var result: [UInt8] = []
        for c in hrp.unicodeScalars {
            result.append(UInt8(c.value >> 5))
        }
        result.append(0)
        for c in hrp.unicodeScalars {
            result.append(UInt8(c.value & 31))
        }
        return result
    }

    private static func bech32Polymod(_ values: [UInt8]) -> UInt32 {
        let gen: [UInt32] = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3]
        var chk: UInt32 = 1
        for v in values {
            let top = chk >> 25
            chk = (chk & 0x1ffffff) << 5 ^ UInt32(v)
            for i in 0..<5 {
                if (top >> i) & 1 == 1 {
                    chk ^= gen[i]
                }
            }
        }
        return chk
    }
}
