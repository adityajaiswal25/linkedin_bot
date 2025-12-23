# Architecture Documentation

## Project Overview

This LinkedIn automation tool is built with Go using the Rod library for browser automation. It implements sophisticated anti-bot detection techniques to simulate human-like behavior.

## Package Architecture

### 1. Configuration (`pkg/config`)

**Purpose**: Centralized configuration management with YAML file support and environment variable overrides.

**Key Features**:
- YAML-based configuration
- Environment variable overrides
- Configuration validation
- Sensible defaults

**Main Types**:
- `Config`: Root configuration structure
- `BrowserConfig`: Browser automation settings
- `StealthConfig`: All anti-bot detection settings
- `LinkedInConfig`: LinkedIn-specific settings

### 2. Logger (`pkg/logger`)

**Purpose**: Structured logging with multiple output formats and log levels.

**Key Features**:
- JSON and text output formats
- Multiple log levels (debug, info, warn, error)
- Contextual field support
- File or stdout output

**Main Types**:
- `Logger`: Main logger instance
- `LogEntry`: Structured log entry
- `LogLevel`: Log level enumeration

### 3. Database (`pkg/database`)

**Purpose**: SQLite-based persistence for profiles, connections, messages, and statistics.

**Key Features**:
- SQLite database with modern driver
- Profile tracking
- Connection request history
- Message history
- Daily statistics tracking

**Main Types**:
- `DB`: Database connection wrapper
- `Profile`: LinkedIn profile information
- `ConnectionRequest`: Connection request record
- `Message`: Message record
- `DailyStats`: Daily activity statistics

### 4. Stealth (`pkg/stealth`)

**Purpose**: Implements all 8 anti-bot detection techniques.

**Key Features**:
- Human-like mouse movement with Bézier curves
- Randomized timing patterns
- Browser fingerprint masking
- Random scrolling behavior
- Realistic typing simulation
- Mouse hovering and movement
- Activity scheduling
- Rate limiting and throttling

**Main Types**:
- `Stealth`: Main stealth instance
- `Point`: 2D coordinate for mouse movement

**Key Methods**:
- `HumanMouseMove()`: Bézier curve mouse movement
- `RandomDelay()`: Randomized think time
- `MaskFingerprint()`: Browser fingerprint masking
- `RandomScroll()`: Variable speed scrolling
- `HumanType()`: Realistic typing with typos
- `RandomHover()`: Random hover events

### 5. Authentication (`pkg/auth`)

**Purpose**: Handles LinkedIn login and session management.

**Key Features**:
- Credential-based login
- Security checkpoint detection (2FA, captcha)
- Session persistence
- Login verification

**Main Types**:
- `Auth`: Authentication handler

**Key Methods**:
- `Login()`: Perform login flow
- `checkSecurityCheckpoints()`: Detect 2FA/captcha
- `verifyLogin()`: Verify successful login

### 6. Search (`pkg/search`)

**Purpose**: LinkedIn profile search and parsing.

**Key Features**:
- Multi-criteria search (job title, company, location, keywords)
- Pagination handling
- Duplicate detection
- Profile information extraction

**Main Types**:
- `Search`: Search handler
- `SearchParams`: Search criteria
- `Profile`: Found profile information

**Key Methods**:
- `SearchProfiles()`: Execute search and collect profiles
- `parseProfiles()`: Extract profiles from page
- `goToNextPage()`: Navigate to next results page

### 7. Connection (`pkg/connection`)

**Purpose**: Sends LinkedIn connection requests.

**Key Features**:
- Personalized connection notes
- Daily limit enforcement
- Connection modal handling
- Bulk connection requests

**Main Types**:
- `Connection`: Connection request handler

**Key Methods**:
- `SendConnectionRequest()`: Send single connection request
- `SendBulkConnectionRequests()`: Send multiple requests
- `findConnectButton()`: Locate connect button
- `handleConnectionModal()`: Handle connection modal

### 8. Messaging (`pkg/messaging`)

**Purpose**: Sends LinkedIn messages.

**Key Features**:
- Follow-up message automation
- Message templates with variables
- Message history tracking
- Bulk messaging

**Main Types**:
- `Messaging`: Message handler

**Key Methods**:
- `SendMessage()`: Send single message
- `SendFollowUpMessages()`: Auto-send to accepted connections
- `SendBulkMessages()`: Send multiple messages
- `findMessageButton()`: Locate message button

## Data Flow

### Search Flow
1. User provides search parameters
2. Search builds LinkedIn search URL
3. Navigate to search page
4. Parse profiles from results
5. Save profiles to database
6. Handle pagination
7. Return collected profiles

### Connection Flow
1. Check daily limit in database
2. Check if connection already sent
3. Navigate to profile page
4. Find and click connect button
5. Handle connection modal
6. Add personalized note
7. Send connection request
8. Save to database
9. Apply cooldown

### Messaging Flow
1. Check for newly accepted connections
2. Navigate to profile
3. Find message button
4. Open message interface
5. Type message with human-like typing
6. Send message
7. Save to database
8. Apply cooldown

## Anti-Bot Detection Implementation

### Technique 1: Human-like Mouse Movement
- **Implementation**: Cubic Bézier curves with randomized control points
- **Features**: Variable speed, overshoot, micro-corrections
- **Location**: `pkg/stealth/stealth.go::HumanMouseMove()`

### Technique 2: Randomized Timing Patterns
- **Implementation**: Configurable min/max delays with random selection
- **Features**: Think time, scroll delays, action intervals
- **Location**: `pkg/stealth/stealth.go::RandomDelay()`, `RandomScrollDelay()`

### Technique 3: Browser Fingerprint Masking
- **Implementation**: User agent rotation, viewport variation, webdriver flag removal
- **Features**: Multiple user agents, random viewports
- **Location**: `pkg/stealth/stealth.go::MaskFingerprint()`

### Technique 4: Random Scrolling Behavior
- **Implementation**: Variable speed scrolling with multiple steps
- **Features**: Scroll-back movements, natural deceleration
- **Location**: `pkg/stealth/stealth.go::RandomScroll()`, `SmoothScroll()`

### Technique 5: Realistic Typing Simulation
- **Implementation**: Variable keystroke delays with typo simulation
- **Features**: Typo probability, backspace corrections
- **Location**: `pkg/stealth/stealth.go::HumanType()`

### Technique 6: Mouse Hovering & Movement
- **Implementation**: Random hover events before interactions
- **Features**: Configurable hover probability and duration
- **Location**: `pkg/stealth/stealth.go::RandomHover()`

### Technique 7: Activity Scheduling
- **Implementation**: Business hours checking and break simulation
- **Features**: Configurable work hours, random breaks
- **Location**: `pkg/stealth/stealth.go::ShouldOperate()`, `RandomBreak()`

### Technique 8: Rate Limiting & Throttling
- **Implementation**: Cooldown periods between actions
- **Features**: Connection and message cooldowns with variance
- **Location**: `pkg/stealth/stealth.go::ConnectionCooldown()`, `MessageCooldown()`

## Error Handling Strategy

1. **Graceful Degradation**: Operations continue even if some steps fail
2. **Detailed Logging**: All errors logged with context
3. **Retry Logic**: Where applicable (future enhancement)
4. **User Feedback**: Clear error messages in logs

## State Management

- **Database**: SQLite for persistent state
- **Profiles**: Tracked to avoid duplicates
- **Connections**: Status tracked (pending, accepted, rejected)
- **Messages**: History maintained
- **Daily Stats**: Limits enforced per day

## Security Considerations

- Credentials stored in environment variables
- No hardcoded secrets
- Session cookies managed by browser
- Database stored locally

## Extension Points

The architecture supports easy extension:

1. **New Stealth Techniques**: Add methods to `pkg/stealth`
2. **Additional Search Criteria**: Extend `SearchParams` in `pkg/search`
3. **New Message Templates**: Add to config file
4. **Custom Logging**: Implement logger interface
5. **Alternative Storage**: Replace database implementation

## Testing Considerations

- Unit tests for each package
- Integration tests for workflows
- Mock browser for testing
- Test database for isolation

## Performance Considerations

- Database queries optimized with indexes
- Efficient profile parsing
- Minimal page navigation
- Cooldown periods prevent rate limiting

