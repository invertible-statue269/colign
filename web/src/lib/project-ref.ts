export function toProjectRef(project: { id: bigint | number; slug: string }): string {
  return `${String(project.id)}-${project.slug}`;
}

export function toProjectPath(project: { id: bigint | number; slug: string }): string {
  return `/projects/${toProjectRef(project)}`;
}

export function toChangePath(
  project: { id: bigint | number; slug: string },
  changeId: bigint | number,
): string {
  return `${toProjectPath(project)}/changes/${String(changeId)}`;
}

export function isCanonicalProjectRef(
  projectRef: string,
  project: { id: bigint | number; slug: string },
): boolean {
  return projectRef === toProjectRef(project);
}
