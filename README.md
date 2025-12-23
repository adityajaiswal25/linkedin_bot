# LinkedIn Automation Tool

Video link - https://drive.google.com/file/d/1stv2ksEP4I1iciLXZ_VKoUJXPzxyf3uw/view?usp=sharing


A sophisticated Go-based LinkedIn automation tool using the Rod library that showcases advanced browser automation capabilities, human-like behavior simulation, and sophisticated anti-bot detection techniques.

## Features

### Core Functionality

- **Authentication System**: Secure login with session persistence, 2FA/captcha detection, and graceful error handling
- **Search & Targeting**: Advanced profile search by job title, company, location, and keywords with pagination support
- **Connection Requests**: Automated connection requests with personalized notes and daily limit enforcement
- **Messaging System**: Automated follow-up messages to accepted connections with template support

### Anti-Bot Detection (8 Techniques)

#### Mandatory Techniques (3)

1. **Human-like Mouse Movement**: Bézier curve trajectories with variable speed, natural overshoot, and micro-corrections
2. **Randomized Timing Patterns**: Realistic delays between actions, variable think time, and scroll speed variance
3. **Browser Fingerprint Masking**: User agent randomization, viewport variation, and webdriver flag disabling

#### Additional Techniques (5)

4. **Random Scrolling Behavior**: Variable scroll speeds, natural acceleration/deceleration, occasional scroll-back movements
5. **Realistic Typing Simulation**: Variable keystroke intervals, occasional typos with corrections, human typing rhythm
6. **Mouse Hovering & Movement**: Random hover events, natural cursor wandering, realistic movement patterns
7. **Activity Scheduling**: Business hours operation, realistic break patterns, human work schedule simulation
8. **Rate Limiting & Throttling**: Connection request quotas, spaced messaging intervals, cooldown periods

## Installation

### Prerequisites

- Go 1.21 or higher
- Chrome/Chromium browser (automatically downloaded by Rod)

### Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd linkedin-automation
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your LinkedIn credentials
```

4. Update configuration file:
```bash
# Edit config/config.yaml with your preferences
```

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
LINKEDIN_EMAIL=your_email@example.com
LINKEDIN_PASSWORD=your_password_here
LINKEDIN_DAILY_LIMIT=50
LINKEDIN_HEADLESS=false
```

### Configuration File

The `config/config.yaml` file contains comprehensive settings for:

- Browser settings (headless mode, viewport, timeout)
- Search parameters (max results, pagination delay)
- Connection request limits and delays
- Messaging templates and follow-up delays
- All stealth/anti-bot detection settings
- Database path
- Logging configuration

## Usage

### Basic Usage

```bash
# Search for profiles
go run main.go -mode=search

# Send connection requests
go run main.go -mode=connect

# Send follow-up messages
go run main.go -mode=message

# Run all operations
go run main.go -mode=all
```

### Command Line Options

- `-config`: Path to configuration file (default: `config/config.yaml`)
- `-mode`: Operation mode - `search`, `connect`, `message`, or `all` (default: `search`)

### Building

```bash
# Build executable
go build -o linkedin-automation main.go

# Run executable
./linkedin-automation -mode=all
```

## Architecture

### Package Structure

```
linkedin-automation/
├── pkg/
│   ├── auth/          # Authentication and session management
│   ├── config/         # Configuration management
│   ├── connection/     # Connection request handling
│   ├── database/       # SQLite database operations
│   ├── logger/         # Structured logging
│   ├── messaging/      # Message sending and templates
│   ├── search/         # Profile search and parsing
│   └── stealth/        # Anti-bot detection techniques
├── config/
│   └── config.yaml     # Configuration file
├── data/               # Database storage (created automatically)
├── main.go             # Main application entry point
└── go.mod              # Go module definition
```

### Key Components

#### Authentication (`pkg/auth`)
- Handles LinkedIn login flow
- Detects security checkpoints (2FA, captcha)
- Manages browser session
- Applies stealth techniques on initialization

#### Search (`pkg/search`)
- Builds LinkedIn search URLs
- Parses profile information from search results
- Handles pagination
- Detects and filters duplicates
- Saves profiles to database

#### Connection (`pkg/connection`)
- Sends connection requests with personalized notes
- Enforces daily limits
- Tracks sent requests in database
- Handles connection modal interactions

#### Messaging (`pkg/messaging`)
- Detects newly accepted connections
- Sends follow-up messages automatically
- Supports message templates with variable substitution
- Tracks message history

#### Stealth (`pkg/stealth`)
- Implements all 8 anti-bot detection techniques
- Provides human-like interaction methods
- Manages timing and rate limiting
- Handles browser fingerprint masking

#### Database (`pkg/database`)
- SQLite-based persistence
- Tracks profiles, connection requests, messages
- Maintains daily statistics
- Enables resumption after interruptions

## Anti-Bot Detection Details

### 1. Human-like Mouse Movement

Uses cubic Bézier curves to create natural mouse trajectories:
- Control points are randomized for each movement
- Variable speed: slower at start/end, faster in middle
- Occasional overshoot with correction
- Micro-corrections during movement

### 2. Randomized Timing Patterns

- Think time: 1-5 seconds (configurable)
- Scroll delays with variance
- Action intervals mimic human cognitive processing

### 3. Browser Fingerprint Masking

- Randomizes user agent from pool of realistic browsers
- Varies viewport dimensions
- Disables `navigator.webdriver` flag
- Uses Rod's stealth plugin

### 4. Random Scrolling Behavior

- Variable scroll speed (70-130% of base)
- Multiple steps for smooth scrolling
- Occasional scroll-back movements (20% probability)
- Natural acceleration/deceleration curves

### 5. Realistic Typing Simulation

- Variable keystroke delays (50-200ms)
- 5% typo probability with automatic correction
- Backspace patterns
- Human typing rhythm variations

### 6. Mouse Hovering & Movement

- 40% probability of hovering over elements
- Random hover duration (500-2000ms)
- Natural cursor wandering
- Hover before click interactions

### 7. Activity Scheduling

- Business hours only mode (9 AM - 5 PM)
- Random break patterns (10% probability)
- Configurable work schedule
- Respects human work patterns

### 8. Rate Limiting & Throttling

- Connection cooldown: 1 minute (with variance)
- Message cooldown: 30 seconds (with variance)
- Daily connection limits enforced
- Prevents rapid-fire actions

## Database Schema

The tool uses SQLite to persist:

- **profiles**: LinkedIn profile information
- **connection_requests**: Sent connection requests with status
- **messages**: Sent messages history
- **daily_stats**: Daily activity tracking

## Logging

Structured logging with support for:
- Multiple log levels (debug, info, warn, error)
- JSON or text format
- File or stdout output
- Contextual information in log entries

## Error Handling

- Comprehensive error detection and logging
- Graceful degradation when operations fail
- Retry mechanisms with exponential backoff (where applicable)
- Detailed error messages with context

## Security Considerations

⚠️ **Important**: This tool is for educational and proof-of-concept purposes only.

- Never share your LinkedIn credentials
- Use environment variables for sensitive data
- Be aware of LinkedIn's Terms of Service
- Respect rate limits to avoid account restrictions
- Use responsibly and ethically

## Limitations

- LinkedIn's UI changes may require selector updates
- 2FA and captcha require manual intervention
- Some features may be rate-limited by LinkedIn
- Browser automation can be detected by advanced systems

## Troubleshooting

### Login Issues

- Verify credentials in `.env` file
- Check for 2FA/captcha requirements
- Ensure browser can launch (check Chrome installation)

### Search Not Working

- LinkedIn may have changed selectors
- Check network connectivity
- Verify search parameters in code

### Connection Requests Failing

- Check daily limit in database
- Verify profile URLs are valid
- Ensure you're not already connected

## Contributing

This is a proof-of-concept project. Contributions should focus on:
- Code quality improvements
- Additional stealth techniques
- Better error handling
- Documentation enhancements

## License

This project is for educational purposes only. Use at your own risk.

## Disclaimer

This tool is provided as-is for educational and research purposes. The authors are not responsible for any misuse or violations of LinkedIn's Terms of Service. Users are responsible for ensuring their use complies with all applicable terms and laws.

