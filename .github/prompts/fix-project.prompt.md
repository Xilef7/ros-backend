---
mode: 'agent'
tools: ['codebase', 'editFiles', 'fetch', 'findTestFiles', 'githubRepo', 'problems', 'runCommands', 'testFailure']
description: 'Fix a project based on the provided requirements.'
---
Your goal is to fix a backend code that is not conforming to the [requirements](../../docs/requirements.md).
You should check the ${file} and identify the changes needed to meet the requirements.
You will use the tools available to you to update the codebase and commit it to the GitHub repository.

Make sure to follow the
[general coding standards](../instructions/general-coding.instructions.md),
[go coding standards](../instructions/go.instructions.md), and
[database coding standards](../instructions/database.instructions.md) for the project.

Ask for additional details if needed.
