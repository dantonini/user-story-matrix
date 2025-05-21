# The Elaborate Command

The `elaborate` command is a powerful tool for transforming vague feature descriptions into well-defined user stories following the INVEST criteria. This guide explains how to use it effectively.

## Purpose

The purpose of the `elaborate` command is to:

1. Take a high-level, potentially vague feature description
2. Guide you through a structured workflow to refine it
3. Produce well-defined, prioritized user stories with acceptance criteria
4. Ensure the user stories follow the INVEST principles (Independent, Negotiable, Valuable, Estimable, Small, Testable)

## Usage

### Basic Usage

```bash
# Start with interactive input
usm elaborate

# Use an existing vague description file
usm elaborate docs/vague/my-feature.md

# Reset the workflow and start from the beginning
usm elaborate --reset docs/vague/my-feature.md

# Use a custom workflow
usm elaborate --workflow=my-custom-workflow docs/vague/my-feature.md
```

### What Happens Behind the Scenes

When you run the `elaborate` command:

1. If no file is provided, you'll be prompted to enter a vague description
2. The description is saved to `docs/vague/elaborate_YYYYMMDD_HHMMSS.md`
3. The workflow begins, guiding you through each step
4. Progress is saved automatically, so you can continue later
5. The final output includes detailed, well-structured user stories

### Options

- `--reset`: Reset the workflow and start from the beginning
- `--workflow=NAME`: Use a custom workflow by name
- `--workflow-path=PATH`: Use a custom workflow from a specific path
- `--debug`: Enable debug output with additional information

## Workflow Steps

The elaborate workflow consists of 7 steps:

1. **Analyze Description**: Analyze the vague description to identify the main problem, core functionality, potential users, etc.
2. **Extract Personas**: Define clear user personas who would interact with the feature
3. **Draft User Stories**: Create initial user stories based on the personas and analysis
4. **Refine User Stories**: Ensure the user stories meet the INVEST criteria
5. **Prioritize User Stories**: Assign priority levels to each story (Must Have, Should Have, Could Have, Won't Have)
6. **Add Acceptance Criteria**: Create detailed, testable acceptance criteria for each story
7. **Final Review**: Review and package all the user stories for implementation

## Example

Let's say you have a vague idea for a "User Access Management Feature" but haven't fully defined the requirements.

```bash
# Start the elaborate workflow
usm elaborate

# When prompted, enter your vague description
```

The command will guide you through each step, ultimately producing detailed user stories such as:

```
Title: Role-Based Access Control
As an administrator,
I want to assign predefined roles to users,
So that I can efficiently manage permissions based on job functions.

Priority: Must Have

Acceptance Criteria:
1. Given I am logged in as an administrator, when I view a user's profile, then I should see an option to assign roles.
2. Given I am assigning a role to a user, when I select a role, then I should see a preview of the permissions included.
3. Given I have assigned a role to a user, when the user logs in, they should only have access to the features permitted by their role.
```

## Continuing a Session

If you need to stop in the middle of the workflow, simply note the file path shown in the output. You can continue later by running:

```bash
usm elaborate docs/vague/elaborate_YYYYMMDD_HHMMSS.md
```

## Tips for Best Results

1. Be as detailed as possible in your initial vague description
2. Think about different types of users who might interact with the feature
3. Consider both functional and non-functional requirements
4. Focus on user needs rather than implementation details
5. Be specific about acceptance criteria to ensure testability 