import { useState, useEffect } from "react";
import type { ObjectDTO, TransitionDef } from "../types";
import { CLASS_LABELS, STATE_LABELS, STATE_COLORS } from "../types";
import { stateBadgeClasses, stateDotClasses } from "../utils";
import { FileText, Tag, Layers, ArrowRight, Link2, Loader2 } from "lucide-react";
import api from "../bindings/api";

interface ObjectEditorProps {
  object: ObjectDTO;
  onTransition: (transition: string) => void;
  onViewBOM: () => void;
}

export default function ObjectEditor({
  object,
  onTransition,
  onViewBOM,
}: ObjectEditorProps) {
  const [transitions, setTransitions] = useState<TransitionDef[]>([]);
  const [activeTab, setActiveTab] = useState<"info" | "metadata" | "lifecycle" | "whereused">(
    "info"
  );
  const [whereUsed, setWhereUsed] = useState<ObjectDTO[]>([]);
  const [whereUsedLoading, setWhereUsedLoading] = useState(false);

  useEffect(() => {
    api
      .availableTransitions(object.id)
      .then(setTransitions)
      .catch(() => setTransitions([]));
  }, [object.id, object.state]);

  const stateColor = STATE_COLORS[object.state] || "gray";
  const stateLabel = STATE_LABELS[object.state] || object.state;
  const classLabel = CLASS_LABELS[object.class] || object.class;

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-mono bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-gray-500">
                {object.id}
              </span>
              <span className="text-xs font-medium px-2 py-0.5 rounded bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400">
                {classLabel}
              </span>
            </div>
            <h1 className="text-xl font-bold">{object.title}</h1>
          </div>
          <div className="flex items-center gap-2">
            <span
              className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium ${stateBadgeClasses[stateColor] || stateBadgeClasses.gray}`}
            >
              <span
                className={`w-2 h-2 rounded-full ${stateDotClasses[stateColor] || stateDotClasses.gray}`}
              />
              {stateLabel}
            </span>
          </div>
        </div>

        {/* Version info */}
        <div className="flex gap-4 mt-2 text-xs text-gray-400">
          <span>Version: {object.version}</span>
          <span>Revision: {object.revision}</span>
          {object.sequence && <span>Sequence: {object.sequence}</span>}
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-gray-200 dark:border-gray-700 px-6">
        {(["info", "metadata", "lifecycle", "whereused"] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              activeTab === tab
                ? "border-blue-500 text-blue-600 dark:text-blue-400"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            <span className="capitalize">{tab}</span>
            {tab === "lifecycle" && transitions.length > 0 && (
              <span className="ml-1.5 px-1.5 py-0.5 text-xs bg-gray-100 dark:bg-gray-800 rounded-full">
                {transitions.length}
              </span>
            )}
          </button>
        ))}

        {/* BOM button for assemblies */}
        {object.class === "asm" && (
          <button
            onClick={onViewBOM}
            className="px-4 py-2 text-sm font-medium text-purple-600 dark:text-purple-400 hover:text-purple-700 border-b-2 border-transparent hover:border-purple-500"
          >
            BOM
          </button>
        )}
      </div>

      {/* Tab content */}
      <div className="flex-1 overflow-y-auto p-6">
        {activeTab === "info" && (
          <InfoTab object={object} />
        )}

        {activeTab === "metadata" && (
          <MetadataTab object={object} />
        )}

        {activeTab === "lifecycle" && (
          <LifecycleTab
            object={object}
            transitions={transitions}
            onTransition={onTransition}
          />
        )}

        {activeTab === "whereused" && (
          <WhereUsedTab objectId={object.id} />
        )}
      </div>
    </div>
  );
}

function InfoTab({ object }: { object: ObjectDTO }) {
  return (
    <div className="max-w-2xl">
      <div className="prose dark:prose-invert">
        <h2>{object.title}</h2>
        <p>
          This is engineering object <code>{object.id}</code> of class{" "}
          <strong>{CLASS_LABELS[object.class] || object.class}</strong>.
        </p>
        <p>
          Current state: <strong>{STATE_LABELS[object.state]}</strong>.
        </p>

        <h3>Quick Actions</h3>
        <ul>
          <li>View lifecycle transitions in the Lifecycle tab</li>
          <li>Edit metadata properties in the Metadata tab</li>
          {object.class === "asm" && (
            <li>View Bill of Materials in the BOM tab</li>
          )}
          <li>
            Full Markdown+YAML editing will be available with Vditor integration
          </li>
        </ul>

        <h3>Object Path</h3>
        <pre className="bg-gray-100 dark:bg-gray-800 p-3 rounded text-sm">
{`objects/${object.id}/
├── ${object.id}.md
├── history.jsonl
└── files/`}
        </pre>
      </div>
    </div>
  );
}

function MetadataTab({ object }: { object: ObjectDTO }) {
  if (!object.metadata || Object.keys(object.metadata).length === 0) {
    return (
      <div className="text-center text-gray-400 py-12">
        <Layers size={48} className="mx-auto mb-4 opacity-50" />
        <p>No metadata fields defined</p>
        <p className="text-sm mt-1">
          Add metadata like unit, material, make_buy in the YAML frontmatter
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-xl">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700">
            <th className="text-left py-2 px-3 font-medium text-gray-500">
              Key
            </th>
            <th className="text-left py-2 px-3 font-medium text-gray-500">
              Value
            </th>
          </tr>
        </thead>
        <tbody>
          {Object.entries(object.metadata).map(([key, value]) => (
            <tr
              key={key}
              className="border-b border-gray-100 dark:border-gray-800"
            >
              <td className="py-2 px-3 font-mono text-xs text-blue-600 dark:text-blue-400">
                {key}
              </td>
              <td className="py-2 px-3">
                {typeof value === "object"
                  ? JSON.stringify(value)
                  : String(value)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
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
      <div className="text-center text-gray-400 py-12">
        <ArrowRight size={48} className="mx-auto mb-4 opacity-50" />
        <p>No transitions available from state "{object.state}"</p>
        <p className="text-sm mt-1">
          This object is in a terminal state or the lifecycle is not configured
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-lg">
      <p className="text-sm text-gray-500 mb-4">
        Available transitions from{" "}
        <strong>{STATE_LABELS[object.state] || object.state}</strong>:
      </p>
      <div className="space-y-2">
        {transitions.map((tr) => (
          <div
            key={tr.name}
            className="flex items-center justify-between p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-700 transition-colors"
          >
            <div>
              <div className="font-medium text-sm">{tr.name}</div>
              <div className="text-xs text-gray-500 mt-0.5">
                → {tr.to}
                {tr.guards.length > 0 && (
                  <span className="ml-2 text-amber-500">
                    guards: {tr.guards.join(", ")}
                  </span>
                )}
                {tr.effects.length > 0 && (
                  <span className="ml-2 text-blue-500">
                    effects: {tr.effects.join(", ")}
                  </span>
                )}
              </div>
            </div>
            <button
              onClick={() => onTransition(tr.name)}
              className="px-3 py-1.5 text-xs font-medium bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            >
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
    api.whereUsed(objectId)
      .then(setObjects)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [objectId]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 size={24} className="animate-spin text-gray-400" />
      </div>
    );
  }

  if (error) {
    return <div className="text-red-500 text-sm py-4">{error}</div>;
  }

  if (objects.length === 0) {
    return (
      <div className="text-center text-gray-400 py-12">
        <Link2 size={48} className="mx-auto mb-4 opacity-50" />
        <p>Not used in any other object</p>
        <p className="text-sm mt-1">No incoming relations found</p>
      </div>
    );
  }

  return (
    <div>
      <p className="text-sm text-gray-500 mb-4">
        Used in {objects.length} object{objects.length !== 1 ? "s" : ""}:
      </p>
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700">
            <th className="text-left py-2 px-3 font-medium text-gray-500">ID</th>
            <th className="text-left py-2 px-3 font-medium text-gray-500">Title</th>
            <th className="text-left py-2 px-3 font-medium text-gray-500">Class</th>
            <th className="text-left py-2 px-3 font-medium text-gray-500">State</th>
          </tr>
        </thead>
        <tbody>
          {objects.map((obj) => (
            <tr key={obj.id} className="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer">
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
