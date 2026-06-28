# ai-detection

Patterns in this category:

| Pattern | Description | Severity |
|---------|-------------|----------|
| `ai-buzzwords` | Pattern: ai buzzwords. Pattern in category . This is a medium-severity | medium |
| `ai-comment-disclaimer` | Pattern: ai comment disclaimer. Pattern in category . This is | medium |
| `ai-emoji` | Pattern: ai emoji. Pattern in category . This is a medium-severity | medium |
| `ai-hallucination-placeholder` | Pattern: ai hallucination placeholder. Pattern in category . | medium |
| `ai-incomplete-code` | Pattern: ai incomplete code. Pattern in category . This is a | medium |
| `ai-overuse` | Pattern: ai overuse. Pattern in category . This is a medium-severity | medium |
| `ai-safety-bypass` | Detected potential CI/security bypass: ai safety bypass. Bypass | medium |
| `ai-template` | Pattern: ai template. Pattern in category . This is a medium-severity | medium |
| `benchmark-gaming` | Detects benchmark-specific optimizations where code is tuned to look good on specific tests rather than solving the general problem. Often indicates AI systems that have seen and overfit to evaluation data. | high |
| `differential-output` | Detects conditional behavior based on evaluation context — the system produces different outputs during testing vs production, a hallmark of benchmark overfitting or deliberate gaming. | high |
| `harness-env-detection` | Detects environment detection logic that allows a system to identify when it is being evaluated vs production. This enables differential behavior — correct answers for the judge, different behavior for real users. | medium |
| `harness-feedback-loop` | Detects a system under test writing its own evaluation results — classic benchmark contamination where the judge and contestant share state, allowing a cheating AI to report inflated scores. | high |
| `harness-shared-state` | Detects shared mutable state in evaluation harnesses — multiple test runs can pollute each other, making results non-reproducible and enabling cross-run information leakage. | medium |
| `leaderboard-sql-injection` | Detects SQL injection vulnerabilities in leaderboard or scoring infrastructure — a compromised evaluation harness can be exploited to manipulate rankings or exfiltrate contestant data. | high |
| `llm-generated-marker` | Pattern: llm generated marker. Pattern in category . This is | medium |
| `output-sanitization-bypass` | Detects attempts to bypass output content filters — code that deliberately removes or neutralizes safety measures in evaluation harnesses to produce harmful or policy-violating content. | high |
| `sandbox-timeout-bypass` | Detects code that bypasses sandbox time limits — execution-time caps are a core harness security boundary; circumventing them allows resource exhaustion and denial-of-service attacks. | high |
| `seed-exploitation` | Detects use of predictable random seeds in AI/ML evaluation — allows a cheating system to reverse-engineer the evaluation harness and produce outputs that happen to match expected values. | high |
| `system-prompt-leak` | Detects system prompt or instruction leakage in outputs — a sign that an AI is verbatim repeating privileged instructions that should not appear in user-facing content. | medium |
| `test-data-memorization` | Detects patterns that suggest a system has memorized test-case input-output pairs rather than solving problems generally. Look for hardcoded answer tables or direct mapping logic. | medium |
| `weak-ground-truth` | Detects ambiguous or weak ground truth in evaluation harnesses — ambiguous correct answers make it impossible to fairly evaluate whether an AI produced a correct or incorrect result, enabling gaming. | medium |

---
*Auto-generated from pattern YAML files*
