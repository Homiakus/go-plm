import { useState, useEffect } from "react";
import type {
  Diagnostic,
  HistoryEntry,
  ObjectDTO,
  ObjectDocumentDTO,
  TransitionDef,
} from "../types";
import { CLASS_LABELS, STATE_LABELS, STATE_COLORS } from "../types";
import { stateBadgeClasses, stateDotClasses } from "../utils";
import {
  AlertTriangle,
  ArrowRight,
  FileText,
  History,
  Layers,
  Link2,
  Loader2,
  Paperclip,
  Save,
  ShieldCheck,
} from "lucide-react";
import api from "../bindings/api";
import MarkdownEditor from "./MarkdownEditor";

type ObjectTab =
  | "document"
  | "metadata"
  | "relations"
  | "files"
  | "validation"
  | "history"
  | "lifecycle"
  | "whereused";

interface ObjectEditorProps {
  object: ObjectDTO;
  onTransition: (transition: string) => void;
  onViewBOM: () => void;
  onObjectUpdated?: () => void;
}

export default function ObjectEditor({
  object,
  onTransition,
  onViewBOM,
  onObjectUpdated,
}: ObjectEditorProps) {
  const [transitions, setTransitions] = useState<TransitionDef[]>([]);
  const [activeTab, setActiveTab] = useState<ObjectTab>("document");

  useEffect(() => {
    api
      .availableTransitions(object.id)
      .then(setTransitions)
      .catch(() => setTransitions([]));
  }, [object.id, object.state]);

  const stateColor = STATE_COLORS[object.state] || "gray";
  const stateLabel = STATE_LABELS[object.state] || object.state;
  const classLabel = CLASS_LABELS[object.class] || object.class;
  const tabs: ObjectTab[] = [
    "document",
    "metadata",
    "relations",
    "files",
    "validation",
    "history",
    "lifecycle",
    "whereused",
  ];

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <div className="border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-medium px-2 py-0.5 rounded bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400">
                {classLabel}
              </span>
              <code className="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-gray-500">
                {object.id}
              </code>
            </div>
            <h1 className="text-xl font-bold truncate">{object.title}</h1>
            <div className="flex gap-4 mt-2 text-xs text-gray-400">
              <span>Version: {object.version}</span>
              <span>Revision: {object.revision}</span>
              {object.sequence && <span>Sequence: {object.sequence}</span>}
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <span
              className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium ${stateBadgeClasses[stateColor] || stateBadgeClasses.gray}`}
            >
              <span
                className={`w-2 h-2 rounded-full ${stateDotClasses[stateColor] || stateDotClasses.gray}`}
              />
              {stateLabel}
            </span>
            {object.class === "asm" && (
              <button
                onClick={onViewBOM}
                className="px-3 py-1.5 text-xs font-medium bg-purple-600 hover:bg-purple-700 text-white rounded-lg"
              >
                BOM
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="flex border-b border-gray-200 dark:border-gray-700 px-6 overflow-x-auto">
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-3 py-2 text-sm font-medium border-b-2 whitespace-nowrap transition-colors ${
              activeTab === tab
                ? "border-blue-500 text-blue-600 dark:text-blue-400"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            {tabLabel(tab)}
            {tab === "lifecycle" && transitions.length > 0 && (
              <span className="ml-1.5 px-1.5 py-0.5 text-xs bg-gray-100 dark:bg-gray-800 rounded-full">
                {transitions.length}
              </span>
            )}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-hidden">
        {activeTab === "document" && (
          <DocumentTab object={object} onObjectUpdated={onObjectUpdated} />
        )}
        {activeTab === "metadata" && (
          <MetadataTab object={object} onObjectUpdated={onObjectUpdated} />
        )}
        {activeTab === "relations" && (
          <RelationsTab object={object} onObjectUpdated={onObjectUpdated} />
        )}
        {activeTab === "files" && (
          <FilesTab object={object} onObjectUpdated={onObjectUpdated} />
        )}
        {activeTab === "validation" && <ValidationTab objectId={object.id} />}
        {activeTab === "history" && <HistoryTab objectId={object.id} />}
        {activeTab === "lifecycle" && (
          <LifecycleTab object={object} transitions={transitions} onTransition={onTransition} />
        )}
        {activeTab === "whereused" && <WhereUsedTab objectId={object.id} />}
      </div>
    </div>
  );
}

function tabLabel(tab: ObjectTab) {
  return {
    document: "Document",
    metadata: "Metadata",
    relations: "Relations",
    files: "Files",
    validation: "Validation",
    history: "History",
    lifecycle: "Lifecycle",
    whereused: "Where Used",
  }[tab];
}

function DocumentTab({
  object,
  onObjectUpdated,
}: {
  object: ObjectDTO;
  onObjectUpdated?: () => void;
}) {
  const [doc, setDoc] = useState<ObjectDocumentDTO | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setDoc(null);
    setError(null);
    api
      .getObjectDocument(object.id)
      .then(setDoc)
      .catch((e) => setError(e.message));
  }, [object.id]);

  if (error) return <PanelError message={error} />;
  if (!doc) return <PanelLoading label="Loading document..." />;

  return (
    <MarkdownEditor
      object={object}
      initialFrontmatter={doc.frontmatter}
      initialBody={doc.body}
      onSave={async (frontmatter, body) => {
        await api.updateObjectDocument({
          object_id: object.id,
          frontmatter,
          body,
        });
        onObjectUpdated?.();
      }}
    />
  );
}

function MetadataTab({
  object,
  onObjectUpdated,
}: {
  object: ObjectDTO;
  onObjectUpdated?: () => void;
}) {
  const [title, setTitle] = useState(object.title);
  const [metadataText, setMetadataText] = useState(
    JSON.stringify(object.metadata || {}, null, 2)
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setTitle(object.title);
    setMetadataText(JSON.stringify(object.metadata || {}, null, 2));
    setError(null);
  }, [object.id, object.title, object.metadata]);

  const save = async () => {
    setSaving(true);
    setError(null);
    try {
      const metadata = JSON.parse(metadataText || "{}");
      await api.updateObject(object.id, title, metadata);
      onObjectUpdated?.();
    } catch (e: any) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  const applyDefaults = () => {
    try {
      const current = JSON.parse(metadataText || "{}");
      setMetadataText(JSON.stringify({ ...defaultMetadataForClass(object.class), ...current }, null, 2));
      setError(null);
    } catch (e: any) {
      setError(e.message);
    }
  };

  return (
    <div className="h-full overflow-y-auto p-6 max-w-3xl">
      <div className="space-y-4">
        <div>
          <label className="text-sm font-medium text-gray-500">Title</label>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full mt-1 px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900"
          />
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">Metadata JSON</label>
          <textarea
            value={metadataText}
            onChange={(e) => setMetadataText(e.target.value)}
            className="w-full h-64 mt-1 p-3 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 font-mono text-sm"
            spellCheck={false}
          />
        </div>
        {error && <div className="text-sm text-red-500">{error}</div>}
        <div className="flex gap-3">
          <button
            onClick={applyDefaults}
            className="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg text-sm font-medium"
          >
            Apply Class Defaults
          </button>
          <button
            onClick={save}
            disabled={saving || !title.trim()}
            className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white rounded-lg text-sm font-medium"
          >
            <Save size={16} />
            {saving ? "Saving..." : "Save Metadata"}
          </button>
        </div>
      </div>
    </div>
  );
}

function RelationsTab({
  object,
  onObjectUpdated,
}: {
  object: ObjectDTO;
  onObjectUpdated?: () => void;
}) {
  const [target, setTarget] = useState("");
  const [type, setType] = useState("contains");
  const [quantity, setQuantity] = useState("1");
  const [unit, setUnit] = useState("pcs");
  const [error, setError] = useState<string | null>(null);

  const add = async () => {
    setError(null);
    try {
      const qty = quantity.trim() ? Number(quantity) : undefined;
      await api.addRelation(object.id, target.trim(), type.trim(), qty, unit.trim());
      setTarget("");
      onObjectUpdated?.();
    } catch (e: any) {
      setError(e.message);
    }
  };

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="max-w-4xl space-y-6">
        <SectionTitle icon={<Link2 size={18} />} title="Relations" />
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700 text-left text-gray-500">
              <th className="py-2">Type</th>
              <th className="py-2">Target</th>
              <th className="py-2">Qty</th>
              <th className="py-2">Unit</th>
            </tr>
          </thead>
          <tbody>
            {(object.relations || []).map((rel, i) => (
              <tr key={`${rel.type}-${rel.to}-${i}`} className="border-b border-gray-100 dark:border-gray-800">
                <td className="py-2">{rel.type}</td>
                <td className="py-2 font-mono text-xs text-blue-600">{rel.to}</td>
                <td className="py-2">{rel.quantity ?? ""}</td>
                <td className="py-2">{rel.unit || ""}</td>
              </tr>
            ))}
            {(object.relations || []).length === 0 && (
              <tr>
                <td className="py-8 text-center text-gray-400" colSpan={4}>
                  No relations
                </td>
              </tr>
            )}
          </tbody>
        </table>

        <div className="rounded-lg border border-gray-200 dark:border-gray-700 p-4 space-y-3">
          <div className="text-sm font-semibold">Add Relation</div>
          <div className="grid grid-cols-1 md:grid-cols-5 gap-2">
            <input value={target} onChange={(e) => setTarget(e.target.value)} placeholder="target object id" className="md:col-span-2 px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm" />
            <input value={type} onChange={(e) => setType(e.target.value)} placeholder="type" className="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm" />
            <input value={quantity} onChange={(e) => setQuantity(e.target.value)} placeholder="qty" className="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm" />
            <input value={unit} onChange={(e) => setUnit(e.target.value)} placeholder="unit" className="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm" />
          </div>
          {error && <div className="text-sm text-red-500">{error}</div>}
          <button onClick={add} disabled={!target.trim() || !type.trim()} className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white rounded-lg text-sm font-medium">
            Add Relation
          </button>
        </div>
      </div>
    </div>
  );
}

function FilesTab({
  object,
  onObjectUpdated,
}: {
  object: ObjectDTO;
  onObjectUpdated?: () => void;
}) {
  const [path, setPath] = useState("");
  const [kind, setKind] = useState("auto");
  const [role, setRole] = useState("auto");
  const [error, setError] = useState<string | null>(null);

  const attach = async () => {
    setError(null);
    try {
      await api.attachArtifact(object.id, path.trim(), kind.trim(), role.trim());
      setPath("");
      onObjectUpdated?.();
    } catch (e: any) {
      setError(e.message);
    }
  };

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="max-w-4xl space-y-6">
        <SectionTitle icon={<Paperclip size={18} />} title="Files" />
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700 text-left text-gray-500">
              <th className="py-2">Role</th>
              <th className="py-2">Kind</th>
              <th className="py-2">Path</th>
              <th className="py-2">Status</th>
              <th className="py-2">Checksum</th>
            </tr>
          </thead>
          <tbody>
            {(object.artifacts || []).map((a) => (
              <tr key={a.id} className="border-b border-gray-100 dark:border-gray-800">
                <td className="py-2">{a.role}</td>
                <td className="py-2">{a.kind}</td>
                <td className="py-2 font-mono text-xs text-blue-600">{a.path}</td>
                <td className="py-2">{a.status}</td>
                <td className="py-2 font-mono text-xs text-gray-400">{a.checksum?.slice(0, 12)}</td>
              </tr>
            ))}
            {(object.artifacts || []).length === 0 && (
              <tr>
                <td className="py-8 text-center text-gray-400" colSpan={5}>
                  No files attached
                </td>
              </tr>
            )}
          </tbody>
        </table>

        <div className="rounded-lg border border-gray-200 dark:border-gray-700 p-4 space-y-3">
          <div className="text-sm font-semibold">Attach File</div>
          <input value={path} onChange={(e) => setPath(e.target.value)} placeholder="absolute local file path" className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm" />
          <div className="grid grid-cols-2 gap-2">
            <input value={kind} onChange={(e) => setKind(e.target.value)} placeholder="kind" className="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm" />
            <input value={role} onChange={(e) => setRole(e.target.value)} placeholder="role" className="px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 text-sm" />
          </div>
          <div className="text-xs text-gray-400">
            Auto mode infers kind/role from extension and object class, then stores as <code>files/.../{object.id}.ext</code>.
          </div>
          {error && <div className="text-sm text-red-500">{error}</div>}
          <button onClick={attach} disabled={!path.trim()} className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white rounded-lg text-sm font-medium">
            Attach File
          </button>
        </div>
      </div>
    </div>
  );
}

function defaultMetadataForClass(classType: string): Record<string, unknown> {
  const common: Record<string, unknown> = {
    lifecycle_stage: "development",
    owner: "local-user",
  };
  const byClass: Record<string, Record<string, unknown>> = {
    prt: { unit: "pcs", make_buy: "make", material: "", thickness_mm: null, manufacturing_method: null },
    asm: { unit: "pcs", assembly_type: "mechanical" },
    drw: { drawing_for: "", format: "A3", projection: "first_angle" },
    doc: { doc_type: "engineering" },
    std: { unit: "pcs", make_buy: "buy", std_class: "", std_id: "", standard: "" },
    mat: { unit: "pcs", make_buy: "buy", material: "", form: "" },
    cut: { source_part: "", process: "laser_cut", material: "", thickness_mm: null },
    bnd: { source_part: "", process: "bending", material: "", thickness_mm: null },
    nc: { source_part: "", process: "machining", machine: null },
    ins: { source_object: "" },
    tpc: { source_object: "" },
    wi: { source_object: "" },
    rel: { release_type: "engineering" },
    cr: { change_type: "engineering_change", priority: "normal" },
  };
  return { ...common, ...(byClass[classType] || {}) };
}

function ValidationTab({ objectId }: { objectId: string }) {
  const [diags, setDiags] = useState<Diagnostic[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const run = async () => {
    setLoading(true);
    setError(null);
    try {
      setDiags(await api.validateObject(objectId));
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    run();
  }, [objectId]);

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="max-w-4xl space-y-4">
        <div className="flex items-center justify-between">
          <SectionTitle icon={<ShieldCheck size={18} />} title="Validation" />
          <button onClick={run} disabled={loading} className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white rounded-lg text-sm">
            {loading ? "Running..." : "Run Validation"}
          </button>
        </div>
        {error && <PanelError message={error} />}
        {!loading && diags.length === 0 && (
          <div className="rounded-lg border border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/20 p-4 text-sm text-green-700 dark:text-green-400">
            Ready. No validation issues found.
          </div>
        )}
        <div className="space-y-2">
          {diags.map((d) => (
            <div key={d.id} className="rounded-lg border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-900/10 p-3">
              <div className="flex items-start gap-2">
                <AlertTriangle size={16} className="text-amber-500 mt-0.5 shrink-0" />
                <div>
                  <div className="text-sm font-medium">{d.message}</div>
                  <div className="mt-1 flex gap-2 text-xs text-gray-500">
                    <code>{d.severity}</code>
                    <code>{d.code}</code>
                    {d.path && <code>{d.path}</code>}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function HistoryTab({ objectId }: { objectId: string }) {
  const [items, setItems] = useState<HistoryEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    api
      .getObjectHistory(objectId)
      .then(setItems)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [objectId]);

  if (loading) return <PanelLoading label="Loading history..." />;
  if (error) return <PanelError message={error} />;

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="max-w-4xl space-y-4">
        <SectionTitle icon={<History size={18} />} title="History" />
        {items.length === 0 && <div className="text-sm text-gray-400">No history entries</div>}
        {items.map((item, i) => (
          <div key={`${item.event_id || i}`} className="rounded-lg border border-gray-200 dark:border-gray-700 p-3">
            <div className="flex items-center justify-between gap-3">
              <div className="text-sm font-medium">{String(item.type || "event")}</div>
              <div className="text-xs text-gray-400">{String(item.time || "")}</div>
            </div>
            <pre className="mt-2 text-xs bg-gray-50 dark:bg-gray-800 rounded p-2 overflow-x-auto">
{JSON.stringify(item.payload || item, null, 2)}
            </pre>
          </div>
        ))}
      </div>
    </div>
  );
}

function LifecycleTab({
  object,
  transitions,
  onTransition,
}: {
  object: ObjectDTO;
  transitions: TransitionDef[];
  onTransition: (t: string) => void;
}) {
  if (transitions.length === 0) {
    return (
      <EmptyPanel
        icon={<ArrowRight size={48} />}
        title={`No transitions available from "${object.state}"`}
        detail="This object is in a terminal state or lifecycle is not configured."
      />
    );
  }

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="max-w-2xl space-y-3">
        {transitions.map((tr) => (
          <div key={tr.name} className="flex items-center justify-between gap-4 p-3 rounded-lg border border-gray-200 dark:border-gray-700">
            <div>
              <div className="font-medium text-sm">{tr.name}</div>
              <div className="text-xs text-gray-500 mt-0.5">
                to {tr.to}
                {tr.guards.length > 0 && <span className="ml-2 text-amber-500">guards: {tr.guards.join(", ")}</span>}
                {tr.effects.length > 0 && <span className="ml-2 text-blue-500">effects: {tr.effects.join(", ")}</span>}
              </div>
            </div>
            <button onClick={() => onTransition(tr.name)} className="px-3 py-1.5 text-xs font-medium bg-blue-600 hover:bg-blue-700 text-white rounded-lg">
              Execute
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}

function WhereUsedTab({ objectId }: { objectId: string }) {
  const [objects, setObjects] = useState<ObjectDTO[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    api
      .whereUsed(objectId)
      .then(setObjects)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [objectId]);

  if (loading) return <PanelLoading label="Loading where-used..." />;
  if (error) return <PanelError message={error} />;
  if (objects.length === 0) {
    return (
      <EmptyPanel
        icon={<Link2 size={48} />}
        title="Not used in any other object"
        detail="No incoming relations found."
      />
    );
  }

  return (
    <div className="h-full overflow-y-auto p-6">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700 text-left text-gray-500">
            <th className="py-2 px-3">ID</th>
            <th className="py-2 px-3">Title</th>
            <th className="py-2 px-3">Class</th>
            <th className="py-2 px-3">State</th>
          </tr>
        </thead>
        <tbody>
          {objects.map((obj) => (
            <tr key={obj.id} className="border-b border-gray-100 dark:border-gray-800">
              <td className="py-2 px-3 font-mono text-xs text-blue-600">{obj.id}</td>
              <td className="py-2 px-3">{obj.title}</td>
              <td className="py-2 px-3 text-xs">{obj.class}</td>
              <td className="py-2 px-3 text-xs">{obj.state}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function SectionTitle({ icon, title }: { icon: React.ReactNode; title: string }) {
  return (
    <h2 className="text-lg font-bold flex items-center gap-2">
      <span className="text-blue-500">{icon}</span>
      {title}
    </h2>
  );
}

function PanelLoading({ label }: { label: string }) {
  return (
    <div className="h-full flex items-center justify-center text-gray-400">
      <Loader2 size={22} className="animate-spin mr-2" />
      <span className="text-sm">{label}</span>
    </div>
  );
}

function PanelError({ message }: { message: string }) {
  return (
    <div className="m-6 rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-4 text-sm text-red-700 dark:text-red-400">
      {message}
    </div>
  );
}

function EmptyPanel({
  icon,
  title,
  detail,
}: {
  icon: React.ReactNode;
  title: string;
  detail: string;
}) {
  return (
    <div className="h-full flex items-center justify-center text-gray-400">
      <div className="text-center">
        <div className="mx-auto mb-4 opacity-50 flex justify-center">{icon}</div>
        <p>{title}</p>
        <p className="text-sm mt-1">{detail}</p>
      </div>
    </div>
  );
}
