## RenderStream Restarter

A utility tool designed to monitor and automatically restart RenderStream layers in a Disguise environment.

### Features
- Monitors RenderStream layer status
- Automatic layer restart functionality
- REST API integration with Disguise
- Configurable timeout settings

### Usage
```cmd
renderstreamstarter.exe -server "ip of director" -layer "layer name"
```

### Parameters
- `server`: IP address of the Disguise director
- `layer`: Name of the RenderStream layer to monitor
- `timeout`: (Optional) Timeout in seconds between checks (default: 5)

### Technical Details
- Built with Go 1.20
- Uses REST API for communication
- Supports automatic GitHub releases
- Windows executable