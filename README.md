# IrisChat

IrisChat is a real-time, anonymous chat application that pairs users randomly for text-based conversations. I wanted to create something Omegle-like to learn more about WebSockets and build a project that isn't heavily dependent on databases. For now, I'm focusing on the backend. Built with Go (Gorilla WebSocket) and React, it provides a seamless and secure chatting experience.

## Features

- **Real-Time Messaging**: Instant text-based communication between paired users.
- **Anonymous Chat**: No personal information is required to start chatting.
- **Random Pairing**: Users are paired randomly in a lobby system.
- **WebSocket Support**: Efficient and low-latency communication using WebSockets.
- **Simple UI**: Clean and intuitive user interface built with React and Tailwind CSS.

## Tech Stack

- **Backend**: Go (Gorilla WebSocket)
- **Frontend**: React + Tailwind CSS
- **Real-Time Communication**: WebSockets
- **Deployment**: Docker, Kubernetes (optional)

## Getting Started

### Prerequisites

- Go (v1.20 or higher)
- Node.js (v16 or higher) for the frontend
- Docker (optional, for containerization)

### Installation

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/whotterre/iris_chat.git
   cd iris_chat
   ```

2. **Set Up the Backend**:
   - Navigate to the backend directory:
     ```bash
     cd backend
     ```
   - Install dependencies:
     ```bash
     go mod download
     ```
   - Start the server:
     ```bash
     go run cmd/server/main.go
     ```

3. **Set Up the Frontend**:
   - Navigate to the frontend directory:
     ```bash
     cd frontend
     ```
   - Install dependencies:
     ```bash
     npm install
     ```
   - Start the development server:
     ```bash
     npm start
     ```

4. **Run with Docker** (optional):
   - Build and run the application using Docker Compose:
     ```bash
     docker-compose up --build
     ```

### Usage

1. Open the application in your browser: `http://localhost:3000`.
2. Click "Start Chat" to enter the lobby.
3. Wait to be paired with another user.
4. Start chatting in real time!

## Project Structure

```
irischat/
├── backend/               # Backend (Go)
│   ├── cmd/               # Entry point for the server
│   ├── internal/          # Application logic
│   │   ├── handlers/      # HTTP and WebSocket handlers
│   │   ├── models/        # Data models (e.g., User, Lobby)
│   │   └── lobby/         # Lobby management logic
│   └── go.mod             # Go module file
├── frontend/              # Frontend (React + Tailwind CSS)
│   ├── public/            # Static assets
│   ├── src/               # React components and logic
│   └── package.json       # Node.js dependencies
├── docker-compose.yml     # Docker Compose configuration
└── README.md              # Project documentation
```

## API Endpoints

- **WebSocket**: `ws://localhost:8080/ws`
  - Handles real-time communication between users.

- **Health Check**: `GET /health`
  - Returns the server status.

- **Statistics**: `GET /stats`
  - Returns the number of active users and rooms.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch: `git checkout -b feature/your-feature`.
3. Commit your changes: `git commit -m 'Add some feature'`.
4. Push to the branch: `git push origin feature/your-feature`.
5. Submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gorilla WebSocket](https://github.com/gorilla/websocket) for WebSocket support.
- [React](https://reactjs.org/) and [Tailwind CSS](https://tailwindcss.com/) for the frontend.

---

Made with ❤️ by [Iwegbu Jeddy](https://github.com/your-username)
```

---

You can now copy this entire block and paste it into your `README.md` file. Let me know if you need further adjustments! 🚀
