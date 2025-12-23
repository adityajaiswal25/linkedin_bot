# Quick Start Guide

## Prerequisites

1. **Install Go**: Download and install Go 1.21+ from [golang.org](https://golang.org/dl/)
2. **Verify Installation**:
   ```bash
   go version
   ```

## Setup Steps

### 1. Navigate to Project Directory
```bash
cd linkedin-automation
```

### 2. Install Dependencies
```bash
go mod download
```

### 3. Configure Environment Variables

Create a `.env` file in the project root:

```env
LINKEDIN_EMAIL=your_email@example.com
LINKEDIN_PASSWORD=your_password
```

**Important**: Never commit the `.env` file to version control!

### 4. Review Configuration

Edit `config/config.yaml` to customize:
- Daily connection limits
- Search parameters
- Stealth settings
- Logging preferences

### 5. Run the Tool

#### Search for Profiles
```bash
go run main.go -mode=search
```

#### Send Connection Requests
```bash
go run main.go -mode=connect
```

#### Send Follow-Up Messages
```bash
go run main.go -mode=message
```

#### Run All Operations
```bash
go run main.go -mode=all
```

## First Run Checklist

- [ ] Go installed and verified
- [ ] Dependencies downloaded (`go mod download`)
- [ ] `.env` file created with credentials
- [ ] `config/config.yaml` reviewed
- [ ] Browser can launch (Chrome/Chromium required)
- [ ] Test run with `-mode=search` successful

## Common Issues

### "go: command not found"
- Install Go from golang.org
- Add Go to your PATH

### "Failed to launch browser"
- Ensure Chrome/Chromium is installed
- Rod will download Chromium automatically on first run

### "Login failed"
- Verify credentials in `.env`
- Check for 2FA/captcha (requires manual intervention)
- Ensure LinkedIn account is not locked

### "Database error"
- Ensure write permissions in project directory
- Check that `data/` directory can be created

## Next Steps

1. **Customize Search**: Edit search parameters in `main.go` `runSearch()` function
2. **Adjust Limits**: Modify daily limits in `config/config.yaml`
3. **Tune Stealth**: Adjust anti-bot detection settings in config
4. **Monitor Logs**: Check logs for operation status

## Building Executable

```bash
# Build for current platform
go build -o linkedin-automation main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o linkedin-automation.exe main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o linkedin-automation main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o linkedin-automation main.go
```

## Usage Examples

### Example 1: Search for Software Engineers
Edit `main.go` in `runSearch()`:
```go
params := search.SearchParams{
    JobTitle: "Software Engineer",
    Location: "San Francisco",
    Keywords: "Python Go",
}
```

### Example 2: Custom Connection Note
Edit `config/config.yaml`:
```yaml
connections:
  default_note: "Hi {name}, I'd like to connect with you about {title}!"
```

### Example 3: Adjust Stealth Settings
Edit `config/config.yaml`:
```yaml
stealth:
  mouse_movement:
    overshoot_probability: 0.5  # More overshoot
  timing:
    min_think_time: 2000  # Slower actions
    max_think_time: 8000
```

## Safety Tips

1. **Start Small**: Begin with low daily limits
2. **Monitor Activity**: Check logs regularly
3. **Respect Limits**: Don't exceed LinkedIn's rate limits
4. **Test First**: Run in non-headless mode to observe behavior
5. **Backup Data**: Database is in `data/` directory

## Getting Help

- Check `README.md` for detailed documentation
- Review `ARCHITECTURE.md` for code structure
- Check logs for error details
- Verify configuration settings

## Troubleshooting

### Tool Runs But No Results
- Check LinkedIn search URL format
- Verify selectors haven't changed (LinkedIn UI updates)
- Check network connectivity

### Connection Requests Not Sending
- Verify daily limit not reached
- Check if already connected to profile
- Ensure connect button is found

### Messages Not Sending
- Verify connection was accepted
- Check message button availability
- Review message template syntax

## Advanced Usage

### Custom Search Implementation
Modify `pkg/search/search.go` to add new search criteria or parsing logic.

### Custom Stealth Techniques
Add new methods to `pkg/stealth/stealth.go` for additional anti-bot techniques.

### Database Queries
Use `pkg/database/database.go` methods to query stored data:
```go
stats, _ := db.GetDailyStats(time.Now())
fmt.Printf("Connections sent today: %d\n", stats.ConnectionsSent)
```

