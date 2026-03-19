export const sddTemplates: Record<string, string> = {
  proposal: `
    <h2>Why</h2>
    <p class="placeholder">Explain the motivation for this change. What problem does this solve?</p>
    <h2>What Changes</h2>
    <p class="placeholder">Describe what will change. Be specific about new capabilities, modifications, or removals.</p>
    <h2>Capabilities</h2>
    <h3>New Capabilities</h3>
    <p class="placeholder">List new capabilities being introduced</p>
    <h3>Modified Capabilities</h3>
    <p class="placeholder">List existing capabilities being changed</p>
    <h2>Impact</h2>
    <p class="placeholder">Affected code, APIs, dependencies, systems</p>
  `,
  design: `
    <h2>Context</h2>
    <p class="placeholder">Background and current state</p>
    <h2>Goals / Non-Goals</h2>
    <p><strong>Goals:</strong></p>
    <p class="placeholder">What this design aims to achieve</p>
    <p><strong>Non-Goals:</strong></p>
    <p class="placeholder">What is explicitly out of scope</p>
    <h2>Decisions</h2>
    <p class="placeholder">Key design decisions and rationale</p>
    <h2>Risks / Trade-offs</h2>
    <p class="placeholder">Known risks and trade-offs</p>
  `,
  spec: `
    <h2>ADDED Requirements</h2>
    <h3>Requirement: (name)</h3>
    <p class="placeholder">Describe the requirement using SHALL/MUST</p>
    <h4>Scenario: (name)</h4>
    <ul>
      <li><strong>WHEN</strong> (condition)</li>
      <li><strong>THEN</strong> (expected outcome)</li>
    </ul>
  `,
  tasks: `
    <h2>1. (Task Group Name)</h2>
    <ul data-type="taskList">
      <li data-type="taskItem" data-checked="false">1.1 (Task description)</li>
      <li data-type="taskItem" data-checked="false">1.2 (Task description)</li>
    </ul>
    <h2>2. (Task Group Name)</h2>
    <ul data-type="taskList">
      <li data-type="taskItem" data-checked="false">2.1 (Task description)</li>
      <li data-type="taskItem" data-checked="false">2.2 (Task description)</li>
    </ul>
  `,
};

export type TemplateType = keyof typeof sddTemplates;
