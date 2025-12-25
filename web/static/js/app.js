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

function throttle(func, wait) {
    let lastCall = 0;
    let timeout = null;

    return function executedFunction(...args) {
        const now = Date.now();
        const timeSinceLastCall = now - lastCall;

        if (timeSinceLastCall >= wait) {
            // Enough time has passed, call immediately
            lastCall = now;
            func(...args);
        } else {
            // Too soon, schedule for later
            clearTimeout(timeout);
            timeout = setTimeout(() => {
                lastCall = Date.now();
                func(...args);
            }, wait - timeSinceLastCall);
        }
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
        this.dropdown = $('#device-dropdown');
        this.devices = [];
        this.isScanning = false;

        this.attachEvents();
        this.autoScanAndConnect();
    }

    attachEvents() {
        // When dropdown is focused (opened), trigger scan if needed
        this.dropdown.addEventListener('focus', () => {
            // If dropdown is empty or has placeholder, trigger scan
            if (!this.isScanning && (this.dropdown.value === '' || this.dropdown.options.length <= 1)) {
                this.scanDevices();
            }
        });

        // When user selects a device
        this.dropdown.addEventListener('change', () => {
            const address = this.dropdown.value;
            if (address) {
                const device = this.devices.find(d => d.address === address);
                if (device) {
                    this.selectDevice(device);
                }
            }
        });
    }

    async autoScanAndConnect() {
        // Load cached devices from localStorage for instant display
        const cachedDevices = localStorage.getItem('cachedDevices');
        const lastDevice = localStorage.getItem('lastDeviceAddress');

        if (cachedDevices) {
            try {
                this.devices = JSON.parse(cachedDevices);
                this.populateDropdown(this.devices);

                // If we have cached devices and a last device, select it immediately
                if (lastDevice && this.devices.some(d => d.address === lastDevice)) {
                    this.dropdown.value = lastDevice;
                    const device = this.devices.find(d => d.address === lastDevice);
                    await this.selectDevice(device);
                    // Successfully connected to cached device, skip scan
                    return;
                } else if (this.devices.length > 0) {
                    this.dropdown.value = this.devices[0].address;
                    await this.selectDevice(this.devices[0]);
                    // Successfully connected to cached device, skip scan
                    return;
                }
            } catch (e) {
                console.error('Failed to load cached devices:', e);
            }
        }

        // Only scan if we don't have cached devices or failed to connect
        await this.scanDevices();

        // Select device from fresh scan
        if (lastDevice && this.devices.some(d => d.address === lastDevice)) {
            this.dropdown.value = lastDevice;
            const device = this.devices.find(d => d.address === lastDevice);
            await this.selectDevice(device);
        } else if (this.devices.length > 0) {
            this.dropdown.value = this.devices[0].address;
            await this.selectDevice(this.devices[0]);
        }
    }

    async scanDevices() {
        if (this.isScanning) return;

        this.isScanning = true;
        this.dropdown.disabled = true;

        // Show scanning state
        const originalHTML = this.dropdown.innerHTML;
        this.dropdown.innerHTML = '<option value="">Scanning...</option>';

        try {
            const response = await fetch(`${API_URL}/scan`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ timeout: '5s' })
            });

            const devices = await response.json();
            this.devices = devices;

            // Cache devices to localStorage
            localStorage.setItem('cachedDevices', JSON.stringify(devices));

            this.populateDropdown(devices);
        } catch (error) {
            console.error('Scan failed:', error);
            this.dropdown.innerHTML = '<option value="">Scan failed - Click to retry</option>';
        } finally {
            this.isScanning = false;
            this.dropdown.disabled = false;
        }
    }

    populateDropdown(devices) {
        this.dropdown.innerHTML = '';

        if (devices.length === 0) {
            this.dropdown.innerHTML = '<option value="">No devices found - Click to scan</option>';
            return;
        }

        // Add placeholder option
        const placeholder = document.createElement('option');
        placeholder.value = '';
        placeholder.textContent = 'Select a device...';
        this.dropdown.appendChild(placeholder);

        // Add devices
        devices.forEach(device => {
            const option = document.createElement('option');
            option.value = device.address;
            option.textContent = `${device.name} (${device.rssi} dBm)`;
            this.dropdown.appendChild(option);
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
                // Save to localStorage
                localStorage.setItem('lastDeviceAddress', device.address);

                // Update state with device info
                this.state.setState({ selectedDevice: result.device });

                // Update UI with current device state
                if (result.device.state) {
                    this.updateUIFromDeviceState(result.device.state);
                }

                // Show controls
                $('#no-device-message').classList.add('hidden');
                $('#controls').classList.remove('hidden');
            }
        } catch (error) {
            console.error('Failed to select device:', error);
        }
    }

    updateUIFromDeviceState(state) {
        // Update power button
        if (state.power_on !== undefined) {
            const powerBtn = $('#power-btn');
            const powerText = $('.power-text');
            if (state.power_on) {
                powerBtn.classList.remove('off');
                powerBtn.classList.add('on');
                powerText.textContent = 'Power On';
            } else {
                powerBtn.classList.remove('on');
                powerBtn.classList.add('off');
                powerText.textContent = 'Power Off';
            }
        }

        // Update brightness
        if (state.brightness !== undefined) {
            $('#brightness-slider').value = state.brightness;
            const percent = Math.round((state.brightness / 255) * 100);
            $('#brightness-value').textContent = `${percent}%`;
        }

        // Update RGB color
        if (state.rgb) {
            $('#red-slider').value = state.rgb.r;
            $('#green-slider').value = state.rgb.g;
            $('#blue-slider').value = state.rgb.b;
            $('#red-value').textContent = state.rgb.r;
            $('#green-value').textContent = state.rgb.g;
            $('#blue-value').textContent = state.rgb.b;

            const colorPreview = $('#color-preview');
            colorPreview.style.background = `rgb(${state.rgb.r}, ${state.rgb.g}, ${state.rgb.b})`;
        }

        // Update white balance
        if (state.white_balance) {
            $('#warm-slider').value = state.white_balance.warm;
            $('#cold-slider').value = state.white_balance.cold;
            $('#warm-value').textContent = state.white_balance.warm;
            $('#cold-value').textContent = state.white_balance.cold;
        }

        // Update effect
        if (state.effect !== null && state.effect !== undefined) {
            // Highlight the active effect card
            $$('.effect-card').forEach(card => {
                card.classList.toggle('active', parseInt(card.dataset.effect) === state.effect);
            });
        }

        if (state.effect_speed !== null && state.effect_speed !== undefined) {
            $('#effect-speed-slider').value = state.effect_speed;
            $('#effect-speed-value').textContent = state.effect_speed;
        }
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
        const throttledSend = throttle((value) => {
            this.ws.sendCommand('brightness', { level: parseInt(value) });
        }, 50);

        this.slider.addEventListener('input', (e) => {
            const value = e.target.value;
            this.updateDisplay(value);
            throttledSend(value);
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
        const throttledSend = throttle((r, g, b) => {
            this.ws.sendCommand('color', { r, g, b });
        }, 50);

        this.throttledSend = throttledSend; // Store for use by picker

        const updateColor = () => {
            const r = parseInt(this.redSlider.value);
            const g = parseInt(this.greenSlider.value);
            const b = parseInt(this.blueSlider.value);

            this.updateDisplay(r, g, b);
            throttledSend(r, g, b);
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

        this.throttledSend(rgb.r, rgb.g, rgb.b);
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
        const throttledSend = throttle((warm, cold) => {
            this.ws.sendCommand('white_balance', { warm, cold });
        }, 50);

        const updateWhiteBalance = () => {
            const warm = parseInt(this.warmSlider.value);
            const cold = parseInt(this.coldSlider.value);

            this.updateDisplay(warm, cold);
            throttledSend(warm, cold);
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

    async initGallery() {
        // Clear gallery
        this.gallery.innerHTML = '';

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

        // Load custom effects from API
        await this.loadCustomEffects();
    }

    async loadCustomEffects() {
        try {
            const response = await fetch(`${API_URL}/effects`);
            const customEffects = await response.json();

            customEffects.forEach(effect => {
                const card = document.createElement('div');
                card.className = 'effect-card custom';
                card.dataset.effectId = effect.id;
                card.dataset.custom = 'true';
                card.innerHTML = `
                    <div class="effect-icon">⭐</div>
                    <div class="effect-name">${effect.name}</div>
                    <div class="delete-effect" data-id="${effect.id}">×</div>
                `;

                // Click card to select effect (handled later)
                card.addEventListener('click', (e) => {
                    if (!e.target.classList.contains('delete-effect')) {
                        console.log('Custom effect selected:', effect);
                        // Custom effects would need special handling
                    }
                });

                // Delete button
                card.querySelector('.delete-effect').addEventListener('click', async (e) => {
                    e.stopPropagation();
                    if (confirm(`Delete effect "${effect.name}"?`)) {
                        await this.deleteCustomEffect(effect.id);
                    }
                });

                this.gallery.appendChild(card);
            });
        } catch (error) {
            console.error('Failed to load custom effects:', error);
        }
    }

    async deleteCustomEffect(id) {
        try {
            await fetch(`${API_URL}/effects/${id}`, { method: 'DELETE' });
            await this.loadCustomEffects();
        } catch (error) {
            console.error('Failed to delete effect:', error);
        }
    }

    attachEvents() {
        const throttledSend = throttle((effect, speed) => {
            this.ws.sendCommand('effect', { effect, speed });
        }, 50);

        this.throttledSend = throttledSend; // Store for use by selectEffect

        this.speedSlider.addEventListener('input', (e) => {
            const speed = parseInt(e.target.value);
            this.currentSpeed = speed;
            this.speedValue.textContent = speed;

            if (this.currentEffect !== null) {
                throttledSend(this.currentEffect, speed);
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
        this.throttledSend(effectIndex, this.currentSpeed);
    }

    updateGallerySelection(effectIndex) {
        $$('.effect-card').forEach(card => {
            card.classList.toggle('active', parseInt(card.dataset.effect) === effectIndex);
        });
    }
}

// Main Tabs Controller
class MainTabsController {
    constructor() {
        this.tabs = $$('.main-tab');
        this.contents = $$('.main-tab-content');

        this.attachEvents();
    }

    attachEvents() {
        this.tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                const tabName = tab.dataset.tab;
                this.switchTab(tabName);
            });
        });
    }

    switchTab(tabName) {
        // Update tabs
        this.tabs.forEach(tab => {
            tab.classList.toggle('active', tab.dataset.tab === tabName);
        });

        // Update content
        this.contents.forEach(content => {
            content.classList.toggle('active', content.id === `${tabName}-tab`);
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
            this.statusEl.classList.add('connected');
        });

        this.ws.on('disconnected', () => {
            this.statusEl.classList.remove('connected');
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

// Custom Effect Modal Controller
class CustomEffectModalController {
    constructor(effectsController) {
        this.effectsController = effectsController;
        this.modal = $('#effect-modal');
        this.form = $('#effect-form');
        this.colorSwatches = $('#color-swatches');
        this.colors = [];

        this.attachEvents();
    }

    attachEvents() {
        // Open modal
        $('#create-effect-btn').addEventListener('click', () => this.openModal());

        // Close modal
        $('#cancel-effect-btn').addEventListener('click', () => this.closeModal());

        // Add color
        $('#add-color-btn').addEventListener('click', () => this.addColorFromPicker());

        // Update speed display
        $('#effect-speed-input').addEventListener('input', (e) => {
            $('#effect-speed-input-value').textContent = e.target.value;
        });

        // Submit form
        this.form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.createEffect();
        });

        // Click outside to close
        this.modal.addEventListener('click', (e) => {
            if (e.target === this.modal) {
                this.closeModal();
            }
        });
    }

    openModal() {
        this.colors = [];
        this.renderColorSwatches();
        this.form.reset();
        this.modal.classList.remove('hidden');
    }

    closeModal() {
        this.modal.classList.add('hidden');
    }

    addColorFromPicker() {
        // Get current RGB from color controller
        const r = parseInt($('#red-slider').value);
        const g = parseInt($('#green-slider').value);
        const b = parseInt($('#blue-slider').value);

        this.colors.push({ r, g, b });
        this.renderColorSwatches();
    }

    removeColor(index) {
        this.colors.splice(index, 1);
        this.renderColorSwatches();
    }

    renderColorSwatches() {
        this.colorSwatches.innerHTML = '';

        this.colors.forEach((color, index) => {
            const swatch = document.createElement('div');
            swatch.className = 'color-swatch';
            swatch.style.background = `rgb(${color.r}, ${color.g}, ${color.b})`;
            swatch.innerHTML = `<div class="remove-color">×</div>`;

            swatch.querySelector('.remove-color').addEventListener('click', (e) => {
                e.stopPropagation();
                this.removeColor(index);
            });

            this.colorSwatches.appendChild(swatch);
        });
    }

    async createEffect() {
        const name = $('#effect-name').value;
        const pattern = $('#effect-pattern').value;
        const speed = parseInt($('#effect-speed-input').value);

        if (!name || this.colors.length === 0) {
            alert('Please provide a name and at least one color');
            return;
        }

        try {
            const response = await fetch(`${API_URL}/effects`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name,
                    colors: this.colors,
                    pattern,
                    speed
                })
            });

            if (response.ok) {
                this.closeModal();
                await this.effectsController.loadCustomEffects();
            } else {
                alert('Failed to create effect');
            }
        } catch (error) {
            console.error('Failed to create effect:', error);
            alert('Failed to create effect');
        }
    }
}

// Sound Controller
class SoundController {
    constructor(wsClient, stateManager) {
        this.ws = wsClient;
        this.state = stateManager;
        this.isActive = false;
        this.audioContext = null;
        this.analyser = null;
        this.mediaStream = null;

        this.toggleBtn = $('#sound-toggle-btn');
        this.btnText = $('#sound-btn-text');
        this.sensitivitySlider = $('#sound-sensitivity-slider');
        this.sensitivityValue = $('#sound-sensitivity-value');
        this.visualizer = $('#audio-visualizer');
        this.canvas = $('#audio-canvas');
        this.canvasCtx = this.canvas.getContext('2d');

        this.sensitivity = 50;

        this.attachEvents();
    }

    attachEvents() {
        this.toggleBtn.addEventListener('click', () => this.toggle());

        this.sensitivitySlider.addEventListener('input', (e) => {
            this.sensitivity = parseInt(e.target.value);
            this.sensitivityValue.textContent = `${this.sensitivity}%`;
        });
    }

    async toggle() {
        if (this.isActive) {
            this.stop();
        } else {
            await this.start();
        }
    }

    async start() {
        try {
            // Request microphone access
            this.mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true });

            // Create audio context
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
            this.analyser = this.audioContext.createAnalyser();
            this.analyser.fftSize = 256;

            const source = this.audioContext.createMediaStreamSource(this.mediaStream);
            source.connect(this.analyser);

            this.isActive = true;
            this.btnText.textContent = 'Stop Sound Reactive';
            this.toggleBtn.classList.add('active');
            this.visualizer.classList.remove('hidden');

            // Start animation loop
            this.animate();
        } catch (error) {
            console.error('Failed to access microphone:', error);
            alert('Failed to access microphone. Please grant permission.');
        }
    }

    stop() {
        this.isActive = false;
        this.btnText.textContent = 'Start Sound Reactive';
        this.toggleBtn.classList.remove('active');
        this.visualizer.classList.add('hidden');

        if (this.mediaStream) {
            this.mediaStream.getTracks().forEach(track => track.stop());
            this.mediaStream = null;
        }

        if (this.audioContext) {
            this.audioContext.close();
            this.audioContext = null;
        }
    }

    animate() {
        if (!this.isActive) return;

        requestAnimationFrame(() => this.animate());

        const bufferLength = this.analyser.frequencyBinCount;
        const dataArray = new Uint8Array(bufferLength);
        this.analyser.getByteFrequencyData(dataArray);

        // Calculate average volume
        const average = dataArray.reduce((a, b) => a + b) / bufferLength;
        const normalizedVolume = (average / 255) * (this.sensitivity / 50);

        // Map volume to color (bass = red, mid = green, high = blue)
        const bass = dataArray.slice(0, bufferLength / 3).reduce((a, b) => a + b) / (bufferLength / 3);
        const mid = dataArray.slice(bufferLength / 3, (bufferLength * 2) / 3).reduce((a, b) => a + b) / (bufferLength / 3);
        const high = dataArray.slice((bufferLength * 2) / 3).reduce((a, b) => a + b) / (bufferLength / 3);

        const r = Math.min(255, Math.floor((bass / 255) * 255 * (this.sensitivity / 50)));
        const g = Math.min(255, Math.floor((mid / 255) * 255 * (this.sensitivity / 50)));
        const b = Math.min(255, Math.floor((high / 255) * 255 * (this.sensitivity / 50)));

        // Send color to lamp (throttled by existing throttle)
        if (normalizedVolume > 0.1) {
            this.ws.sendCommand('color', { r, g, b });
        }

        // Draw visualizer
        this.drawVisualizer(dataArray, bufferLength);
    }

    drawVisualizer(dataArray, bufferLength) {
        const width = this.canvas.width;
        const height = this.canvas.height;

        this.canvasCtx.fillStyle = 'rgba(28, 28, 30, 0.5)';
        this.canvasCtx.fillRect(0, 0, width, height);

        const barWidth = (width / bufferLength) * 2.5;
        let barHeight;
        let x = 0;

        for (let i = 0; i < bufferLength; i++) {
            barHeight = (dataArray[i] / 255) * height;

            const hue = (i / bufferLength) * 360;
            this.canvasCtx.fillStyle = `hsl(${hue}, 70%, 50%)`;
            this.canvasCtx.fillRect(x, height - barHeight, barWidth, barHeight);

            x += barWidth + 1;
        }
    }
}

// Twitch Controller
class TwitchController {
    constructor(wsClient) {
        this.ws = wsClient;
        this.statusPollInterval = null;

        // Form elements
        this.enabledCheckbox = $('#twitch-enabled');
        this.channelInput = $('#twitch-channel');
        this.botUsernameInput = $('#twitch-bot-username');
        this.clientIdInput = $('#twitch-client-id');
        this.tokenInput = $('#twitch-token');
        this.effectDurationSlider = $('#effect-duration');
        this.effectDurationValue = $('#effect-duration-value');
        this.globalCooldownSlider = $('#global-cooldown');
        this.globalCooldownValue = $('#global-cooldown-value');
        this.userCooldownSlider = $('#user-cooldown');
        this.userCooldownValue = $('#user-cooldown-value');
        this.vipBypassCheckbox = $('#vip-bypass');
        this.subBypassCheckbox = $('#sub-bypass');
        this.modBypassCheckbox = $('#mod-bypass');
        this.saveBtn = $('#save-twitch-config');
        this.getOAuthBtn = $('#get-oauth-btn');

        // Status elements
        this.statusIndicator = $('#twitch-connection-status');
        this.statusText = $('#twitch-connection-text');

        // Active effect elements
        this.activeEffectDiv = $('#active-effect');
        this.effectUsername = $('#effect-username');
        this.effectCommand = $('#effect-command');
        this.effectRemaining = $('#effect-remaining');

        // Available commands
        this.availableColors = $('#available-colors');
        this.availableEffects = $('#available-effects');

        this.attachEvents();
        this.loadConfig();
        this.loadAvailableCommands();
        this.startStatusPolling();
    }

    attachEvents() {
        // Slider value updates
        this.effectDurationSlider.addEventListener('input', (e) => {
            this.effectDurationValue.textContent = `${e.target.value}s`;
        });

        this.globalCooldownSlider.addEventListener('input', (e) => {
            this.globalCooldownValue.textContent = `${e.target.value}s`;
        });

        this.userCooldownSlider.addEventListener('input', (e) => {
            this.userCooldownValue.textContent = `${e.target.value}s`;
        });

        // Save button
        this.saveBtn.addEventListener('click', () => this.saveConfig());

        // OAuth button
        this.getOAuthBtn.addEventListener('click', () => this.openOAuthURL());

        // WebSocket listeners
        this.ws.on('twitch_status', (message) => this.handleTwitchStatus(message));
        this.ws.on('twitch_command', (message) => this.handleTwitchCommand(message));
    }

    async loadConfig() {
        try {
            const response = await fetch(`${API_URL}/twitch/config`);
            const config = await response.json();

            this.enabledCheckbox.checked = config.enabled || false;
            this.channelInput.value = config.channel || '';
            this.botUsernameInput.value = config.bot_username || '';
            this.effectDurationSlider.value = config.effect_duration_sec || 30;
            this.effectDurationValue.textContent = `${config.effect_duration_sec || 30}s`;
            this.globalCooldownSlider.value = config.global_cooldown_sec || 5;
            this.globalCooldownValue.textContent = `${config.global_cooldown_sec || 5}s`;
            this.userCooldownSlider.value = config.user_cooldown_sec || 30;
            this.userCooldownValue.textContent = `${config.user_cooldown_sec || 30}s`;
            this.vipBypassCheckbox.checked = config.vip_bypass_cooldown !== false;
            this.subBypassCheckbox.checked = config.sub_bypass_cooldown !== false;
            this.modBypassCheckbox.checked = config.mod_bypass_cooldown !== false;
        } catch (error) {
            console.error('Failed to load Twitch config:', error);
        }
    }

    async saveConfig() {
        const config = {
            enabled: this.enabledCheckbox.checked,
            channel: this.channelInput.value,
            bot_username: this.botUsernameInput.value,
            effect_duration_sec: parseInt(this.effectDurationSlider.value),
            global_cooldown_sec: parseInt(this.globalCooldownSlider.value),
            user_cooldown_sec: parseInt(this.userCooldownSlider.value),
            vip_bypass_cooldown: this.vipBypassCheckbox.checked,
            sub_bypass_cooldown: this.subBypassCheckbox.checked,
            mod_bypass_cooldown: this.modBypassCheckbox.checked
        };

        // Only include token if it was entered
        if (this.tokenInput.value) {
            config.access_token = this.tokenInput.value;
        }

        try {
            const response = await fetch(`${API_URL}/twitch/config`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });

            if (response.ok) {
                this.showMessage('Twitch configuration saved successfully', 'success');
                // Clear token field after saving for security
                this.tokenInput.value = '';
                // Reload status to reflect changes
                await this.loadStatus();
            } else {
                this.showMessage('Failed to save Twitch configuration', 'error');
            }
        } catch (error) {
            console.error('Failed to save Twitch config:', error);
            this.showMessage('Failed to save Twitch configuration', 'error');
        }
    }

    async loadStatus() {
        try {
            const response = await fetch(`${API_URL}/twitch/status`);
            const status = await response.json();

            this.updateStatusUI(status.connected);
            this.updateActiveEffect(status.active_effect);
        } catch (error) {
            console.error('Failed to load Twitch status:', error);
        }
    }

    async loadAvailableCommands() {
        try {
            const response = await fetch(`${API_URL}/twitch/commands`);
            const commands = await response.json();

            this.availableColors.textContent = commands.colors.map(c => `!lamp ${c}`).join(', ');
            this.availableEffects.textContent = commands.effects.map(e => `!lamp ${e}`).join(', ');
        } catch (error) {
            console.error('Failed to load available commands:', error);
            this.availableColors.textContent = 'Failed to load';
            this.availableEffects.textContent = 'Failed to load';
        }
    }

    async openOAuthURL() {
        // Open window immediately to avoid popup blockers
        const authWindow = window.open('', '_blank', 'width=600,height=800,scrollbars=yes,resizable=yes');
        if (authWindow) {
            authWindow.document.write('<html><body style="background:#1a1a1a;color:#fff;font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100%;"><h3>Loading Twitch Login...</h3></body></html>');
            authWindow.document.close();
        }

        try {
            const response = await fetch(`${API_URL}/twitch/oauth`);
            const data = await response.json();

            const authUrl = data.url || data.oauth_url;
            if (authUrl) {

                if (authWindow) {
                    authWindow.location.href = authUrl;
                } else {
                    window.open(authUrl, '_blank', 'noopener,noreferrer');
                }
                this.showMessage('OAuth page opened. Copy the token from the URL after authorizing.', 'info');
            } else {
                console.error('OAuth response error:', data);
                let errorMsg = data.error || data.message || 'No URL returned from server';
                if (errorMsg.toLowerCase().includes('invalid client')) {
                    errorMsg += '. Please check your Twitch Client ID in the configuration.';
                }
                if (authWindow) {
                    authWindow.document.body.innerHTML = `<h3 style="color:#ff5555">Failed to get OAuth URL</h3><p>${errorMsg}</p>`;
                }
                this.showMessage(`Failed to get OAuth URL: ${errorMsg}`, 'error');
            }
        } catch (error) {
            if (authWindow) {
                authWindow.document.body.innerHTML = `<h3 style="color:#ff5555">Connection Error</h3><p>${error.message}</p>`;
            }
            console.error('Failed to get OAuth URL:', error);
            this.showMessage('Failed to get OAuth URL', 'error');
        }
    }

    updateStatusUI(connected) {
        if (connected) {
            this.statusIndicator.classList.remove('disconnected');
            this.statusIndicator.classList.add('connected');
            this.statusText.textContent = 'Connected';
        } else {
            this.statusIndicator.classList.remove('connected');
            this.statusIndicator.classList.add('disconnected');
            this.statusText.textContent = 'Disconnected';
        }
    }

    updateActiveEffect(activeEffect) {
        if (activeEffect && activeEffect.username) {
            this.activeEffectDiv.classList.remove('hidden');
            this.effectUsername.textContent = activeEffect.username;
            this.effectCommand.textContent = activeEffect.command;
            this.effectRemaining.textContent = activeEffect.remaining_time_sec || 0;
        } else {
            this.activeEffectDiv.classList.add('hidden');
        }
    }

    handleTwitchStatus(message) {
        if (message.status) {
            this.updateStatusUI(message.status.connected);
            this.updateActiveEffect(message.status.active_effect);
        }
    }

    handleTwitchCommand(message) {
        // Show notification when a viewer triggers a command
        const notification = `${message.username} triggered: !lamp ${message.command}`;
        this.showMessage(notification, 'info');

        // Reload status to update active effect
        this.loadStatus();
    }

    startStatusPolling() {
        // Poll status every 5 seconds to update remaining time
        this.statusPollInterval = setInterval(() => {
            this.loadStatus();
        }, 5000);
    }

    stopStatusPolling() {
        if (this.statusPollInterval) {
            clearInterval(this.statusPollInterval);
            this.statusPollInterval = null;
        }
    }

    showMessage(message, type) {
        const errorDisplay = $('#error-display');
        errorDisplay.textContent = message;
        errorDisplay.className = `message ${type}`;
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
        this.customEffectModalController = new CustomEffectModalController(this.effectsController);
        this.soundController = new SoundController(this.wsClient, this.stateManager);
        this.twitchController = new TwitchController(this.wsClient);
        this.mainTabsController = new MainTabsController();
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
