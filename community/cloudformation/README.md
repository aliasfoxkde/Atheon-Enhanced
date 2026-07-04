# AWS CloudFormation Security Patterns

Detects security misconfigurations in AWS CloudFormation templates.

## Patterns

- `cloudformation-s3-public-access`: Detects S3 buckets with public access
- `cloudformation-s3-no-encryption`: Detects S3 buckets without server-side encryption
- `cloudformation-iam-lambda-assume-role`: Detects overly permissive IAM assume roles

## References

- [AWS CloudFormation Security Best Practices](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/security-best-practices.html)
- [CIS AWS Foundations Benchmark](https://www.cisecurity.org/benchmark/aws)
