// Copyright (c) 2025 User Story Matrix
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package workflow

// UsmElaborateWorkflowName is the name of the elaborate workflow
const UsmElaborateWorkflowName = "usm-elaborate"

// UsmElaborateWorkflowSteps defines the predefined sequence of steps in the elaborate workflow
var UsmElaborateWorkflowSteps = []WorkflowStep{
	{
		ID:          "01-analyze-description",
		Description: "Analyze the vague description and identify potential user stories",
		Prompt: `# Step 1: Analyze the Vague Description

## Input
You are being provided with a vague feature description in the file ` + "`{{.ChangeRequestFilePath}}`" + `. 

## Task
Your task is to analyze this description and identify:

1. The main problem or need that this feature aims to address
2. The core functionality required
3. Potential user types (personas) who would use this feature
4. Potential acceptance criteria or constraints
5. Dependencies or integration points with existing features
6. Any potential edge cases or special scenarios

## Guidelines
- Be thorough but concise in your analysis
- Identify any ambiguities or missing information in the description
- Think about both functional and non-functional requirements
- Consider different user perspectives and use cases

## Expected Output
Provide a structured analysis with clear sections for each of the areas identified above. This analysis will be used as input for the next steps in generating well-defined user stories.

Begin by reading the content of the vague feature description, then create a new file named ` + "`{{.ChangeRequestFilePath}}.processed.md`" + `:
- write the vague feature description in the section ` + "`# Initial vague description`" + `
- write your analysis in the section ` + "`# Analysis`" + `.`,
	},
	{
		ID:          "02-extract-personas",
		Description: "Extract user personas from the description",
		Prompt: `# Step 2: Extract User Personas

## Input
You have analyzed the vague feature description and now need to identify and define the user personas who will interact with this feature.

## Task
Based on your analysis from the previous step, identify and define the key user personas who would use this feature. For each persona:

1. Define their role and responsibilities
2. Identify their primary goals and motivations
3. Describe their technical skill level and context
4. Outline their pain points that this feature should address
5. Consider their workflow and how this feature fits into it

## Guidelines
- Focus on the most important personas - typically 2-4 personas are sufficient
- Make personas specific and realistic, not generic
- Consider both primary users (who directly use the feature) and secondary users (who are affected by it)
- Think about different skill levels, contexts, and use cases
- Ensure personas are distinct from each other with clear differences

## Expected Output
Update the file ` + "`{{.ChangeRequestFilePath}}.processed.md`" + ` by adding a new section ` + "`# User Personas`" + ` that includes:

1. A brief introduction explaining the identified personas
2. For each persona, provide:
   - Name and role
   - Background and context
   - Goals and motivations
   - Technical skill level
   - Pain points this feature should address
   - How they would interact with the feature

This persona definition will guide the creation of user stories in the next steps.`,
	},
	{
		ID:          "03-draft-user-stories",
		Description: "Draft initial user stories based on the analysis",
		Prompt: `# Step 3: Draft Initial User Stories

## Input
You have analyzed the feature description and identified user personas. Now it's time to draft initial user stories.

## Task
Based on your analysis and the identified personas, create initial user stories that capture the key functionality needed. For each user story:

1. Use the standard format: "As a [persona], I want [goal] so that [benefit]"
2. Focus on the core functionality first
3. Ensure each story is independent and valuable
4. Keep stories at an appropriate level of granularity
5. Consider the user's workflow and journey

## Guidelines
- Start with the most important functionality for each persona
- Each story should deliver value to the user
- Stories should be testable and have clear acceptance criteria potential
- Avoid technical implementation details in the story description
- Focus on the "what" and "why", not the "how"
- Consider both happy path and important alternative scenarios

## Expected Output
Update the file ` + "`{{.ChangeRequestFilePath}}.processed.md`" + ` by adding a new section ` + "`# Initial User Stories`" + ` that includes:

1. A list of user stories in the standard format
2. For each story, include:
   - The user story statement
   - A brief description of what it accomplishes
   - Which persona(s) it serves
   - Any initial thoughts on acceptance criteria

These initial stories will be refined in the next step to ensure they meet INVEST criteria.`,
	},
	{
		ID:          "04-refine-user-stories",
		Description: "Refine user stories to ensure they meet INVEST criteria",
		Prompt: `# Step 4: Refine User Stories to Meet INVEST Criteria

## Input
You have drafted initial user stories. Now you need to refine them to ensure they meet the INVEST criteria.

## Task
Review and refine each user story to ensure it meets the INVEST criteria:

- **Independent**: Stories should be self-contained and not depend on other stories
- **Negotiable**: Stories should be flexible and open to discussion about implementation
- **Valuable**: Each story should deliver clear value to the user
- **Estimable**: Stories should be clear enough to estimate effort
- **Small**: Stories should be small enough to complete in a reasonable timeframe
- **Testable**: Stories should have clear criteria for determining when they're done

## Guidelines
- Split large stories into smaller, more manageable ones
- Combine very small stories if they don't provide value independently
- Clarify vague or ambiguous stories
- Ensure each story can be developed and tested independently
- Remove or merge duplicate functionality
- Make sure the value proposition is clear for each story

## Expected Output
Update the file ` + "`{{.ChangeRequestFilePath}}.processed.md`" + ` by adding a new section ` + "`# Refined User Stories`" + ` that includes:

1. The refined list of user stories that meet INVEST criteria
2. For each story, include:
   - The refined user story statement
   - Explanation of how it meets INVEST criteria
   - Any changes made from the initial version
   - Dependencies or relationships with other stories (if any)

Also include a brief section explaining what changes were made during refinement and why.`,
	},
	{
		ID:          "05-prioritize-user-stories",
		Description: "Prioritize the refined user stories",
		Prompt: `# Step 5: Prioritize User Stories

## Input
You have refined user stories that meet INVEST criteria. Now you need to prioritize them for implementation.

## Task
Prioritize the user stories based on:

1. **Business Value**: How much value does this story deliver to users and the business?
2. **User Impact**: How many users will be affected and how significantly?
3. **Dependencies**: Are there stories that must be implemented before others?
4. **Risk**: Are there stories that address high-risk areas or assumptions?
5. **Effort**: Consider the estimated effort required (simpler stories might be prioritized for early wins)

## Guidelines
- Use a clear prioritization scheme (e.g., High/Medium/Low or MoSCoW)
- Consider the Minimum Viable Product (MVP) - what's the smallest set of stories that delivers value?
- Think about user workflows - prioritize stories that complete important user journeys
- Balance quick wins with foundational work
- Consider technical dependencies and logical implementation order
- Involve stakeholder perspectives in prioritization decisions

## Expected Output
Update the file ` + "`{{.ChangeRequestFilePath}}.processed.md`" + ` by adding a new section ` + "`# Prioritized User Stories`" + ` that includes:

1. User stories organized by priority level
2. For each priority level, explain the rationale
3. Identification of the MVP set of stories
4. Any dependencies or recommended implementation order
5. Brief explanation of the prioritization criteria used

Consider grouping stories into releases or iterations if appropriate.`,
	},
	{
		ID:          "06-acceptance-criteria",
		Description: "Add acceptance criteria to each user story",
		Prompt: `# Step 6: Add Acceptance Criteria to User Stories

## Input
You have prioritized user stories. Now you need to add detailed acceptance criteria to each story.

## Task
For each user story, define clear acceptance criteria that:

1. Specify the conditions that must be met for the story to be considered complete
2. Cover both functional and non-functional requirements
3. Include happy path scenarios and important edge cases
4. Are testable and verifiable
5. Provide enough detail for developers and testers

## Guidelines
- Use clear, unambiguous language
- Write criteria from the user's perspective
- Include both positive and negative test cases
- Consider error handling and edge cases
- Make criteria specific and measurable
- Ensure criteria are realistic and achievable
- Consider accessibility, performance, and security where relevant

## Expected Output
Update the file ` + "`{{.ChangeRequestFilePath}}.processed.md`" + ` by adding a new section ` + "`# User Stories with Acceptance Criteria`" + ` that includes:

1. Each user story with its complete set of acceptance criteria
2. For each story, organize acceptance criteria by:
   - Functional requirements
   - Non-functional requirements (if applicable)
   - Edge cases and error handling
   - Any specific constraints or assumptions

Use a clear format such as:
- Given/When/Then scenarios
- Bullet points with clear conditions
- Numbered criteria for easy reference

Ensure that each criterion is testable and provides clear guidance for implementation and testing.`,
	},
	{
		ID:          "07-final-review",
		Description: "Final review and packaging of the user stories",
		Prompt: `# Step 7: Final Review and Packaging of User Stories

## Input
You've created a set of prioritized user stories with detailed acceptance criteria. Now it's time for a final review and to package everything together.

## Task
Your task is to review all the user stories and their acceptance criteria, make any final improvements, and package them in a format that can be easily used by the development team.

1. Review all user stories for consistency in format and level of detail
2. Check for any remaining dependencies or conflicts between stories
3. Verify that all acceptance criteria are clear and testable
4. Ensure the prioritization makes sense for implementation
5. Format everything in a clear, structured manner

## Guidelines
- Look for any inconsistencies in terminology across stories
- Identify any gaps in the feature coverage
- Make sure each story stands on its own but fits into the overall feature
- Ensure user stories are properly formatted and easy to understand
- Check that acceptance criteria are meaningful and cover all necessary aspects
- Consider the implementation flow and dependencies one final time

## Expected Output
Produce a final document that includes:

1. A brief summary of the feature being elaborated
2. The identified user personas and their key needs
3. The complete set of user stories, organized by priority
4. For each story, include:
   - Title
   - User story in standard format
   - Priority level
   - Acceptance criteria
   - Any implementation notes or dependencies

This final document will serve as a comprehensive specification for implementing the feature, broken down into clearly defined, prioritized user stories with testable acceptance criteria. 

For each user story, create a new file in the directory ` + "`docs/user-stories`" + ` with the following name: ` + "`<UserStoryID>-<title-of-the-user-story>.md`" + `.

Write the user story in the file, using the following format:

` + "```" + `
{{template "shared/user-story-format.md"}}
` + "```" + ``,
	},
}

// createUsmElaborateStandardWorkflow creates the elaborate workflow definition
func createUsmElaborateStandardWorkflow() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        UsmElaborateWorkflowName,
		Description: "Workflow for refining vague feature descriptions into well-defined INVEST user stories",
		Steps:       UsmElaborateWorkflowSteps,
	}
} 