// ── Domain types matching Go DTOs ──

export interface ObjectDTO {
  id: string;
  project: string;
  class: string;
  sequence?: string;
  version: string;
  revision: string;
  state: string;
  title: string;
  metadata?: Record<string, unknown>;
  relations?: RelationDTO[];
  artifacts?: ArtifactDTO[];
}

export interface RelationDTO {
  to: string;
  type: string;
  quantity?: number;
  unit?: string;
}

export interface ArtifactDTO {
  id: string;
  kind: string;
  role: string;
  path: string;
  original_name?: string;
  checksum?: string;
  size_bytes?: number;
  generated: boolean;
  required: boolean;
  status: string;
}

export interface ObjectDocumentDTO {
  object_id: string;
  frontmatter: string;
  body: string;
}

export interface UpdateObjectDocumentRequest {
  object_id: string;
  frontmatter: string;
  body: string;
}

export interface ProjectCreateRequest {
  path: string;
  code: string;
  title: string;
}

export interface HistoryEntry {
  event_id?: string;
  time?: string;
  actor?: string;
  type?: string;
  object_id?: string;
  payload?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface TreeNodeDTO {
  id: string;
  class: string;
  title: string;
  state: string;
  icon: string;
  status_color: string;
  has_children: boolean;
  children_count: number;
  warnings: number;
  blockers: number;
  thumbnail_url?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateObjectRequest {
  project_path: string;
  class: string;
  title: string;
  parent_object_id?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateObjectResponse {
  object_id: string;
  diagnostics?: Diagnostic[];
}

export interface TransitionRequest {
  object_id: string;
  transition: string;
}

export interface TransitionResponse {
  new_state: string;
  diagnostics?: Diagnostic[];
}

export interface BOMRequest {
  root_id: string;
  flat: boolean;
}

export interface BOMResponse {
  root_id: string;
  rows: BOMRowDTO[];
}

export interface BOMRowDTO {
  row_id: string;
  parent_id: string;
  child_id: string;
  child_class: string;
  quantity: number;
  unit: string;
  position?: string;
  make_buy?: string;
  level: number;
}

export interface SearchRequest {
  query: string;
  limit: number;
}

export interface SearchResult {
  object_id: string;
  title: string;
  class: string;
  state: string;
  excerpt?: string;
}

export interface TreeRequest {
  filter?: string;
  sort?: string;
  offset: number;
  limit: number;
}

export interface Diagnostic {
  id: string;
  severity: string;
  code: string;
  object_id: string;
  path?: string;
  message: string;
  suggested_actions?: string[];
}

export interface ReleaseResponse {
  ready: boolean;
  blockers?: Diagnostic[];
}

export interface ProjectInfo {
  code: string;
  title: string;
  description: string;
  version: number;
  root: string;
}

export interface TransitionDef {
  name: string;
  to: string;
  guards: string[];
  effects: string[];
}

export type ViewMode = "browse" | "edit" | "bom" | "search" | "stdparts" | "gitchanges" | "release";

export const CLASS_ICONS: Record<string, string> = {
  prt: "component",
  asm: "package",
  drw: "ruler",
  doc: "file-text",
  std: "nut",
  mat: "brick-wall",
  cut: "scissors",
  bnd: "corner-down-right",
  nc: "cpu",
  ins: "search-check",
  tpc: "clipboard-list",
  wi: "list-checks",
  bom: "table",
  rel: "tag",
  cr: "git-pull-request",
};

export const CLASS_LABELS: Record<string, string> = {
  prt: "Part",
  asm: "Assembly",
  drw: "Drawing",
  doc: "Document",
  std: "Standard Part",
  mat: "Material",
  cut: "Cutting File",
  bnd: "Bending File",
  nc: "NC Program",
  ins: "Inspection",
  tpc: "Tech Process Card",
  wi: "Work Instruction",
  bom: "Formal BOM",
  rel: "Release Package",
  cr: "Change Request",
};

export const STATE_COLORS: Record<string, string> = {
  draft: "gray",
  in_review: "amber",
  approved: "blue",
  released: "green",
  blocked: "red",
  obsolete: "slate",
  archived: "slate",
};

export const STATE_LABELS: Record<string, string> = {
  draft: "Draft",
  in_review: "In Review",
  approved: "Approved",
  released: "Released",
  blocked: "Blocked",
  obsolete: "Obsolete",
  archived: "Archived",
};
