// LampControl Web UI - Main Application
// WebSocket-based LED lamp controller with color picker, brightness, white balance, and effects

// ===== Configuration =====
const WS_URL = `ws://${window.location.host}/ws`;
const API_URL = `/api`;

// ===== Utility Functions =====
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function $(selector) {
    return document.querySelector(selector);
}

function $$(selector) {
    return document.querySelectorAll(selector);
}

// ===== WebSocket Client =====
class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.messageQueue = [];
        this.listeners = {};
        this.connect();
    }

    connect() {
        console.log('Connecting to WebSocket...');
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => this.onOpen();
        this.ws.onmessage = (event) => this.onMessage(event);
        this.ws.onclose = () => this.onClose();
        this.ws.onerror = (error) => this.onError(error);
    }

    onOpen() {
        console.log('WebSocket connected');
        this.reconnectDelay = 1000;
        this.emit('connected');

        // Send queued messages
        while (this.messageQueue.length > 0) {
            const message = this.messageQueue.shift();
            this.send(message);
        }
    }

    onMessage(event) {
        try {
            const message = JSON.parse(event.data);
            this.emit('message', message);

            // Emit specific event types
            if (message.type) {
                this.emit(message.type, message);
            }
        } catch (error) {
            console.error('Failed to parse message:', error);
        }
    }

    onClose() {
        console.log('WebSocket disconnected');
        this.emit('disconnected');
        this.reconnect();
    }

    onError(error) {
        console.error('WebSocket error:', error);
        this.emit('error', error);
    }

    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            this.messageQueue.push(message);
        }
    }

    sendCommand(action, payload) {
        this.send({
            type: 'command',
            action: action,
            payload: payload
        });
    }

    reconnect() {
        setTimeout(() => {
            console.log('Reconnecting...');
            this.connect();
            this.reconnectDelay = Math.min(
                this.reconnectDelay * 2,
                this.maxReconnectDelay
            );
        }, this.reconnectDelay);
    }

    on(event, callback) {
        if (!this.listeners[event]) {
            this.listeners[event] = [];
        }
        this.listeners[event].push(callback);
    }

    emit(event, data) {
        if (this.listeners[event]) {
            this.listeners[event].forEach(callback => callback(data));
        }
    }
}

// ===== State Manager =====
class StateManager {
    constructor() {
        this.state = {
            connected: false,
            selectedDevice: null,
            devices: [],
            deviceState: {
                power_on: false,
                brightness: 255,
                rgb: { r: 255, g: 0, b: 0 },
                white_balance: null,
                effect: null,
                effect_speed: null
            }
        };
        this.listeners = [];
    }

    setState(newState) {
        this.state = { ...this.state, ...newState };
        this.notify();
    }

    updateDeviceState(deviceState) {
        this.state.deviceState = { ...this.state.deviceState, ...deviceState };
        this.notify();
    }

    subscribe(listener) {
        this.listeners.push(listener);
    }

    notify() {
        this.listeners.forEach(listener => listener(this.state));
    }

    getState() {
        return this.state;
    }
}

// ===== Color Utilities =====
class ColorUtils {
    static hsvToRgb(h, s, v) {
        h = h / 360;
        s = s / 100;
        v = v / 100;

        let r, g, b;
        const i = Math.floor(h * 6);
        const f = h * 6 - i;
        const p = v * (1 - s);
        const q = v * (1 - f * s);
        const t = v * (1 - (1 - f) * s);

        switch (i % 6) {
            case 0: r = v; g = t; b = p; break;
            case 1: r = q; g = v; b = p; break;
            case 2: r = p; g = v; b = t; break;
            case 3: r = p; g = q; b = v; break;
            case 4: r = t; g = p; b = v; break;
            case 5: r = v; g = p; b = q; break;
        }

        return {
            r: Math.round(r * 255),
            g: Math.round(g * 255),
            b: Math.round(b * 255)
        };
    }

    static rgbToHsv(r, g, b) {
        r = r / 255;
        g = g / 255;
        b = b / 255;

        const max = Math.max(r, g, b);
        const min = Math.min(r, g, b);
        const diff = max - min;

        let h = 0;
        if (diff !== 0) {
            if (max === r) {
                h = 60 * (((g - b) / diff) % 6);
            } else if (max === g) {
                h = 60 * ((b - r) / diff + 2);
            } else {
                h = 60 * ((r - g) / diff + 4);
            }
        }
        if (h < 0) h += 360;

        const s = max === 0 ? 0 : (diff / max) * 100;
        const v = max * 100;

        return { h, s, v };
    }
}

// ===== Color Picker =====
class ColorPicker {
    constructor(canvas, onColorChange) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.onColorChange = onColorChange;
        this.size = canvas.width;
        this.center = this.size / 2;
        this.radius = this.size / 2 - 10;

        this.isDragging = false;
        this.currentColor = { r: 255, g: 0, b: 0 };

        this.draw();
        this.attachEvents();
    }

    draw() {
        // Clear canvas
        this.ctx.clearRect(0, 0, this.size, this.size);

        // Draw HSV color wheel
        for (let angle = 0; angle < 360; angle += 1) {
            for (let r = 0; r < this.radius; r += 1) {
                const sat = (r / this.radius) * 100;
                const val = 100;
                const color = ColorUtils.hsvToRgb(angle, sat, val);

                const rad = (angle * Math.PI) / 180;
                const x = this.center + r * Math.cos(rad);
                const y = this.center + r * Math.sin(rad);

                this.ctx.fillStyle = `rgb(${color.r}, ${color.g}, ${color.b})`;
                this.ctx.fillRect(x, y, 2, 2);
            }
        }

        // Draw brightness gradient in center
        const gradient = this.ctx.createRadialGradient(
            this.center, this.center, 0,
            this.center, this.center, this.radius * 0.3
        );
        gradient.addColorStop(0, 'white');
        gradient.addColorStop(1, 'transparent');

        this.ctx.fillStyle = gradient;
        this.ctx.beginPath();
        this.ctx.arc(this.center, this.center, this.radius * 0.3, 0, Math.PI * 2);
        this.ctx.fill();
    }

    attachEvents() {
        this.canvas.addEventListener('mousedown', (e) => this.handleStart(e));
        this.canvas.addEventListener('mousemove', (e) => this.handleMove(e));
        this.canvas.addEventListener('mouseup', () => this.handleEnd());
        this.canvas.addEventListener('mouseleave', () => this.handleEnd());

        // Touch events
        this.canvas.addEventListener('touchstart', (e) => {
            e.preventDefault();
            this.handleStart(e.touches[0]);
        });
        this.canvas.addEventListener('touchmove', (e) => {
            e.preventDefault();
            this.handleMove(e.touches[0]);
        });
        this.canvas.addEventListener('touchend', () => this.handleEnd());
    }

    handleStart(e) {
        this.isDragging = true;
        this.updateColor(e);
    }

    handleMove(e) {
        if (this.isDragging) {
            this.updateColor(e);
        }
    }

    handleEnd() {
        this.isDragging = false;
    }

    updateColor(e) {
        const rect = this.canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        const dx = x - this.center;
        const dy = y - this.center;
        const distance = Math.sqrt(dx * dx + dy * dy);

        if (distance > this.radius) return;

        const angle = Math.atan2(dy, dx) * (180 / Math.PI);
        const hue = angle < 0 ? angle + 360 : angle;
        const sat = (distance / this.radius) * 100;
        const val = 100;

        const rgb = ColorUtils.hsvToRgb(hue, sat, val);
        this.currentColor = rgb;

        if (this.onColorChange) {
            this.onColorChange(rgb);
        }
    }

    setColor(r, g, b) {
        this.currentColor = { r, g, b };
    }
}

// ===== Controllers =====

// Device Controller
class DeviceController {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;
        this.scanBtn = $('#scan-btn');
        this.deviceList = $('#device-list');
        this.selectedDeviceEl = $('#selected-device');
        this.scanStatus = $('#scan-status');

        this.attachEvents();
    }

    attachEvents() {
        this.scanBtn.addEventListener('click', () => this.scanDevices());

        this.ws.on('scan_result', (message) => {
            this.displayDevices(message.devices);
        });
    }

    async scanDevices() {
        this.scanBtn.disabled = true;
        this.scanStatus.classList.remove('hidden');
        this.scanStatus.textContent = 'Scanning for devices...';

        try {
            const response = await fetch(`${API_URL}/scan`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ timeout: '15s' })
            });

            const devices = await response.json();
            this.displayDevices(devices);
            this.scanStatus.textContent = `Found ${devices.length} device(s)`;

            setTimeout(() => {
                this.scanStatus.classList.add('hidden');
            }, 3000);
        } catch (error) {
            console.error('Scan failed:', error);
            this.scanStatus.textContent = 'Scan failed';
        } finally {
            this.scanBtn.disabled = false;
        }
    }

    displayDevices(devices) {
        this.state.setState({ devices });

        if (devices.length === 0) {
            this.deviceList.innerHTML = '<p class="message info">No devices found. Try scanning again.</p>';
            return;
        }

        this.deviceList.innerHTML = '';
        devices.forEach(device => {
            const deviceEl = document.createElement('div');
            deviceEl.className = 'device-item';
            deviceEl.innerHTML = `
                <div class="device-name">${device.name}</div>
                <div class="device-details">
                    <span>Address: ${device.address}</span>
                    <span>Signal: ${device.rssi} dBm</span>
                </div>
            `;

            deviceEl.addEventListener('click', () => this.selectDevice(device));
            this.deviceList.appendChild(deviceEl);
        });
    }

    async selectDevice(device) {
        try {
            const response = await fetch(`${API_URL}/device/select`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ address: device.address })
            });

            const result = await response.json();
            if (result.success) {
                this.state.setState({ selectedDevice: result.device });
                this.updateSelectedDevice(result.device);

                // Show controls
                $('#no-device-message').classList.add('hidden');
                $('#controls').classList.remove('hidden');

                // Update device list selection
                $$('.device-item').forEach(el => el.classList.remove('selected'));
                event.currentTarget.classList.add('selected');
            }
        } catch (error) {
            console.error('Failed to select device:', error);
        }
    }

    updateSelectedDevice(device) {
        this.selectedDeviceEl.innerHTML = `
            <div><strong>Selected Device:</strong> ${device.name}</div>
            <div><small>Address: ${device.address}</small></div>
        `;
    }
}

// Power Controller
class PowerController {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;
        this.powerBtn = $('#power-btn');
        this.powerText = $('.power-text');

        this.attachEvents();
    }

    attachEvents() {
        this.powerBtn.addEventListener('click', () => this.togglePower());

        this.state.subscribe((state) => {
            this.updateUI(state.deviceState.power_on);
        });
    }

    togglePower() {
        const currentState = this.state.getState().deviceState.power_on;
        const newState = !currentState;

        // Optimistic update
        this.updateUI(newState);

        this.ws.sendCommand('power', { on: newState });
    }

    updateUI(powerOn) {
        if (powerOn) {
            this.powerBtn.classList.remove('off');
            this.powerBtn.classList.add('on');
            this.powerText.textContent = 'Power On';
        } else {
            this.powerBtn.classList.remove('on');
            this.powerBtn.classList.add('off');
            this.powerText.textContent = 'Power Off';
        }
    }
}

// Brightness Controller
class BrightnessController {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;
        this.slider = $('#brightness-slider');
        this.valueDisplay = $('#brightness-value');

        this.attachEvents();
    }

    attachEvents() {
        const debouncedSend = debounce((value) => {
            this.ws.sendCommand('brightness', { level: parseInt(value) });
        }, 150);

        this.slider.addEventListener('input', (e) => {
            const value = e.target.value;
            this.updateDisplay(value);
            debouncedSend(value);
        });

        this.state.subscribe((state) => {
            if (state.deviceState.brightness !== undefined) {
                this.slider.value = state.deviceState.brightness;
                this.updateDisplay(state.deviceState.brightness);
            }
        });
    }

    updateDisplay(value) {
        const percent = Math.round((value / 255) * 100);
        this.valueDisplay.textContent = `${percent}%`;
    }
}

// Color Controller
class ColorController {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;

        this.redSlider = $('#red-slider');
        this.greenSlider = $('#green-slider');
        this.blueSlider = $('#blue-slider');

        this.redValue = $('#red-value');
        this.greenValue = $('#green-value');
        this.blueValue = $('#blue-value');

        this.colorPreview = $('#color-preview');

        this.picker = new ColorPicker($('#color-picker'), (rgb) => {
            this.updateFromPicker(rgb);
        });

        this.attachEvents();
    }

    attachEvents() {
        const debouncedSend = debounce((r, g, b) => {
            this.ws.sendCommand('color', { r, g, b });
        }, 100);

        const updateColor = () => {
            const r = parseInt(this.redSlider.value);
            const g = parseInt(this.greenSlider.value);
            const b = parseInt(this.blueSlider.value);

            this.updateDisplay(r, g, b);
            debouncedSend(r, g, b);
        };

        this.redSlider.addEventListener('input', updateColor);
        this.greenSlider.addEventListener('input', updateColor);
        this.blueSlider.addEventListener('input', updateColor);

        this.state.subscribe((state) => {
            if (state.deviceState.rgb) {
                const { r, g, b } = state.deviceState.rgb;
                this.redSlider.value = r;
                this.greenSlider.value = g;
                this.blueSlider.value = b;
                this.updateDisplay(r, g, b);
            }
        });
    }

    updateFromPicker(rgb) {
        this.redSlider.value = rgb.r;
        this.greenSlider.value = rgb.g;
        this.blueSlider.value = rgb.b;

        this.updateDisplay(rgb.r, rgb.g, rgb.b);

        this.ws.sendCommand('color', rgb);
    }

    updateDisplay(r, g, b) {
        this.redValue.textContent = r;
        this.greenValue.textContent = g;
        this.blueValue.textContent = b;

        this.colorPreview.style.background = `rgb(${r}, ${g}, ${b})`;
        this.picker.setColor(r, g, b);
    }
}

// White Balance Controller
class WhiteBalanceController {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;

        this.warmSlider = $('#warm-slider');
        this.coldSlider = $('#cold-slider');

        this.warmValue = $('#warm-value');
        this.coldValue = $('#cold-value');

        this.attachEvents();
    }

    attachEvents() {
        const debouncedSend = debounce((warm, cold) => {
            this.ws.sendCommand('white_balance', { warm, cold });
        }, 150);

        const updateWhiteBalance = () => {
            const warm = parseInt(this.warmSlider.value);
            const cold = parseInt(this.coldSlider.value);

            this.updateDisplay(warm, cold);
            debouncedSend(warm, cold);
        };

        this.warmSlider.addEventListener('input', updateWhiteBalance);
        this.coldSlider.addEventListener('input', updateWhiteBalance);

        this.state.subscribe((state) => {
            if (state.deviceState.white_balance) {
                const { warm, cold } = state.deviceState.white_balance;
                this.warmSlider.value = warm;
                this.coldSlider.value = cold;
                this.updateDisplay(warm, cold);
            }
        });
    }

    updateDisplay(warm, cold) {
        this.warmValue.textContent = warm;
        this.coldValue.textContent = cold;
    }
}

// Effects Controller
class EffectsController {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;

        this.gallery = $('#effects-gallery');
        this.speedSlider = $('#effect-speed-slider');
        this.speedValue = $('#effect-speed-value');

        this.currentEffect = null;
        this.currentSpeed = 128;

        this.initGallery();
        this.attachEvents();
    }

    initGallery() {
        // Create effect cards for common effects (0-20)
        const effectNames = [
            'Seven Color Jump', 'Red Gradual', 'Green Gradual', 'Blue Gradual',
            'Yellow Gradual', 'Cyan Gradual', 'Purple Gradual', 'White Gradual',
            'Red-Green Jump', 'Red-Blue Jump', 'Green-Blue Jump',
            'Seven Color Strobe', 'Red Strobe', 'Green Strobe', 'Blue Strobe',
            'Yellow Strobe', 'Cyan Strobe', 'Purple Strobe', 'White Strobe',
            'Rainbow Fade', 'Random Colors'
        ];

        effectNames.forEach((name, index) => {
            const card = document.createElement('div');
            card.className = 'effect-card';
            card.dataset.effect = index;
            card.innerHTML = `
                <div class="effect-icon">✨</div>
                <div class="effect-name">${name}</div>
            `;

            card.addEventListener('click', () => this.selectEffect(index));
            this.gallery.appendChild(card);
        });
    }

    attachEvents() {
        const debouncedSend = debounce((effect, speed) => {
            this.ws.sendCommand('effect', { effect, speed });
        }, 200);

        this.speedSlider.addEventListener('input', (e) => {
            const speed = parseInt(e.target.value);
            this.currentSpeed = speed;
            this.speedValue.textContent = speed;

            if (this.currentEffect !== null) {
                debouncedSend(this.currentEffect, speed);
            }
        });

        this.state.subscribe((state) => {
            if (state.deviceState.effect !== null && state.deviceState.effect !== undefined) {
                this.currentEffect = state.deviceState.effect;
                this.updateGallerySelection(state.deviceState.effect);
            }
            if (state.deviceState.effect_speed !== null && state.deviceState.effect_speed !== undefined) {
                this.currentSpeed = state.deviceState.effect_speed;
                this.speedSlider.value = this.currentSpeed;
                this.speedValue.textContent = this.currentSpeed;
            }
        });
    }

    selectEffect(effectIndex) {
        this.currentEffect = effectIndex;
        this.updateGallerySelection(effectIndex);
        this.ws.sendCommand('effect', {
            effect: effectIndex,
            speed: this.currentSpeed
        });
    }

    updateGallerySelection(effectIndex) {
        $$('.effect-card').forEach(card => {
            card.classList.toggle('active', parseInt(card.dataset.effect) === effectIndex);
        });
    }
}

// Mode Tabs Controller
class ModeTabsController {
    constructor() {
        this.tabs = $$('.mode-tab');
        this.contents = $$('.mode-content');

        this.attachEvents();
    }

    attachEvents() {
        this.tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const mode = tab.dataset.mode;
                this.switchMode(mode);
            });
        });
    }

    switchMode(mode) {
        // Update tabs
        this.tabs.forEach(tab => {
            tab.classList.toggle('active', tab.dataset.mode === mode);
        });

        // Update content
        this.contents.forEach(content => {
            content.classList.toggle('active', content.id === `${mode}-mode`);
        });
    }
}

// Connection Status Controller
class ConnectionStatusController {
    constructor(wsClient) {
        this.ws = wsClient;
        this.statusEl = $('#connection-status');

        this.attachEvents();
    }

    attachEvents() {
        this.ws.on('connected', () => {
            this.statusEl.classList.remove('disconnected');
            this.statusEl.classList.add('connected');
            this.statusEl.querySelector('.status-text').textContent = 'Connected';
        });

        this.ws.on('disconnected', () => {
            this.statusEl.classList.remove('connected');
            this.statusEl.classList.add('disconnected');
            this.statusEl.querySelector('.status-text').textContent = 'Disconnected';
        });
    }
}

// State Update Handler
class StateUpdateHandler {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;

        this.attachEvents();
    }

    attachEvents() {
        this.ws.on('state_update', (message) => {
            console.log('State update received:', message);

            if (message.device) {
                this.state.setState({
                    selectedDevice: message.device
                });

                if (message.device.state) {
                    this.state.updateDeviceState(message.device.state);
                }
            }
        });

        this.ws.on('error', (message) => {
            console.error('WebSocket error:', message);
            this.showError(message.message || 'An error occurred');
        });
    }

    showError(message) {
        const errorDisplay = $('#error-display');
        errorDisplay.textContent = message;
        errorDisplay.classList.remove('hidden');

        setTimeout(() => {
            errorDisplay.classList.add('hidden');
        }, 5000);
    }
}

// ===== Application Initialization =====
class App {
    constructor() {
        this.wsClient = new WebSocketClient(WS_URL);
        this.stateManager = new StateManager();

        // Initialize controllers
        this.deviceController = new DeviceController(this.wsClient, this.stateManager);
        this.powerController = new PowerController(this.wsClient, this.stateManager);
        this.brightnessController = new BrightnessController(this.wsClient, this.stateManager);
        this.colorController = new ColorController(this.wsClient, this.stateManager);
        this.whiteBalanceController = new WhiteBalanceController(this.wsClient, this.stateManager);
        this.effectsController = new EffectsController(this.wsClient, this.stateManager);
        this.modeTabsController = new ModeTabsController();
        this.connectionStatusController = new ConnectionStatusController(this.wsClient);
        this.stateUpdateHandler = new StateUpdateHandler(this.wsClient, this.stateManager);

        console.log('LampControl Web UI initialized');
    }
}

// Start the application when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.app = new App();
});
