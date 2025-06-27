---
applyTo: "**/*_test.go"
---
# Guidelines for Go Tests

Apply the [go coding guidelines](./go.instructions.md) to go code.
Read the [requirements](../../docs/requirements.md) for the expected behavior.

## Philosophy
- Write effective tests
- Quality over quantity
- Mock the state
- Avoid mocking business logic

## Dependencies
- Use https://github.com/stretchr/testify for testing
