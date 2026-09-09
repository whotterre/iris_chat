document.addEventListener("DOMContentLoaded", () => {
    function getOrCreateSessionID() {
        let id = localStorage.getItem("iris_chat_session_id");
        if (!id) {
            const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
            let res = "";
            for (let i = 0; i < 32; i++) {
                if (i > 0 && i % 8 === 0) res += "-";
                res += charset[Math.floor(Math.random() * charset.length)];
            }
            id = res;
            localStorage.setItem("iris_chat_session_id", id);
        }
        return id;
    }

    const sessionId = getOrCreateSessionID();

    const messageBoard = document.getElementById("messageboard");
    const messageInput = document.getElementById("message_input");
    const chatForm = document.getElementById("chat_form");
    const nextBtn = document.getElementById("next_btn");
    const statusDot = document.getElementById("status_dot");
    const statusText = document.getElementById("status_text");
    const statusBanner = document.getElementById("status_banner");
    const bannerText = document.getElementById("banner_text");
    const onlineCountDisplay = document.getElementById("connection_count");
    const soundToggleBtn = document.getElementById("sound_toggle_btn");
    const typingIndicator = document.getElementById("typing_indicator");

    let soundEnabled = true;
    let ws = null;
    let reconnectAttempts = 0;
    let reconnectTimer = null;
    let currentStatus = "connecting";
    let isTypingSent = false;
    let typingTimer = null;

    let chatAudio = null;
    try {
        chatAudio = new Audio("./assets/chat_sound.mp3");
    } catch (e) {}

    function playChatSound() {
        if (soundEnabled && chatAudio) {
            chatAudio.currentTime = 0;
            chatAudio.play().catch(() => {});
        }
    }

    soundToggleBtn.addEventListener("click", () => {
        soundEnabled = !soundEnabled;
        soundToggleBtn.textContent = soundEnabled ? "sound on" : "sound off";
    });

    messageInput.addEventListener("input", () => {
        if (!ws || ws.readyState !== WebSocket.OPEN || currentStatus !== "connected") return;

        if (!isTypingSent) {
            isTypingSent = true;
            ws.send(JSON.stringify({ type: "typing", is_typing: "true" }));
        }

        if (typingTimer) clearTimeout(typingTimer);
        typingTimer = setTimeout(() => {
            isTypingSent = false;
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: "typing", is_typing: "false" }));
            }
        }, 2000);
    });

    function stopTyping() {
        if (typingTimer) clearTimeout(typingTimer);
        if (isTypingSent && ws && ws.readyState === WebSocket.OPEN) {
            isTypingSent = false;
            ws.send(JSON.stringify({ type: "typing", is_typing: "false" }));
        }
    }

    function getWebSocketURL() {
        const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
        const host = window.location.host || "localhost:4000";
        return `${protocol}//${host}/ws?session_id=${encodeURIComponent(sessionId)}`;
    }

    function updateStatusUI(status, message) {
        currentStatus = status;

        statusDot.className = "dot";
        statusBanner.className = "status-banner";

        switch (status) {
            case "connected":
                statusDot.classList.add("dot-connected");
                statusText.textContent = "MATCHED";
                statusBanner.classList.add("banner-connected");
                bannerText.textContent = message || "Connected with a stranger. Say hello.";
                break;
            case "waiting":
                statusDot.classList.add("dot-waiting");
                statusText.textContent = "SEARCHING";
                statusBanner.classList.add("banner-waiting");
                bannerText.textContent = message || "Searching for an available stranger...";
                break;
            case "offline":
                statusDot.classList.add("dot-offline");
                statusText.textContent = "OFFLINE";
                statusBanner.classList.add("banner-offline");
                bannerText.textContent = message || "Connection lost. Reconnecting...";
                break;
            default:
                statusDot.classList.add("dot-connecting");
                statusText.textContent = "CONNECTING";
                statusBanner.classList.add("banner-waiting");
                bannerText.textContent = message || "Connecting to iris server...";
                break;
        }
    }

    function connectWebSocket() {
        if (ws) {
            try { ws.close(); } catch (e) {}
        }

        updateStatusUI("connecting", "Connecting to server...");

        const wsUrl = getWebSocketURL();

        try {
            ws = new WebSocket(wsUrl);
        } catch (err) {
            scheduleReconnect();
            return;
        }

        ws.onopen = () => {
            reconnectAttempts = 0;
            if (reconnectTimer) {
                clearTimeout(reconnectTimer);
                reconnectTimer = null;
            }
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                handleIncomingMessage(data);
            } catch (err) {}
        };

        ws.onclose = () => {
            updateStatusUI("offline", "Connection lost. Reconnecting...");
            scheduleReconnect();
        };

        ws.onerror = () => {
            ws.close();
        };
    }

    function scheduleReconnect() {
        if (reconnectTimer) return;

        reconnectAttempts++;
        const delay = Math.min(1000 * Math.pow(1.5, reconnectAttempts - 1), 10000);
        const seconds = Math.round(delay / 1000);

        updateStatusUI("offline", `Reconnecting in ${seconds}s (Attempt ${reconnectAttempts})...`);

        reconnectTimer = setTimeout(() => {
            reconnectTimer = null;
            connectWebSocket();
        }, delay);
    }

    function handleIncomingMessage(data) {
        if (data.type === "connections" && data.count !== undefined) {
            if (onlineCountDisplay) {
                onlineCountDisplay.textContent = data.count;
            }
            return;
        }

        if (data.type === "status") {
            if (data.status === "connected") {
                updateStatusUI("connected", "Connected with a stranger. Say hello.");
                appendSystemAura("Channel matched with anonymous stranger.");
            } else if (data.status === "waiting") {
                updateStatusUI("waiting", "Searching for an available stranger...");
            }
            return;
        }

        if (data.type === "partner_left") {
            updateStatusUI("waiting", "Stranger disconnected. Searching for a new match...");
            appendSystemAura(data.message || "Stranger has left the conversation.");
            if (typingIndicator) typingIndicator.classList.add("hidden");
            return;
        }

        if (data.type === "typing") {
            if (data.is_typing === "true" || data.is_typing === true) {
                if (typingIndicator) typingIndicator.classList.remove("hidden");
            } else {
                if (typingIndicator) typingIndicator.classList.add("hidden");
            }
            return;
        }

        if (data.type === "chat" || data.message) {
            if (typingIndicator) typingIndicator.classList.add("hidden");
            const time = data.timestamp || getCurrentTime();
            appendChatMessage("Stranger", data.message, false, time);
            playChatSound();
            return;
        }

        if (data.type === "error") {
            appendSystemAura(data.message);
        }
    }

    chatForm.addEventListener("submit", (e) => {
        e.preventDefault();

        const messageText = messageInput.value.trim();
        if (!messageText) return;

        if (!ws || ws.readyState !== WebSocket.OPEN) {
            appendSystemAura("Connection offline.");
            return;
        }

        stopTyping();

        ws.send(JSON.stringify({
            type: "chat",
            message: messageText
        }));

        appendChatMessage("You", messageText, true, getCurrentTime());
        playChatSound();

        messageInput.value = "";
    });

    nextBtn.addEventListener("click", (e) => {
        e.preventDefault();

        stopTyping();
        if (typingIndicator) typingIndicator.classList.add("hidden");

        if (!ws || ws.readyState !== WebSocket.OPEN) {
            connectWebSocket();
            return;
        }

        ws.send(JSON.stringify({ type: "find_next" }));
        updateStatusUI("waiting", "Searching for a new stranger...");
        appendSystemAura("Finding a new stranger...");
    });

    function appendChatMessage(senderLabel, text, isSent, time) {
        const msgDiv = document.createElement("div");
        msgDiv.className = `chat-message ${isSent ? "sent" : "received"}`;

        const sanitizedText = escapeHTML(text);

        msgDiv.innerHTML = `
            <div class="message-text">${sanitizedText}</div>
            <div class="meta-line">
                <span>${escapeHTML(senderLabel)}</span>
                <span>•</span>
                <span>${escapeHTML(time)}</span>
            </div>
        `;

        messageBoard.appendChild(msgDiv);
        scrollToBottom();

        // Time Dissipation Effect (Glow softens after 15s, dissipates after 60s)
        setTimeout(() => {
            msgDiv.classList.add("aged-medium");
        }, 15000);

        setTimeout(() => {
            msgDiv.classList.add("aged-old");
        }, 60000);
    }

    function appendSystemAura(text) {
        const entryDiv = document.createElement("div");
        entryDiv.className = "system-aura";
        entryDiv.innerHTML = `<span>${escapeHTML(text)}</span>`;

        messageBoard.appendChild(entryDiv);
        scrollToBottom();
    }

    function scrollToBottom() {
        messageBoard.scrollTop = messageBoard.scrollHeight;
    }

    function getCurrentTime() {
        return new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false });
    }

    function escapeHTML(str) {
        if (typeof str !== "string") return "";
        return str
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }

    connectWebSocket();
});
