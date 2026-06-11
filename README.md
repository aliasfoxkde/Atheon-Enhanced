# Leakr

![Scanners](https://img.shields.io/badge/scanners-5-blue)
![Java](https://img.shields.io/badge/Java-17%2B-orange)
![Build](https://img.shields.io/badge/build-Maven-red)
![License](https://img.shields.io/badge/license-proprietary-lightgrey)


**A secret and credential scanner — a Java library you embed in your code and a CLI tool you run anywhere.**

---

## What is Leakr?

Leakr scans files, directories, environment variables, and stdin for leaked secrets: API keys, tokens, and credentials. It does one thing and does it well.

It works two ways:

- **As a library** — call `runner.scanString(...)` from your Java application, CI pipeline, or review tooling. No subprocess. No native binary. Just a dependency.
- **As a CLI tool** — run `leakr scan <path>` from any terminal on any platform with a JRE.

The scanner engine is the only thing that grows over time. The CLI, the library API, the output formats, the exit codes — all of it is done.

---

## Why Java?

Leakr is Java because Java is everywhere.

- **Embeddable** — drop it into any Maven or Gradle project. Call it from application code, not a shell.
- **One JAR, any platform** — runs on Windows, Linux, and macOS without architecture-specific binaries or package managers. If you have a JRE, you have Leakr.
- **Enterprise-native** — Java lives in CI pipelines, build servers, and backend services. Leakr lives there with it, no wrapper script required.
- **Parallel by default** — directory scans use a thread pool sized to the host CPU automatically.

---

## Why Leakr when gitleaks exists?

[gitleaks](https://github.com/gitleaks/gitleaks) is excellent at what it does: scanning git history for secrets that were committed. That is not what Leakr does.

|                              | Leakr          | gitleaks       |
|------------------------------|:--------------:|:--------------:|
| Scan git history             |                | ✓              |
| Scan files and directories   | ✓              | ✓              |
| Scan environment variables   | ✓              |                |
| Scan stdin / arbitrary text  | ✓              | partial        |
| **Embeddable Java library**  | **✓**          | **✗**          |
| Zero native dependencies     | ✓ (JRE only)   | ✗ (Go binary)  |
| Add a scanner = one class    | ✓              | config + rules |

If you want git history scanning, use gitleaks. If you want to embed secret detection in a Java application, scan environment variables at startup, integrate it into a CI step without a binary dependency, or pipe arbitrary content through a scanner — use Leakr.

---

## Quickstart

### Clone and build

```bash
git clone https://github.com/HoraDomu/Leakr.git
cd Leakr
mvn package -q
```

This produces `target/leakr.jar` — a single self-contained JAR with all dependencies bundled.

### Run the CLI

```bash
# Scan a directory
java -jar target/leakr.jar scan /path/to/project

# Scan a single file
java -jar target/leakr.jar scan config.env

# Scan environment variables
java -jar target/leakr.jar scan --env

# Pipe content from another command
git diff | java -jar target/leakr.jar scan --stdin

# JSON output (for CI parsing or downstream tooling)
java -jar target/leakr.jar scan /path/to/project --json

# Exclude directories
java -jar target/leakr.jar scan . --exclude target,dist,node_modules

# Limit to specific file extensions
java -jar target/leakr.jar scan . --ext .env,.yaml,.json,.tf

# List all registered scanners
java -jar target/leakr.jar list
```

Exit code `0` means clean. Exit code `1` means findings were detected. Pipe-friendly and CI-ready out of the box — no configuration required.

### Sample output

```
[CRITICAL] aws-access-key
  file:  src/config/dev.properties:14
  desc:  Detects AWS access key IDs (AKIA...)
  match: AKIA**********************MPLE

─────────────────────────────
found 1 potential secret(s)

files: 42  size: 318.7 KB  time: 84ms
```

---

## Use as a Library

Install to your local Maven repository:

```bash
mvn install -q
```

Add the dependency to your project:

```xml
<dependency>
    <groupId>leakr</groupId>
    <artifactId>leakr</artifactId>
    <version>1.0</version>
</dependency>
```

Then call it directly from your code:

```java
import leakr.core.*;
import java.nio.file.*;
import java.util.List;

Registry registry = new Registry();
Runner runner = new Runner(registry);

// Scan a string
List<Finding> findings = runner.scanString("AKIAIOSFODNN7EXAMPLE");

// Scan a file
List<Finding> findings = runner.scanFile(Path.of("config.yaml"));

// Scan a directory (parallel, skips binaries and common noise dirs automatically)
List<Finding> findings = runner.scanDir(Path.of("/path/to/repo"));

// Scan environment variables
List<Finding> findings = runner.scanEnv();
```

Each `Finding` exposes: `scanner`, `severity`, `file`, `line`, `description`, `match`.

---

## Registered Scanners

| Scanner | Detects | Severity |
|---|---|---|
| `aws-access-key` | AWS access key IDs (`AKIA...`) | CRITICAL |
| `github-pat` | GitHub personal access tokens (`ghp_...`) | HIGH |
| `stripe-secret-key` | Stripe secret keys (`sk_live_...`) | CRITICAL |
| `slack-bot-token` | Slack bot tokens (`xoxb-...`) | HIGH |
| `twilio-account-sid` | Twilio account SIDs (`AC...`) | MEDIUM |

---

## Adding a Scanner

Leakr auto-discovers every class in the `leakr.scanners` package that implements `Scanner`. No registration, no config file. Drop a class in, rebuild, done.

### 1. Create the class

```java
package leakr.scanners;

import leakr.core.*;
import java.util.*;
import java.util.regex.*;

public class MyServiceScanner implements Scanner {
    private static final Pattern PATTERN = Pattern.compile("myservice_[a-zA-Z0-9]{32}");

    public String name()        { return "myservice-api-key"; }
    public String description() { return "Detects MyService API keys"; }
    public Severity severity()  { return Severity.HIGH; }

    public List<String> scan(String input) {
        List<String> matches = new ArrayList<>();
        Matcher m = PATTERN.matcher(input);
        while (m.find()) matches.add(m.group());
        return matches;
    }
}
```

Place it in `src/leakr/scanners/MyServiceScanner.java`.

### 2. Rebuild

```bash
mvn package -q
```

### 3. Confirm it registered

```bash
java -jar target/leakr.jar list
```

Your scanner appears automatically. No other changes needed.

### 4. Add a test case

Open `src/leakr/test/ScannerTest.java` and add one entry to `CASES`:

```java
new Case("myservice-api-key",
    "myservice_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",  // must produce a match
    "myservice_tooshort")                            // must NOT produce a match
```

### 5. Run the test

```bash
java -cp target/leakr.jar leakr.test.ScannerTest
```

---

## Testing

`ScannerTest` is a universal, plug-and-play test runner. Each `Case` has three fields:

| Field | Purpose |
|---|---|
| `scanner` | The exact name returned by `scanner.name()` |
| `hit` | A string that **must** produce at least one match |
| `miss` | A string that **must not** produce any match |

Run all cases at once to confirm every scanner works and nothing regressed:

```bash
java -cp target/leakr.jar leakr.test.ScannerTest
```

```
PASS     aws-access-key
PASS     github-pat
PASS     stripe-secret-key
PASS     slack-bot-token
PASS     twilio-account-sid

5 passed, 0 failed
```

Exit code `0` on full pass, `1` on any failure.

---

## License

Copyright (c) 2026 Dominick McEvoy. All rights reserved. See [LICENSE](LICENSE) for terms.
