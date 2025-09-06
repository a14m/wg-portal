// Application state and configuration
const App = {
    apiBase: "/api",
    elements: {
        connectionList: document.getElementById('connections__container'),
        statusArea: document.getElementById('status__container'),
        messageArea: document.getElementById('notifications__container')
    }
};

// Utility functions
const Utils = {
    // Show message to user
    renderError(element, message) {
        Utils.renderMessage(element, message, "error");
    },
    renderWarning(element, message) {
        Utils.renderMessage(element, message, "warning");
    },
    renderSuccess(element, message) {
        Utils.renderMessage(element, message, "success");
    },
    renderMessage(element, message, type = '') {
        element.innerHTML = `
            <div class="message ${type}">${message}</div>
        `;
    },
    // Make API calls with proper error handling
    async apiCall(endpoint, options = {}) {
        try {
            const response = await fetch(`${App.apiBase}${endpoint}`, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || `HTTP ${response.status}`);
            }

            if (!data.success) {
                throw new Error(data.error || 'API request failed');
            }

            return data.data;
        } catch (error) {
            console.error('API call failed:', error);
            throw error;
        }
    }
};

// Connection management
const ConnectionManager = {
    // Load and display all connections
    async loadConnections() {
        try {
            const connections = await Utils.apiCall('/connections');
            this.renderConnections(connections);
        } catch (error) {
            Utils.renderError(App.elements.connectionList,
                `Failed to load connections: ${error.message}`);
        }
    },

    // Render connections to the DOM
    renderConnections(connections) {
        if (!connections || connections.length === 0) {
            Utils.renderWarning(App.elements.connectionList,
                "No WireGuard connections found in /etc/wireguard/");
            return;
        }

        const html = connections.map(conn => `
            <div class="connection ${conn.active ? 'active' : ''}"
                    data-connection="${conn.name}"
                    onclick="ConnectionManager.toggleConnection('${conn.name}')">
                <div class="connection__name ${conn.active ? 'active' : ''}">${conn.name}</div>
            </div>
        `).join('');
        App.elements.connectionList.innerHTML = html;
    },

    // Toggle connection state
    async toggleConnection(name) {
        const connection = document.querySelector(`[data-connection="${name}"]`);

        try {
            connection.disabled = true;
            connection.textContent = 'Processing...';
            connection.className = "connection loading"

            await Utils.apiCall('/connections/toggle', {
                method: 'POST',
                body: JSON.stringify({ name })
            });
            await this.loadConnections(); // Refresh the list
            await StatusManager.loadStatus(); // Refresh the status
            Utils.renderSuccess(App.elements.messageArea, `No Errors Found.`);
        } catch (error) {
            Utils.renderError(App.elements.messageArea, `Failed to toggle ${name}: ${error.message}`);
            connection.disabled = false;
            connection.textContent = name;
            connection.className = "connection"
        }
    }
};

// Status management
const StatusManager = {
    // Load and display WireGuard status
    async loadStatus() {
        try {
            const statusData = await Utils.apiCall('/status');
            const statusText = statusData.status;
            if (statusText) {
                Utils.renderSuccess(App.elements.statusArea, statusText);
            } else {
                Utils.renderWarning(App.elements.statusArea, "No active connections.");
            }
        } catch (error) {
            Utils.renderError(App.elements.statusArea, error.message);
        }
    },
};

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    StatusManager.loadStatus();
    ConnectionManager.loadConnections();
});
