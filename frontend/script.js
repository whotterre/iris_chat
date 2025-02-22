document.addEventListener("DOMContentLoaded", () => {
    // Connect to the WebSocket server
    const sockConn = new WebSocket("ws://localhost:4000/ws");

    // DOM elements
    const messageInput = document.querySelector("#message_form"); // Input field for the message
    const submitBtn = document.querySelector("#submit_btn");   // Form element
    const leaveBtn = document.querySelector("#leave_btn");
    const messageBoard = document.querySelector("#messageboard")
    // Function to send a message
    function sendMessage(event) {
        event.preventDefault();
        const audioSound = new Audio("./assets/chat_sound.mp3")
        audioSound.play()
        const message = messageInput.value.trim(); // Get the message from the input field
        if (message !== "") {
            // Send the message as a JSON object

            sockConn.send(message);
            const messageDisplay = document.createElement("div");
            messageDisplay.textContent = `[You]: ${message}`;
            messageBoard.appendChild(messageDisplay);
            // Clear the input field
            messageInput.value = "";
        }
    }


    function toggleButton(event) {
        event.preventDefault();

        submitBtn.classList.toggle("hidden");
        leaveBtn.classList.toggle("hidden");
    }


    function leaveRoom(event) {
        event.preventDefault()
        sockConn.send("User left the room. Repairing you....")
        sockConn.send(JSON.stringify({ type: "leave" }));
        sockConn.close();
    }
    // Event listener for the submit button
    submitBtn.addEventListener("click", sendMessage);
    submitBtn.addEventListener("dblclick", toggleButton)
    leaveBtn.addEventListener("click", leaveRoom)
    leaveBtn.addEventListener("dblclick", toggleButton)
    // WebSocket event listeners
    sockConn.onopen = () => {
        console.log("WebSocket connection established.");
    };

    sockConn.onmessage = (event) => {
        try {
            const messageObj = JSON.parse(event.data);
            if (Notification.permission === "granted") {
                const notif = new Notification("Iris Chat", {
                    body: `Received message: ${messageObj.message || "New Message"}`
                });
            } else if (Notification.permission !== "denied") {
                Notification.requestPermission().then(permission => {
                    if (permission === "granted") {
                        const notif = new Notification("Iris Chat", {
                            body: `Received message: ${messageObj.message || "New Message"}`
                        });
                    }
                });
            }

            const messageDisplay = document.createElement("div");
            messageDisplay.textContent = `[${messageObj.sender || "Unknown"}]: ${messageObj.message || "No message content"}`;
            messageBoard.appendChild(messageDisplay);
        } catch (e) {
            console.error("Failed to parse message:", event.data);
            const errorDisplay = document.createElement("div");
            errorDisplay.textContent = "Error: Failed to parse incoming message.";
            messageBoard.appendChild(errorDisplay);
        }
    };

    sockConn.onclose = (event) => {
        console.log(event)
        alert("WebSocket closed:", event.type, event.reason);
    };


    sockConn.onerror = (error) => {
        alert("WebSocket error:", error);
    };
});
