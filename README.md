# URL Shortener

## Overview

The URL Shortener is a simple and efficient application that allows users to shorten long URLs, making them easier to share and manage. It provides both a client interface for URL shortening and a server component that handles requests, generates short URLs, and stores mappings between short and original URLs.

## Features

- Shorten long URLs and generate unique short links.
- Retrieve original URLs from short links.
- Handle invalid URL submissions and provide appropriate error messages.
- Lightweight and easy to deploy.

## Directory Structure

- **cmd/shortener**: Contains the main server code for handling URL shortening requests. This directory will be compiled into the binary application.
- **cmd/client**: Contains the client code for interacting with the server and shortening URLs. This directory will also be compiled into a binary application.
- **config**: Contains configuration files and settings for the application, allowing for easy customization of server parameters and behavior.

## Getting Started

### Prerequisites

- Go (version 1.18 or higher)
- A working database (if using persistent storage)

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-name>
2. Install the required dependencies:
go mod tidy

3. Configure your settings in the config directory as needed.

4. Run the server:
go run cmd/shortener/main.go

5. Use the client to shorten URLs by running:
go run cmd/client/main.go

### Usage

To shorten a URL, use the client application to send a request with the original URL. The server will respond with the shortened link.

### Testing

To run tests, use the following command:
go test ./...

### License

This project is licensed under the MIT License - see the LICENSE file for details.

### Contributing

Contributions are welcome! Please open an issue or submit a pull request to contribute to the project.

### Contact

For inquiries, please reach out to [hairutdinl@protonmail.com].