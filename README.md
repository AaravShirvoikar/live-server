# live-server

A simple live server CLI in Go that serves files from a specified directory and automatically reloads the browser when files are modified.

## Features

- Serves static files (HTML, CSS, JS) from a specified directory.
- Automatically reloads the browser when any file in the directory is changed.
- Uses WebSocket to update files in real-time and display them in your terminal.

## Getting Started

### Prerequisites

- Go installed on your machine. You can download it from [golang.org](https://golang.org/dl/).
- Ensure that `$GOPATH/bin` is included in your system's `PATH`.

### Installation

If you have Go installed, you can use:

```bash
go install github.com/AaravShirvoikar/live-server@latest
```

### Usage

1. Run the server with the following command:

   ```bash
   live-server -dir <path-to-directory> -port <port-number>
   ```

   - Replace `<path-to-directory>` with the directory you want to serve files from (default is the current directory).
   - Replace `<port-number>` with the port you want the server to run on (default is `8080`).

   Example:

   ```bash
   live-server -dir . -port 8080
   ```

2. Open your web browser and go to `http://localhost:<port-number>` to access the server.

### Notes

- Ensure that your `index.html` file is present in the specified directory, as it will be served by default when accessing the root url (`/`).
