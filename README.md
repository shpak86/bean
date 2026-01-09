# Bean — Behavioral Metrics Analysis System

Bean is a lightweight service for collecting, storing, and analyzing user behavioral metrics. It enables the detection and classification of automated behavior based on customizable rules.

## Overview

On the client side, a JavaScript script collects user behavioral metrics (traces) and regularly sends them to the server. On the server side, metrics are aggregated and a score is calculated to classify user actions. Classification is performed based on flexible trace processing rules. The server provides a REST API that allows retrieving the score by user identifier.

## Metrics

The script collects the following metrics:

- **timestamp** (string) — event timestamp in ISO 8601 format, indicating when the trace was recorded.
- **mouseMoves** (int) — total number of recorded mouse movements during the session.
- **clicks** (int) — total number of clicks (left, right, middle) during the session.
- **clickTimingMin** (int) — minimum time (in milliseconds) between consecutive clicks.
- **clickTimingMax** (int) — maximum time (in milliseconds) between consecutive clicks.
- **clickTimingAvg** (int) — average time (in milliseconds) between clicks during the session.
- **clickTimingCount** (int) — number of measured intervals between clicks (useful for avg normalization).
- **scrolls** (int) — total number of scroll events (wheel, touch, keys) during the session.
- **scrollTimingMin** (int) — minimum time (in ms) between scroll events.
- **scrollTimingMax** (int) — maximum time (in ms) between scroll events.
- **scrollTimingAvg** (int) — average time (in ms) between scroll events.
- **scrollTimingCount** (int) — number of recorded scroll intervals.
- **textInputEvents** (int) — number of text input events (keydown, input, etc.).
- **textInputTimingMin** (int) — minimum time (in ms) between characters during input.
- **textInputTimingMax** (int) — maximum time (in ms) between characters during input.
- **textInputTimingAvg** (int) — average input speed (in ms per character).
- **textInputTimingCount** (int) — number of recorded input intervals (pairs of key presses).
- **sessionDuration** (int) — session duration in milliseconds from its start.
- **userAgent** (string) — browser User-Agent string containing client information.
- **language** (string) — browser's preferred language (e.g., "ru-RU", "en-US").
- **platform** (string) — device platform (e.g., "Win32", "Linux x86_64", "MacIntel").
- **screenWidth** (int) — device screen width in pixels.
- **screenHeight** (int) — device screen height in pixels.
- **timezone** (string) — client timezone in IANA format (e.g., "Europe/Moscow").
- **cookiesEnabled** (bool) — flag indicating whether cookies are enabled in the browser.
- **onLine** (bool) — network status flag: true if the browser considers itself connected to the internet.
- **deviceMemory** (int) — device RAM size in gigabytes (estimate, not available in all browsers).
- **maxTouchPoints** (int) — maximum number of simultaneous touch points (0 — no touch screen; 1 or more — touch device).
- **browserName** (string) — browser name (e.g., "Chrome", "Firefox", "Safari").
- **browserVersion** (string) — browser version (e.g., "125.0.0").
- **osName** (string) — operating system name (e.g., "Windows", "Android", "iOS").
- **osVersion** (string) — operating system version (e.g., "10", "14.5").

The client is uniquely identified by one of the cookies. The script sends metrics along with browser cookies; to properly process metrics on the server side, you need to configure the name of the cookie from which the client identifier is extracted.

## API

The service provides the following REST API endpoints:

- **POST /api/v1/traces** — accept a new trace
- **GET /api/v1/scores/{token}** — retrieve score by token
- **GET /static/...** — serve static files (if enabled)

## Build

### Server

```bash
# Build
go build -o bean cmd/bean/main.go

# Run
./bean --config config.yaml
```

### Script

The script can be embedded on a page using the following tag:

```html
<script src="/static/collector.js"></script>
```

After that, create an instance of the collector:

```js
const collector = new BehavioralMetricsCollector({
    enableLogging: false,   
    reportInterval: 5000,
    skipEmpty: true,
    address: "/api/v1/traces",
});
```

## Configuration

Bean is configured through a YAML configuration file. Below is a detailed description of all parameters, their purposes, and allowed values.

### General Structure

```yaml
logger:
  level: info

server:
  address: ":8080"
  static: "./public"

analysis:
  token: token
  traces_length: 10
  traces_ttl: 10m
  scorers:
    - type: ml
      model: default
      url: http://127.0.0.1:8000
    - type: rules
      rules: /etc/bean/rules.yaml

dataset:
  file: /var/log/bean/dataset.log
  size: 1024
  amount: 10
```

### logger

Settings for the logging component.

#### level (required)

Log detail level. Supported values (case-insensitive):
- debug — detailed logs for development
- info — informational messages (default)
- warn or warning — warnings
- error — critical errors only

### server

HTTP server parameters.

#### address (required)

Address and port where the server will run. Use :8080 to listen on all interfaces on port 8080.

#### static

Path to the directory with static files (e.g., collector.js). If specified, files will be available at the /static/ route.
Can be left empty if static file serving is not needed.

### analysis

Behavioral analysis settings.

#### token (required)

Cookie name used for session identification. Bean expects the client (browser) to send this cookie with each trace request. This is not a secret, just a key for binding session data.

#### scorers (обязательный)

The list of scores performing the analysis. The scores perform the analysis in the order in which they are specified. Possible types: ML and rule.
Example:

```yaml
- type: ml
  model: default
  url: http://127.0.0.1:8000

- type: rules
  rules: /etc/bean/rules.yaml
```

For ML scorer, you must specify the URL of the inference service and the model name. For rule, you must specify the path to the rules file. The file must exist and contain the correct rules in the CEL language.

#### traces_length

Maximum number of traces stored per session. When exceeded, old traces are deleted (FIFO). Recommended value: 20–100, depending on sending frequency.

#### traces_ttl

Time after which a session is considered inactive and removed from memory.

Supported units:
- s — seconds
- m — minutes
- h — hours

### dataset

Dataset collection settings. This is optional parameter. If it is defined, then all received traces will be written to the dataset file.

#### file

Dataset file path

#### size

Maximum dataset file size.

#### amount

Amount of storing datasets.

### Environment Variables

Bean automatically supports parameter overriding through environment variables. Priority: environment variables > YAML values. Variable names are formed according to the pattern:

```bash
LOGGER_LEVEL=debug
SERVER_ADDRESS=:9090
ANALYSIS_TRACES_TTL=30m
```

### Configuration Validation

Bean validates the configuration on startup:

- All required fields must be specified.
- The logging level must be valid.
- The rules file must exist.
- On error, startup is stopped with a problem description.

Use `bean --config config.yaml` to load the config from a file.

## Rules

Bean uses **Common Expression Language (CEL)** to describe behavioral analysis rules. Rules allow you to evaluate user actions and assign scores for behavioral patterns (e.g., automation, bots).

When requesting a score, all collected traces are analyzed and scores are assigned. If a trace satisfies a rule, the scores are changed by the specified values.

### Rule File Format

Rules are defined in a **YAML file**, which is loaded when the server starts. The file contains a list of rules; each rule consists of a condition and score increments:

- when — condition in CEL language (should return true or false)
- then — object with scores that will be added to the final result

```yaml
- when: mouseMoves > 10 && clicks > 5
  then:
    human: 0.3
    automation: -0.1

- when: deviceMemory < 2
  then:
    automation: 0.2
```

### Variables

The following variables can be used in when expressions:

| Metric | Type | Description |
|--------|------|-------------|
| mouseMoves | int | Number of mouse movements |
| clicks | int | Number of clicks |
| clickTimingMin | int | Minimum delay between clicks (ms) |
| clickTimingMax | int | Maximum delay between clicks (ms) |
| clickTimingAvg | int | Average delay between clicks (ms) |
| clickTimingCount | int | Number of measured click intervals |
| scrolls | int | Number of scrolls |
| scrollTimingMin | int | Minimum delay between scrolls |
| scrollTimingMax | int | Maximum delay between scrolls |
| scrollTimingAvg | int | Average delay between scrolls |
| scrollTimingCount | int | Number of measured scroll intervals |
| textInputEvents | int | Number of text input events |
| textInputTimingMin | int | Minimum delay between characters |
| textInputTimingAvg | int | Average delay between characters |
| textInputTimingMax | int | Maximum delay between characters |
| textInputTimingCount | int | Number of measured input intervals |
| sessionDuration | int | Session duration (ms) |
| userAgent | string | Full User-Agent string |
| language | string | Browser language (e.g., ru-RU) |
| platform | string | Platform (e.g., Win32) |
| screenWidth | int | Screen width (px) |
| screenHeight | int | Screen height (px) |
| timezone | string | Timezone (e.g., Europe/Moscow) |
| cookiesEnabled | bool | Are cookies enabled |
| onLine | bool | Internet connection status |
| deviceMemory | int | Estimated RAM in GB |
| maxTouchPoints | int | Maximum number of touch points |
| browserName | string | Browser name (Chrome, Firefox, etc.) |
| browserVersion | string | Browser version |
| osName | string | Operating system name (Windows, Android, etc.) |
| osVersion | string | Operating system version |

### Expression Syntax (CEL)

#### Conditions (when)

If the when condition returns true, Bean adds the specified scores to the final result. Conditions use CEL syntax. CEL supports:

- Arithmetic operations: `+`, `-`, `*`, `/`, `%`
- Logical operations: `&&`, `||`, `!`
- Comparisons: `==`, `!=`, `<`, `>`, `<=`, `>=`
- String methods: `browserName == "Chrome"`, `language.startsWith("en")`, etc.

Learn more about CEL: https://github.com/google/cel-spec

#### Scores (then)

Each score is a key (string) and a value from 0.0 to 1.0.
All scores are summed by key but limited to the range [0.0, 1.0].

### Rule Examples

1. Suspicious lack of activity

```yaml
- when: mouseMoves < 3 && sessionDuration > 30000
  then:
    inactive: 0.8
```

2. Fast text input

```yaml
- when: textInputTimingAvg < 80 && textInputEvents > 5
  then:
    automation: 0.7
```

3. Device with little memory

```yaml
- when: deviceMemory < 2
  then:
    device: 0.6
```

4. No scrolling

```yaml
- when: scrolls == 0 && sessionDuration > 10000
  then:
    automation: 0.5
```

5. Headless Chrome

```yaml
- when: browserName.contains("HeadlessChrome")
  then:
    automation: 1.0
```

### Important Notes

- Rules are applied to each trace — if a user sent 10 traces, each rule is checked 10 times.
- Scores accumulate — if two rules fire, their then-values are summed.
- Maximum score value per key — 1.0 — score cannot exceed 1.0 (saturation).
- Expression error — causes the rule to be skipped (does not stop analysis).
- Rule order — not important, but it is recommended to group by logic.