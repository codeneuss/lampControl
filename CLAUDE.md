# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LampControl is a cross-platform tool for controlling ELK-BLEDOM LED lamps via Bluetooth. It provides both a CLI interface and a web-based UI with real-time WebSocket communication.

## Build & Development Commands

```bash
# Build
make build                  # Build binary to bin/lamp
make build-release         # Build optimized release binary

# Testing
make test                  # Run all tests
make test-protocol         # Run protocol tests only
make test-coverage         # Generate coverage report (coverage.html)

# Code Quality
make fmt                   # Format all Go code
make lint                  # Run golangci-lint (must be installed)
make tidy                  # Tidy Go module dependencies

# Web Server
./bin/lamp web --port 8080 --host localhost
```

## Architecture

### Clean Architecture Layers

The codebase follows Clean Architecture with strict dependency rules (inner layers never depend on outer layers):

```
Domain (innermost) → Application → Infrastructure → Presentation (outermost)
```

**Domain Layer** (`internal/domain/`):
- Pure business entities with no external dependencies
- `Device`: Represents a BLE lamp device with connection state and current state (power, brightness, RGB, etc.)
- `CustomEffect`: User-defined lighting effects with colors, pattern, and speed
- Domain errors defined in `errors.go`

**Application Layer** (`internal/application/`):
- `DeviceService`: Orchestrates device operations (scan, connect, command execution)
- Thread-safe with `sync.RWMutex` for concurrent access
- Manages device registry and state tracking

**Infrastructure Layer** (`internal/infrastructure/`):
- `bluetooth/adapter.go`: BLE communication using tinygo-org/bluetooth
  - `Scan()`: Discovers devices (returns empty array if already connected)
  - `Connect()`: Establishes connection and enables notifications
  - `SendCommand()`: Writes 9-byte protocol commands (3 retries with delays)
- `storage/effect_storage.go`: JSON file persistence for custom effects in `~/.lampcontrol/custom_effects.json`

**Presentation Layer** (`internal/presentation/`):
- CLI commands in `cmd/` (Cobra framework)
- Web API in `internal/presentation/api/`:
  - REST endpoints for device management and custom effects
  - WebSocket hub for real-time bidirectional communication
  - Middleware: logging (with Hijacker support for WebSocket), CORS, recovery

### ELK-BLEDOM Protocol (`pkg/protocol/`)

9-byte command format: `[0x7E] [0x00] [CMD] [PARAMS...] [0xEF]`

**Critical Implementation Details**:
- Power command uses `CmdPower = 0x04` (not 0x08)
- Power ON bytes: `[0xF0, 0x00, 0x01]`
- Power OFF bytes: `[0x00, 0x00, 0x00]`
- RGB command: `[0x7E, 0x00, 0x05, 0x03, R, G, B, 0x00, 0xEF]`
- All commands sent 3 times with 20ms delay (lamp reliability)
- Service UUID: `0000fff0-0000-1000-8000-00805f9b34fb`
- Characteristic UUID: `0000fff3-0000-1000-8000-00805f9b34fb`

See tests in `pkg/protocol/elkbledom_test.go` for expected byte patterns.

### Web UI (`web/static/`)

**Frontend Architecture** (Vanilla JavaScript, no framework):
- Single-page application with glassmorphism design (Apple-inspired)
- WebSocket client with auto-reconnect and exponential backoff
- Controllers for: Device, Power, Brightness, Color, WhiteBalance, Effects, Sound
- Throttling (50ms) on all slider inputs to prevent command flooding
- localStorage caching:
  - `lastDeviceAddress`: Auto-reconnects on page reload
  - `cachedDevices`: Instant device list without scanning

**State Management**:
- Frontend state syncs with backend via WebSocket `state_update` messages
- Backend broadcasts state changes to all connected clients
- On device selection, UI immediately reflects current lamp state

**Sound-Reactive Mode**:
- Uses Web Audio API with microphone access
- Maps audio frequencies to RGB (bass→red, mid→green, high→blue)
- Real-time visualization with HTML5 canvas

**Custom Effects**:
- Stored persistently in backend (`~/.lampcontrol/custom_effects.json`)
- Modal-based creation UI
- Support for multiple colors with patterns (fade, jump, strobe, pulse)

### WebSocket Protocol

**Client → Server**:
```json
{
  "type": "command",
  "action": "power|color|brightness|white_balance|effect",
  "payload": { /* command-specific data */ }
}
```

**Server → Client**:
```json
{
  "type": "state_update",
  "device": {
    "address": "...",
    "state": {
      "power_on": true,
      "brightness": 255,
      "rgb": {"r": 255, "g": 0, "b": 0}
    }
  }
}
```

## Critical Implementation Notes

### Bluetooth Scanning Behavior
- **Connected devices do NOT appear in scan results** (normal BLE behavior)
- Solution: Cache device list in localStorage and skip auto-scan on reload if cached device exists
- Only scan when: no cached devices OR user manually triggers scan via dropdown

### WebSocket Message Handling
- DO NOT batch multiple messages with newlines (causes JSON parse errors)
- Send each message as a separate WebSocket frame
- Logging middleware MUST implement `http.Hijacker` interface for WebSocket upgrade

### Command Delays
- Originally: 1000ms post-command delay (too slow)
- Optimized: 50ms post-command delay
- Between retries: 20ms (sufficient for reliability)

### State Persistence
- Device connection state: Auto-reconnects using `lastDeviceAddress` from localStorage
- Custom effects: Persisted in `~/.lampcontrol/custom_effects.json`
- No database required—simple file-based storage

## Web Server Implementation

The web server combines REST API and WebSocket:
- Chi router for HTTP routing
- Gorilla WebSocket for real-time communication
- Hub pattern for WebSocket client management (broadcast channels)
- Server graceful shutdown on SIGINT/SIGTERM

**Server State** (`internal/presentation/api/state/`):
- Thread-safe with `sync.RWMutex`
- Stores currently selected device address
- References to DeviceService and WebSocket Hub
- Broadcasts state changes to all connected clients

## Testing Protocol Changes

When modifying protocol commands:
1. Update constants in `pkg/protocol/elkbledom.go`
2. Update expected bytes in `pkg/protocol/elkbledom_test.go`
3. Run `make test-protocol` to verify
4. Test with real device before committing

Example test format:
```go
expected := Command{0x7E, 0x00, 0x04, 0xF0, 0x00, 0x01, 0xFF, 0x00, 0xEF}
actual := NewPowerCommand(true)
assert.Equal(t, expected, actual)
```

## API Endpoints

**Device Management**:
- `GET /api/health` - Health check
- `GET /api/devices` - List discovered devices
- `POST /api/scan` - Scan for devices (5s timeout)
- `POST /api/device/select` - Select and connect to device
- `GET /api/device/current` - Get current device with state

**Custom Effects**:
- `GET /api/effects` - List all custom effects
- `POST /api/effects` - Create new effect
- `DELETE /api/effects/{id}` - Delete effect

**WebSocket**:
- `GET /ws` - WebSocket endpoint for bidirectional communication

## Common Pitfalls

1. **Empty scan results**: If device is already connected, it won't appear in scan. This is expected.
2. **WebSocket upgrade fails**: Ensure logging middleware implements `http.Hijacker`.
3. **Power commands not working**: Verify command code is `0x04` and byte patterns match tests.
4. **UI not syncing state**: Check that `updateUIFromDeviceState()` is called after device selection.
5. **Throttling not working**: Use `throttle()` not `debounce()` for sliders—throttle ensures consistent rate.

## Platform-Specific Notes

**macOS**:
- Uses CoreBluetooth (stable)
- Requires Bluetooth permissions in System Preferences
- Xcode Command Line Tools required for build

**Linux**:
- Uses BlueZ D-Bus backend (stable)
- BlueZ 5.48+ required (usually pre-installed)
- May require `sudo` for Bluetooth access

**Windows**:
- Experimental support
- Basic scanning only
