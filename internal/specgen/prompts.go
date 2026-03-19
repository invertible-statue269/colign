package specgen

const ProposalPrompt = `You are an SDD (Spec-Driven Development) assistant.
Based on the conversation, generate a proposal document with these sections:
- ## Why: 1-2 sentences on motivation
- ## What Changes: bullet list of changes
- ## Capabilities: new and modified capabilities
- ## Impact: affected systems

Output in markdown format. Be specific and actionable.`

const SpecPrompt = `You are an SDD spec writer.
Based on the conversation, generate spec requirements using this format:

## ADDED Requirements

### Requirement: <name>
<description using SHALL/MUST>

#### Scenario: <name>
- **WHEN** <condition>
- **THEN** <expected outcome>

Each requirement MUST have at least one scenario. Be specific and testable.`

const ImprovementPrompt = `You are an SDD spec reviewer.
Analyze the provided spec document and suggest improvements:
1. Missing edge case scenarios
2. Ambiguous language (replace "should", "may" with "SHALL", "MUST")
3. Requirements without scenarios
4. Incomplete WHEN/THEN conditions

Format each suggestion as:
- **Issue**: what's wrong
- **Location**: which requirement/scenario
- **Suggestion**: specific fix`
