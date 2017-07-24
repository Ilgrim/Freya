package packet

import (
    "bytes"
    "share/network"
    "share/rpc"
    "share/models/subpasswd"
    "share/models/character"
)

// GetMyChartr Packet
func GetMyChartr(session *network.Session, reader *network.Reader) {
    if !session.Data.Verified {
        log.Error("Unauthorized connection from", session.GetEndPnt())
        session.Close()
        return
    }

    // fetch subpassword
    var req = subpasswd.FetchReq{session.Data.AccountId}
    var res = subpasswd.FetchRes{}
    g_RPCHandler.Call(rpc.FetchSubPassword, req, &res)

    session.Data.SubPassword = &res.Details

    var subpasswdExist = 0
    if res.Password != "" {
        subpasswdExist = 1
    }

    // fetch characters
    var reqList = character.ListReq{session.Data.AccountId, byte(g_ServerSettings.ServerId)}
    var resList = character.ListRes{}
    g_RPCHandler.Call(rpc.LoadCharacters, reqList, &resList)

    session.Data.CharacterList = resList.List

    var packet = network.NewWriter(GETMYCHARTR)
    packet.WriteInt32(subpasswdExist)
    packet.WriteBytes(make([]byte, 10))
    packet.WriteInt32(0x00) // selected character id
    packet.WriteInt32(0x00) // slot order

    for i := 0; i < len(resList.List); i ++ {
        var char = resList.List[i]
        packet.WriteInt32(char.Id)
        packet.WriteInt64(char.Created.Unix())
        packet.WriteUint32(char.Style.Get())
        packet.WriteUint32(char.Level)
        packet.WriteByte(char.SwordRank)
        packet.WriteByte(char.MagicRank)
        packet.WriteInt16(0x00)     // padding for skill ranks
        packet.WriteUint64(char.Alz)
        packet.WriteByte(char.Nation)
        packet.WriteByte(char.World)
        packet.WriteUint16(char.X)
        packet.WriteUint16(char.Y)
        packet.WriteBytes(char.Equipment.SerializeKind())
        packet.WriteBytes(make([]byte, 88))
        packet.WriteByte(len(char.Name) + 1)
        packet.WriteString(char.Name + "\x00")
    }

    session.Send(packet)
}

// NewMyChartr Packet
func NewMyChartr(session *network.Session, reader *network.Reader) {
    var style        = reader.ReadUint32()
    var _            = reader.ReadByte() // beginner join guild
    var slot         = reader.ReadByte()
    var nameLength   = reader.ReadByte()
    var name         = string(bytes.Trim(reader.ReadBytes(int(nameLength)), "\x00"))

    var charId   = session.Data.AccountId * 8 + int32(slot)
    var newStyle = character.Style{}
    newStyle.Set(style)

    var packet = network.NewWriter(NEWMYCHARTR)

    if !newStyle.Verify() || slot > 5 || nameLength > 16 {
        // invalid style, slot or nameLength
        packet.WriteInt32(0x00)
        packet.WriteByte(character.NowAllowed)

        session.Send(packet)
        return
    }

    // check if slot is used
    var charList = session.Data.CharacterList
    for i := 0; i < len(charList); i ++ {
        if charList[i].Id == charId {
            packet.WriteInt32(0x00)
            packet.WriteByte(character.SlotInUse)

            session.Send(packet)
            return
        }
    }

    var req = character.CreateReq{
        byte(g_ServerSettings.ServerId),
        character.Character{Id: charId, Name: name, Style: newStyle},
    }
    var res = character.CreateRes{}
    g_RPCHandler.Call(rpc.CreateCharacter, req, &res)


    if res.Result == character.Success {
        packet.WriteInt32(charId)
        packet.WriteByte(res.Result + newStyle.BattleStyle)
        // update character with it's data
        session.Data.CharacterList = append(session.Data.CharacterList, res.Character)
    } else {
        packet.WriteInt32(0x00)
        packet.WriteByte(res.Result)
    }

    session.Send(packet)
}