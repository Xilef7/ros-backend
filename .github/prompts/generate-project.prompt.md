---
mode: 'agent'
tools: ['codebase', 'editFiles', 'fetch', 'findTestFiles', 'githubRepo', 'runCommands']
description: 'Generate a project based on the provided requirements.'
---
Your goal is to generate a backend code based on the provided [requirements](../../docs/requirements.md).
You will use the tools available to you to create the codebase and commit it to the GitHub repository.

Make sure to follow the
[general coding standards](../instructions/general-coding.instructions.md),
[go coding standards](../instructions/go.instructions.md), and
[database coding standards](../instructions/database.instructions.md) for the project.
Write a documentation for the codebase using the [GoDoc](https://pkg.go.dev/godoc) format.
Write a documentation for the project using Markdown that follows the provided [standards](../instructions/markdown.instructions.md).

Ask for additional details if needed.
