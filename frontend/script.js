document.addEventListener("DOMContentLoaded", () => {
    // Connect to WebSocket
    const sockConn = new WebSocket("https://iris-chat-900j.onrender.com/ws");

    // DOM elements
    const messageInput = document.querySelector("#message_form"); // Make sure this ID matches your HTML
    const submitBtn = document.querySelector("#submit_btn");
    const leaveBtn = document.querySelector("#leave_btn");
    const messageBoard = document.querySelector("#messageboard");

    // Function to send a message
    function sendMessage(event) {
        event.preventDefault();

        const audioSound = new Audio("./assets/chat_sound.mp3");
        audioSound.play();

        const message = messageInput.value.trim();
        if (message !== "") {
            const messageData = { message: message };
            sockConn.send(JSON.stringify(messageData));

            // Generate timestamp
            const timestamp = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: true });

            // Display message in chat
            const messageDisplay = document.createElement("div");
            // messageDisplay.classList.add("message", "sent");
            messageDisplay.innerHTML =

                `
<div class="chat-message sent">
    <span class="message-text"><strong>You:</strong> ${message}</span>
    <span class="timestamp">${timestamp}</span>
</div>
`;


            messageBoard.appendChild(messageDisplay);
            messageBoard.scrollTop = messageBoard.scrollHeight;

            // Clear input field
            messageInput.value = "";
        }
    }

    function toggleButton(event) {
        event.preventDefault();
        submitBtn.classList.toggle("hidden");
        leaveBtn.classList.toggle("hidden");
    }

    function leaveRoom(event) {
        event.preventDefault();
        sockConn.send(JSON.stringify({ type: "leave", message: "User left the room." }));
        sockConn.close();
    }

    // Event listeners
    submitBtn.addEventListener("click", sendMessage);
    submitBtn.addEventListener("dblclick", toggleButton);
    leaveBtn.addEventListener("click", leaveRoom);
    leaveBtn.addEventListener("dblclick", toggleButton);

    // WebSocket event listeners
    sockConn.onopen = () => {
        console.log("WebSocket connection established.");
    };

    sockConn.onmessage = (event) => {
        try {
            const messageObj = JSON.parse(event.data);

            const receivedMessage = messageObj.message?.trim() || "No message content";

            if (Notification.permission === "granted") {
                new Notification("Iris Chat", { body: `Received message: ${receivedMessage}` });
            } else if (Notification.permission !== "denied") {
                Notification.requestPermission().then(permission => {
                    if (permission === "granted") {
                        new Notification("Iris Chat", { body: `Received message: ${receivedMessage}` });
                    }
                });
            }

            // Generate timestamp
            const timestamp = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: true });

            // Display message in chat
            const messageDisplay = document.createElement("div");
            // messageDisplay.classList.add("message", "received");
            messageDisplay.innerHTML = `
            <div class="chat-message recieved">
            <span class="message-text"><strong>Stranger:</strong> ${receivedMessage}</span>
            <span class="timestamp">${timestamp}</span>
            </div> `


            messageBoard.appendChild(messageDisplay);
            messageBoard.scrollTop = messageBoard.scrollHeight;
        } catch (e) {
            console.error("Failed to parse message:", event.data);
        }
    };


    sockConn.onclose = (event) => {
        console.log("WebSocket closed:", event);
        alert("WebSocket connection closed.");
    };

    sockConn.onerror = (error) => {
        console.error("WebSocket error:", error);
        alert("WebSocket encountered an error.");
    };
});
