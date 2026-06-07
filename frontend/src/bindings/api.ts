// ── go-plm API bindings ──
// Auto-detects the best transport:
//   1. HTTP JSON-RPC (embedded local server mode)
//   2. Mock data (development without backend)

import type {
  ObjectDTO,
  TreeNodeDTO,
  CreateObjectRequest,
  CreateObjectResponse,
  TransitionRequest,
  TransitionResponse,
  BOMRequest,
  BOMResponse,
  SearchRequest,
  SearchResult,
  TreeRequest,
  Diagnostic,
  ReleaseResponse,
  ProjectInfo,
  TransitionDef,
  ObjectDocumentDTO,
  UpdateObjectDocumentRequest,
  ProjectCreateRequest,
  HistoryEntry,
} from "../types";

// ── Transport detection ──

type Transport = "http" | "mock";

function detectTransport(): Transport {
  const params = new URLSearchParams(window.location.search);
  return params.get("mock") === "1" ? "mock" : "http";
}

const transport: Transport = detectTransport();

// ── JSON-RPC 2.0 HTTP client ──

let requestId = 0;

async function callHTTP<T>(method: string, ...args: unknown[]): Promise<T> {
  const resp = await fetch("/api", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      jsonrpc: "2.0",
      id: ++requestId,
      method,
      params: args.length === 1 ? args[0] : args,
    }),
  });

  if (!resp.ok) {
    throw new Error(`HTTP ${resp.status}: ${resp.statusText}`);
  }

  const data = await resp.json();

  if (data.error) {
    throw new Error(data.error.message || "RPC error");
  }

  return data.result as T;
}

// ── Mock data (for development without backend) ──

const mockObjects: ObjectDTO[] = [
  {
    id: "demo-asm-0100-v1.0",
    project: "demo",
    class: "asm",
    sequence: "0100",
    version: "1.0",
    revision: "1.0",
    state: "draft",
    title: "Motor Assembly",
    metadata: { unit: "pcs", lifecycle_stage: "development" },
  },
  {
    id: "demo-prt-0001-v1.0",
    project: "demo",
    class: "prt",
    sequence: "0001",
    version: "1.0",
    revision: "1.0",
    state: "draft",
    title: "Bracket Motor",
    metadata: { unit: "pcs", make_buy: "make", material: "al5052", thickness_mm: 2.0 },
  },
  {
    id: "demo-prt-0002-v1.0",
    project: "demo",
    class: "prt",
    sequence: "0002",
    version: "1.0",
    revision: "1.0",
    state: "in_review",
    title: "Shaft Drive",
    metadata: { unit: "pcs", make_buy: "make", material: "steel_4140" },
  },
  {
    id: "demo-std-0001-v1.0",
    project: "demo",
    class: "std",
    sequence: "0001",
    version: "1.0",
    revision: "1.0",
    state: "approved",
    title: "Hex Bolt ISO 4017 M6x30 A2",
    metadata: { std_class: "fst", standard: "ISO 4017", thread: "M6", length_mm: 30, material_grade: "A2", make_buy: "buy" },
  },
  {
    id: "demo-drw-0001-v1.0",
    project: "demo",
    class: "drw",
    sequence: "0001",
    version: "1.0",
    revision: "1.0",
    state: "released",
    title: "Motor Assembly Drawing",
    metadata: { format: "A3", scale: "1:2" },
  },
];

const stateColorMap: Record<string, string> = {
  draft: "gray", in_review: "amber", approved: "blue",
  released: "green", blocked: "red", obsolete: "slate", archived: "slate",
};

const mockTreeNodes: TreeNodeDTO[] = mockObjects.map((o) => ({
  id: o.id,
  class: o.class,
  title: o.title,
  state: o.state,
  icon: classIcon(o.class),
  status_color: stateColorMap[o.state] || "gray",
  has_children: o.class === "asm",
  children_count: o.class === "asm" ? 3 : 0,
  warnings: 0,
  blockers: 0,
}));

function classIcon(c: string): string {
  const m: Record<string, string> = {
    asm: "package", prt: "component", drw: "ruler", doc: "file-text",
    std: "nut", mat: "brick-wall", cut: "scissors", bnd: "corner-down-right",
    nc: "cpu", ins: "search-check", tpc: "clipboard-list", wi: "list-checks",
    bom: "table", rel: "tag", cr: "git-pull-request",
  };
  return m[c] || "file";
}

// ── Unified call dispatcher ──

async function call<T>(method: string, ...args: unknown[]): Promise<T> {
  if (transport === "mock") {
    return mockCall<T>(method, args);
  }

  // Try HTTP first (it's the default standalone mode)
  return callHTTP<T>(method, ...args);
}

function mockCall<T>(method: string, args: unknown[]): T {
  switch (method) {
    case "ListObjects":
      return mockObjects as T;
    case "GetObject":
      return (mockObjects.find((o) => o.id === (args[0] as string)) || null) as T;
    case "GetTree":
      return mockTreeNodes as T;
    case "SearchObjects": {
      const q = ((args[0] as SearchRequest)?.query || "").toLowerCase();
      return mockObjects
        .filter((o) => o.title.toLowerCase().includes(q) || o.id.toLowerCase().includes(q))
        .map((o): SearchResult => ({ object_id: o.id, title: o.title, class: o.class, state: o.state })) as T;
    }
    case "GetBOM": {
      const req = args[0] as BOMRequest;
      return {
        root_id: req.root_id,
        rows: [
          { row_id: "r1", parent_id: req.root_id, child_id: "demo-prt-0001-v1.0", child_class: "prt", quantity: 1, unit: "pcs", position: "10", make_buy: "make", level: 0 },
          { row_id: "r2", parent_id: req.root_id, child_id: "demo-prt-0002-v1.0", child_class: "prt", quantity: 2, unit: "pcs", position: "20", make_buy: "make", level: 0 },
          { row_id: "r3", parent_id: req.root_id, child_id: "demo-std-0001-v1.0", child_class: "std", quantity: 4, unit: "pcs", position: "30", make_buy: "buy", level: 0 },
        ],
      } as T;
    }
    case "CreateObject":
      return { object_id: `demo-prt-${Date.now().toString(36)}-v1.0` } as T;
    case "RunTransition":
      return { new_state: "in_review" } as T;
    case "AvailableTransitions":
      return [
        { name: "submit_review", to: "in_review", guards: [], effects: [] },
        { name: "approve", to: "approved", guards: ["no_blocking_issues"], effects: [] },
        { name: "release", to: "released", guards: ["no_release_blockers"], effects: ["create_git_tag"] },
      ] as T;
    case "ProjectInfo":
      return { code: "demo", title: "Demo Project", description: "A demonstration PLM project", version: 1, root: "." } as T;
    case "GetNamingInfo":
      return { project_code: "demo", standard: "v10", pattern: "[project]-[class]-[sequence]-v[major].[minor]" } as T;
    case "GetNextSequence":
      return 1 as T;
    case "OpenProject":
      return { code: "demo", title: "Demo Project", description: "A demonstration PLM project", version: 1, root: args[0] as string } as T;
    case "CreateProject": {
      const req = args[0] as ProjectCreateRequest;
      return { code: req.code, title: req.title, description: "", version: 1, root: req.path } as T;
    }
    case "GetObjectDocument": {
      const id = args[0] as string;
      const obj = mockObjects.find((o) => o.id === id)!;
      return {
        object_id: id,
        frontmatter: `id: ${obj.id}\nproject: ${obj.project}\nclass: ${obj.class}\nsequence: ${obj.sequence}\nversion: ${obj.version}\nrevision: ${obj.revision}\nstate: ${obj.state}\ntitle: ${obj.title}\nmetadata: ${JSON.stringify(obj.metadata || {})}`,
        body: `# ${obj.title}\n\nDescribe the engineering object here.\n`,
      } as T;
    }
    case "ValidateObject":
      return [] as T;
    case "GetObjectHistory":
      return [] as T;
    case "Stats":
      return { objects: 5, relations: 4, artifacts: 2 } as T;
    case "CheckReleaseReadiness":
      return { ready: false, blockers: [] } as T;
    case "GitStatus":
      return { modified: [], added: [], deleted: [], clean: true } as T;
    default:
      console.warn(`[mock] Unimplemented: ${method}`);
      return null as T;
  }
}

// ── Exported API ──

export const api = {
  // Projects
  openProject: (path: string) => call<ProjectInfo>("OpenProject", path),
  createProject: (req: ProjectCreateRequest) => call<ProjectInfo>("CreateProject", req),

  // Objects
  listObjects: () => call<ObjectDTO[]>("ListObjects"),
  getObject: (id: string) => call<ObjectDTO | null>("GetObject", id),
  createObject: (req: CreateObjectRequest) => call<CreateObjectResponse>("CreateObject", req),
  deleteObject: (id: string, force = false) => call<void>("DeleteObject", id, force),
  searchObjects: (req: SearchRequest) => call<SearchResult[]>("SearchObjects", req),

  // Transitions
  runTransition: (req: TransitionRequest) => call<TransitionResponse>("RunTransition", req),
  availableTransitions: (objectId: string) => call<TransitionDef[]>("AvailableTransitions", objectId),

  // BOM
  getBOM: (req: BOMRequest) => call<BOMResponse>("GetBOM", req),
  validateBOM: (rootId: string) => call<Diagnostic[]>("ValidateBOM", rootId),
  detectBOMCycles: (rootId: string) => call<Diagnostic[]>("DetectBOMCycles", rootId),

  // Release
  checkReleaseReadiness: (rootId: string) => call<ReleaseResponse>("CheckReleaseReadiness", rootId),
  buildReleaseScope: (rootId: string) => call<string[]>("BuildReleaseScope", rootId),

  // Tree
  getTree: (req: TreeRequest) => call<TreeNodeDTO[]>("GetTree", req),

  // Git
  createCheckpoint: (message: string) => call<string>("CreateCheckpoint", message),
  gitStatus: () => call<{ modified: string[]; added: string[]; deleted: string[]; clean: boolean }>("GitStatus"),

  // Index
  rebuildIndex: () => call<void>("RebuildIndex"),
  stats: () => call<{ objects: number; relations: number; artifacts: number }>("Stats"),

  // Project
  projectInfo: () => call<ProjectInfo>("ProjectInfo"),
  getNamingInfo: () => call<{ project_code: string; standard: string; pattern: string }>("GetNamingInfo"),
  getNextSequence: (classType: string) => call<number>("GetNextSequence", classType),

  // Document source
  getObjectDocument: (objectId: string) => call<ObjectDocumentDTO>("GetObjectDocument", objectId),
  updateObjectDocument: (req: UpdateObjectDocumentRequest) => call<void>("UpdateObjectDocument", req),
  validateObject: (objectId: string) => call<Diagnostic[]>("ValidateObject", objectId),
  getObjectHistory: (objectId: string) => call<HistoryEntry[]>("GetObjectHistory", objectId),

  // Attachments
  attachArtifact: (objectId: string, localPath: string, kind: string, role: string) =>
    call<void>("AttachArtifact", objectId, localPath, kind, role),

  // Where-Used
  whereUsed: (objectId: string) => call<ObjectDTO[]>("WhereUsed", objectId),

  // BOM Export
  exportBOM: (rootId: string, format: string) => call<string>("ExportBOM", rootId, format),

  // Revision Bump
  bumpRevision: (objectId: string) => call<string>("BumpRevision", objectId),

  // Relations
  addRelation: (fromId: string, toId: string, relType: string, quantity?: number, unit?: string) =>
    call<void>("AddRelation", fromId, toId, relType, quantity, unit),

  // Recovery
  recoverTransactions: () => call<{ id: string; work_dir: string }[]>("RecoverTransactions"),
  rollbackTransaction: (id: string) => call<void>("RollbackTransaction", id),

  // Save/Update
  updateObject: (id: string, title: string, metadata?: Record<string, unknown>) =>
    call<void>("UpdateObject", id, title, metadata || {}),

  // Duplicate
  duplicateObject: (sourceId: string) => call<string>("DuplicateObject", sourceId),

  // Release Package
  createReleasePackage: (rootId: string, title: string) =>
    call<string>("CreateReleasePackage", rootId, title),
};

export default api;
