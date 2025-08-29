# Go AppService Example

A comprehensive example demonstrating how to implement a **Daemon/Process Supervisor** in Go that can be installed and managed as an OS service on both Linux and Windows.

## 🎯 Overview

This project showcases a robust implementation of a process supervisor that can:
- **Start, monitor, and restart** child processes
- **Install and manage** itself as a system service
- **Handle graceful shutdowns** with proper signal management
- **Cross-platform support** for Linux and Windows
- **Comprehensive error handling** and logging

## 🏗️ Architecture

The application is structured with a clean separation of concerns:

```
├── cmd/           # CLI command implementations
│   ├── daemon.go  # Daemon service management
│   ├── run.go     # Direct execution with signal handling
│   ├── service.go # OS service management
│   └── root.go    # Root command setup
├── pkg/           # Core packages
│   └── daemon/    # Process supervisor implementation
└── main.go        # Application entry point
```

## 🚀 Features

### Process Supervision
- **Automatic restart** on process failure
- **Graceful shutdown** with configurable timeouts
- **Signal handling** (SIGTERM, SIGINT)
- **Environment variable** management
- **Output redirection** to custom writers

### Service Management
- **Cross-platform** service installation
- **Systemd support** on Linux
- **Windows Service** support
- **Service lifecycle** management (start/stop/restart/install/uninstall)

### CLI Interface
- **Intuitive commands** using Cobra
- **Flag-based configuration**
- **Helpful error messages**
- **Exit mode simulation** for testing

## 📦 Installation

### Prerequisites
- Go 1.25.0 or later
- Root/Administrator privileges for service installation

### Build
```bash
go build -o svcapp
```

## 🎮 Usage

### Direct Execution
Run the application directly with various exit modes:

```bash
# Run with random exit mode (for testing)
./svcapp run

# Run with specific exit mode
./svcapp run --exit-with err --timeout 10s

# Available exit modes: nil, rand, err, panic, fatal
```

### Service Management
Install and manage as a system service:

```bash
# Install the service (requires root/admin)
sudo ./svcapp service install

# Start the service
sudo ./svcapp service start

# Check service status
sudo systemctl status svcapp  # Linux
sc query svcapp               # Windows

# Stop the service
sudo ./svcapp service stop

# Uninstall the service
sudo ./svcapp service uninstall
```

### Daemon Mode
Run as a daemon process supervisor:

```bash
# Run in daemon mode
sudo ./svcapp daemon

# Run with additional arguments
sudo ./svcapp daemon --exit-with err
```

## 🔧 Configuration

### Service Configuration

The application automatically detects the operating system and applies appropriate configurations:

**Linux (systemd):**
```go
{
    Name: "svcapp",
    DisplayName: "SvcApp",
    Description: "A simple example of a Go application that can be installed as a service",
    WorkingDirectory: "~/.",
    Arguments: []string{"daemon"},
    Option: kardianos.KeyValue{
        "LogOutput": false,
        "PIDFile": "/var/run/svcapp.pid",
        "Restart": "on-success",
        "SuccessExitStatus": "0 2 SIGKILL",
        "LimitNOFILE": -1,
    },
    Dependencies: []string{
        "After=network-online.target",
        "Wants=network-online.target",
    },
}
```

**Windows:**
```go
{
    Name: "svcapp",
    DisplayName: "SvcApp", 
    Description: "A simple example of a Go application that can be installed as a service",
    WorkingDirectory: "~/.",
    Arguments: []string{"daemon"},
    Option: kardianos.KeyValue{
        "StartType": "automatic",
        "OnFailure": "restart",
        "OnFailureDelayDuration": "10s",
    },
}
```

### Daemon Configuration

```go
daemon.NewDaemon(&daemon.DaemonConfig{
    Executable: "",           // Auto-detected if empty
    Args: []string{"run"},    // Arguments to pass to child process
    EnvVars: []string{},      // Additional environment variables
    OutWriter: os.Stdout,     // Stdout writer
    ErrWriter: os.Stderr,     // Stderr writer
    ExitTimeout: 5 * time.Second, // Graceful shutdown timeout
})
```

## 🧪 Testing

The application includes built-in testing capabilities:

```bash
# Test different exit scenarios
./svcapp run --exit-with err --timeout 5s
./svcapp run --exit-with panic --timeout 5s
./svcapp run --exit-with fatal --timeout 5s

# Test service installation and management
sudo ./svcapp service install
sudo ./svcapp service start
sudo ./svcapp service stop
sudo ./svcapp service uninstall
```

## 🔍 Key Implementation Details

### Process Supervision
The `Daemon` type implements the `kardianos.Interface` and provides:

1. **Start()**: Launches the child process and begins supervision
2. **Stop()**: Sends SIGTERM and waits for graceful shutdown
3. **Timeout handling**: Force kills if graceful shutdown fails
4. **Exit handling**: Manages service lifecycle based on child process exit

### Signal Handling
- **SIGTERM**: Graceful shutdown request
- **SIGINT**: Interactive termination (Ctrl+C)
- **Context cancellation**: Supports Go context-based cancellation

### Error Handling
- **Comprehensive error wrapping** with `fmt.Errorf`
- **User-friendly error messages**
- **Proper exit codes** for different failure scenarios

## 🛠️ Dependencies

- **[github.com/spf13/cobra](https://github.com/spf13/cobra)**: CLI framework
- **[github.com/lucasdecamargo/kardianos](https://github.com/lucasdecamargo/kardianos)**: Service management (fork of kardianos/service)
- **Go standard library**: `os/exec`, `sync`, `syscall`, etc.

## 📝 Logging

The application uses structured logging with `log/slog`:

```json
{"time":"2024-01-15T10:30:00.000Z","level":"INFO","msg":"Running...","timeLeft":"25s"}
{"time":"2024-01-15T10:30:01.000Z","level":"INFO","msg":"Running...","timeLeft":"24s"}
{"time":"2024-01-15T10:30:25.000Z","level":"INFO","msg":"Timed out."}
{"time":"2024-01-15T10:30:25.000Z","level":"INFO","msg":"Exiting...","mode":"err"}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Based on the excellent [kardianos/service](https://github.com/kardianos/service) package
- Uses [Cobra](https://github.com/spf13/cobra) for CLI functionality
- Inspired by real-world process supervision needs
