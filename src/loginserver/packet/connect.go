package packet

import (
    "time"
    "share/network"
)

// Connect2Svr Packet
func Connect2Svr(session *network.Session, reader *network.Reader) {
    session.AuthKey = uint32(time.Now().Unix())

    var packet = network.NewWriter(CONNECT2SVR)
    packet.WriteUint32(session.Encryption.Key.Seed2nd)
    packet.WriteUint32(session.AuthKey)
    packet.WriteUint16(session.UserIdx)
    packet.WriteUint16(session.Encryption.RecvXorKeyIdx)

    session.Send(packet)
}

// CheckVersion Packet
func CheckVersion(session *network.Session, reader *network.Reader) {
    var version1 = reader.ReadInt32()

    var sessionData      = session.Data
    sessionData.Verified = true

    if version1 != int32(g_ServerConfig.Version) {
        log.Errorf("Client version mismatch (Client: %d, Server: %d)",
            version1,
            g_ServerConfig.Version,
        )

        sessionData.Verified = false
    }

    var packet = network.NewWriter(CHECKVERSION)
    packet.WriteInt32(g_ServerConfig.Version)
    packet.WriteInt32(0x00)
    packet.WriteInt32(0x00)
    packet.WriteInt32(0x00)

    session.Send(packet)
}