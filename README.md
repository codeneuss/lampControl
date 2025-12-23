# LampControl

A cross-platform CLI tool for controlling duoCo StripX LED lamps via Bluetooth using the ELK-BLEDOM protocol.

## Features

- **Device Discovery**: Scan for ELK-BLEDOM compatible LED devices
- **Power Control**: Turn lamps on/off
- **RGB Color Control**: Set any RGB color (0-255 per channel)
- **Brightness Control**: Adjust brightness (0-255)
- **White Balance**: Control warm/cold white balance
- **Effects**: Trigger built-in effects and scenes
- **Cross-Platform**: Supports macOS and Linux (Windows experimental)

## Installation

### Prerequisites

- Go 1.21 or later
- Bluetooth adapter
- **macOS**: Xcode Command Line Tools
- **Linux**: BlueZ 5.48+ (usually pre-installed)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/codeneuss/lampcontrol.git
cd lampcontrol

# Build the binary
make build

# Or using go directly
go build -o bin/lamp ./cmd
```

## Usage

### Scan for Devices

Find available ELK-BLEDOM devices:

```bash
lamp scan
```

Example output:
```
Scanning for devices (timeout: 10s)...

Found 2 device(s):

1. ELK-BLEDOM
   Address: AA:BB:CC:DD:EE:FF
   RSSI: -45 dBm

2. LED Strip
   Address: 11:22:33:44:55:66
   RSSI: -52 dBm
```

### Control Power

Turn device on or off:

```bash
# Turn on
lamp power on --device AA:BB:CC:DD:EE:FF

# Turn off
lamp power off -d AA:BB:CC:DD:EE:FF
```

### Set RGB Color

Set any RGB color:

```bash
# Red
lamp color --device AA:BB:CC:DD:EE:FF --rgb 255,0,0

# Green
lamp color -d AA:BB:CC:DD:EE:FF -r 0,255,0

# Blue
lamp color -d AA:BB:CC:DD:EE:FF -r 0,0,255

# Purple
lamp color -d AA:BB:CC:DD:EE:FF -r 128,0,128
```

### Set Brightness

Adjust brightness (0-255):

```bash
# Full brightness
lamp brightness --device AA:BB:CC:DD:EE:FF --level 255

# 50% brightness
lamp brightness -d AA:BB:CC:DD:EE:FF -l 128

# Dim
lamp brightness -d AA:BB:CC:DD:EE:FF -l 50
```

### Set White Balance

Control warm/cold white:

```bash
# 50/50 balance
lamp white --device AA:BB:CC:DD:EE:FF --warm 128 --cold 128

# Full warm
lamp white -d AA:BB:CC:DD:EE:FF -w 255 -c 0

# Full cold
lamp white -d AA:BB:CC:DD:EE:FF -w 0 -c 255
```

### Set Effect

Trigger built-in effects:

```bash
# Effect 1 with speed 50
lamp effect --device AA:BB:CC:DD:EE:FF --index 1 --speed 50

# Effect 5 with fast speed
lamp effect -d AA:BB:CC:DD:EE:FF -i 5 -s 200
```

## Development

### Project Structure

```
lampcontrol/
├── cmd/                    # CLI commands
├── internal/
│   ├── domain/            # Business logic
│   ├── application/       # Use cases
│   ├── infrastructure/    # BLE adapter, config
│   └── presentation/      # CLI/API handlers
├── pkg/
│   └── protocol/          # ELK-BLEDOM protocol
└── tests/                 # Tests
```

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run protocol tests only
go test ./pkg/protocol/... -v
```

### Build

```bash
# Build for current platform
make build

# Clean build artifacts
make clean
```

## Protocol Documentation

This tool implements the ELK-BLEDOM protocol:

- **Service UUID**: `0000fff0-0000-1000-8000-00805f9b34fb`
- **Characteristic UUID**: `0000fff3-0000-1000-8000-00805f9b34fb`
- **Command Format**: 9-byte packets `[0x7E] [0x00] [CMD] [PARAMS...] [0xEF]`

For detailed protocol information, see [ELK-BLEDOM Protocol](https://github.com/FergusInLondon/ELK-BLEDOM/blob/master/PROTCOL.md).

## Platform Support

| Platform | Status | Notes |
|----------|--------|-------|
| macOS    | ✅ Stable | CoreBluetooth backend |
| Linux    | ✅ Stable | BlueZ D-Bus backend |
| Windows  | ⚠️ Experimental | Basic scanning only |

## Troubleshooting

### Bluetooth Permission Issues (macOS)

Grant Bluetooth permissions to your terminal application in:
`System Preferences → Security & Privacy → Bluetooth`

### Device Not Found

- Ensure the lamp is powered on
- Move closer to the device
- Try increasing scan timeout: `lamp scan --timeout 30s`

### Connection Timeout

- Check if another device is connected to the lamp
- Power cycle the lamp
- Reduce distance to the device

## Roadmap

### Phase 2: REST API (Upcoming)
- [ ] HTTP REST API server
- [ ] WebSocket support for real-time updates
- [ ] OpenAPI specification

### Phase 3: Advanced Features
- [ ] Custom effect programming
- [ ] Scene management
- [ ] Scheduling/timers
- [ ] Multi-device control

### Phase 4: Distribution
- [ ] Pre-built binaries (macOS, Linux, Windows)
- [ ] Homebrew formula
- [ ] Debian package
- [ ] Docker image

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [ELK-BLEDOM Protocol Documentation](https://github.com/FergusInLondon/ELK-BLEDOM) by FergusInLondon
- [tinygo-org/bluetooth](https://github.com/tinygo-org/bluetooth) - Go Bluetooth library
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## References

- [ELK-BLEDOM Protocol Spec](https://github.com/FergusInLondon/ELK-BLEDOM/blob/master/PROTCOL.md)
- [Linux BLE Control Guide](https://linuxthings.co.uk/blog/control-an-elk-bledom-bluetooth-led-strip)
- [Reverse Engineering BLE Devices](https://urish.medium.com/reverse-engineering-a-bluetooth-lightbulb-56580fcb7546)
