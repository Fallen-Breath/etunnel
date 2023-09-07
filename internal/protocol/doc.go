package protocol

// Stream format: [ Magic ] [ Address data ] [ Payload ... )
// Protocol (1 Byte):
//   bit 0~3: protocol type
//   bit 4~7: address type
// Address data:
//   ipv4: 4 bytes
//   ipv6: 16 bytes
//   hostname: 1 byte length + bytes
//   string: varint length + bytes
// Payload:
//   tcp: the tcp stream
//   udp: 2 bytes length + packet buffer