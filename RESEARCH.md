# Entropy-Based Filtering Research

## Problem Statement
Pure regex pattern matching produces false positives on high-entropy strings that look like secrets but aren't (e.g., base64-encoded common words, long random-looking identifiers that are actually safe).

## Solution
Add Shannon entropy calculation to filter matches. Real secrets (API keys, tokens, passwords) have high entropy. False positives typically have lower entropy.

## Shannon Entropy Formula
```
H = -Σ p(x) * log2(p(x))
```
where p(x) is the frequency of byte x in the string.

## Entropy Thresholds
- 0-2.0: Very low (common words, false positives)
- 2.0-3.0: Low (likely false positives)
- 3.0-4.0: Medium (possible secrets)
- 4.0+: High (likely real secrets)

## References
- TruffleHog entropy filtering
- GitLeaks false positive reduction techniques
- NIST recommendations on secret detection
