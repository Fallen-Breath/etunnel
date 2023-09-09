package proto

// Request Stream Format: [ Magic ] [ Tunnel ID ] [ Payload ... )
// Magic (1 Byte):
//   marker | tunnel type
//   marker: 0b10110000 (176)
//   tunnel type:
//     stream: 0
//     packet: 1
// Tunnel ID:
//   1 byte length n + n bytes as string
// Payload:
//   tcp: the tcp stream
//   udp: 2 bytes length + packet buffer
//
// Response Stream Format: [ Response ] [ Payload ... )
// Response (1 Byte):
//   ok: 0
//   bad id: 1
//   oversize: 2 (e.g. packet too large for udp)
// Payload:
//   tcp: the tcp stream
//   udp: N/A
//
