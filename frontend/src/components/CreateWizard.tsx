// 3-step Create Object Wizard: Identity → Properties → Preview & Create
import { useState } from "react";
import { ChevronRight, ChevronLeft, Check, FileText, Layers, Paperclip } from "lucide-react";
import { CLASS_LABELS } from "../types";

interface CreateWizardProps {
  onCreate: (classType: string, title: string, metadata?: Record<string, unknown>, parentId?: string) => void;
  onCancel: () => void;
  suggestedParentId?: string;
}

const STEPS = ["Identity", "Properties", "Preview"] as const;

const ALL_CLASSES = ["prt", "asm", "drw", "doc", "std", "mat", "cut", "bnd", "nc", "ins", "tpc", "wi", "bom", "rel", "cr"];

const QUICK_CLASSES = ["prt", "asm", "drw", "doc", "std"];

export default function CreateWizard({ onCreate, onCancel, suggestedParentId }: CreateWizardProps) {
  const [step, setStep] = useState(0);
  const [classType, setClassType] = useState("prt");
  const [title, setTitle] = useState("");
  const [parentId, setParentId] = useState(suggestedParentId || "");
  const [metadata, setMetadata] = useState<[string, string][]>([["unit", "pcs"]]);

  const addMetaField = () => setMetadata([...metadata, ["", ""]]);
  const updateMeta = (i: number, key: string, value: string) => {
    const m = [...metadata];
    m[i] = [key, value];
    setMetadata(m);
  };
  const removeMeta = (i: number) => setMetadata(metadata.filter((_, idx) => idx !== i));

  const metaObj = (): Record<string, unknown> => {
    const obj: Record<string, unknown> = {};
    metadata.forEach(([k, v]) => { if (k.trim()) obj[k.trim()] = v; });
    return obj;
  };

  const canNext = () => {
    if (step === 0) return classType.length > 0;
    if (step === 1) return title.trim().length > 0;
    return true;
  };

  const handleCreate = () => {
    onCreate(classType, title.trim(), metaObj(), parentId || undefined);
  };

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col">
        {/* Header with steps */}
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-bold">Create New Object</h2>
          <div className="flex items-center gap-2 mt-3">
            {STEPS.map((s, i) => (
              <div key={s} className="flex items-center gap-2">
                <div className={`flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium ${
                  i <= step
                    ? "bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400"
                    : "bg-gray-100 dark:bg-gray-800 text-gray-400"
                }`}>
                  <span className={`w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold ${
                    i < step
                      ? "bg-blue-500 text-white"
                      : i === step
                      ? "bg-blue-500 text-white"
                      : "bg-gray-300 dark:bg-gray-600 text-gray-500"
                  }`}>
                    {i < step ? <Check size={12} /> : i + 1}
                  </span>
                  {s}
                </div>
                {i < STEPS.length - 1 && <ChevronRight size={16} className="text-gray-300" />}
              </div>
            ))}
          </div>
        </div>

        {/* Step content */}
        <div className="flex-1 overflow-y-auto p-6">
          {step === 0 && (
            <div className="space-y-4">
              <p className="text-sm text-gray-500">Choose the type of engineering object to create.</p>

              {/* Quick classes */}
              <div className="grid grid-cols-5 gap-2">
                {QUICK_CLASSES.map((c) => (
                  <button
                    key={c}
                    onClick={() => setClassType(c)}
                    className={`p-3 rounded-xl border-2 text-center transition-all ${
                      classType === c
                        ? "border-blue-500 bg-blue-50 dark:bg-blue-900/30 shadow-sm"
                        : "border-gray-200 dark:border-gray-700 hover:border-gray-300"
                    }`}
                  >
                    <div className="text-lg mb-1">
                      {c === "prt" ? "🖼" : c === "asm" ? "📦" : c === "drw" ? "📐" : c === "doc" ? "📄" : "🔩"}
                    </div>
                    <div className="text-xs font-medium">{CLASS_LABELS[c] || c}</div>
                    <div className="text-[10px] text-gray-400">{c}</div>
                  </button>
                ))}
              </div>

              {/* All classes */}
              <details className="mt-2">
                <summary className="text-sm text-gray-400 cursor-pointer hover:text-gray-600">
                  All 15 classes...
                </summary>
                <div className="grid grid-cols-4 gap-2 mt-2">
                  {ALL_CLASSES.filter(c => !QUICK_CLASSES.includes(c)).map((c) => (
                    <button
                      key={c}
                      onClick={() => setClassType(c)}
                      className={`p-2 rounded-lg border text-center text-xs transition-all ${
                        classType === c
                          ? "border-blue-500 bg-blue-50 dark:bg-blue-900/30"
                          : "border-gray-200 dark:border-gray-700 hover:border-gray-300"
                      }`}
                    >
                      {CLASS_LABELS[c] || c}
                    </button>
                  ))}
                </div>
              </details>

              {/* Parent object */}
              <div className="mt-4">
                <label className="text-sm font-medium text-gray-500">Parent Assembly (optional)</label>
                <input
                  type="text"
                  value={parentId}
                  onChange={(e) => setParentId(e.target.value)}
                  placeholder="e.g., demo-asm-0100-v1.0"
                  className="w-full mt-1 px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-gray-400 mt-1">A &apos;contains&apos; relation will be created automatically</p>
              </div>
            </div>
          )}

          {step === 1 && (
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-gray-500">Title *</label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Object title (e.g., Bracket Motor)"
                  className="w-full mt-1 px-4 py-2.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  autoFocus
                />
              </div>

              {/* Metadata fields */}
              <div>
                <div className="flex items-center justify-between">
                  <label className="text-sm font-medium text-gray-500">Metadata</label>
                  <button onClick={addMetaField} className="text-xs text-blue-500 hover:text-blue-700">
                    + Add field
                  </button>
                </div>
                <div className="space-y-2 mt-2">
                  {metadata.map(([k, v], i) => (
                    <div key={i} className="flex gap-2">
                      <input
                        type="text"
                        value={k}
                        onChange={(e) => updateMeta(i, e.target.value, v)}
                        placeholder="key"
                        className="flex-1 px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      <input
                        type="text"
                        value={v}
                        onChange={(e) => updateMeta(i, k, e.target.value)}
                        placeholder="value"
                        className="flex-[2] px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                      <button
                        onClick={() => removeMeta(i)}
                        className="px-2 text-gray-400 hover:text-red-500 text-lg"
                      >
                        ×
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <p className="text-sm text-gray-500">Review before creating:</p>

              <div className="bg-gray-50 dark:bg-gray-800 rounded-xl p-4 space-y-3">
                <PreviewRow label="Class" value={CLASS_LABELS[classType] || classType} />
                <PreviewRow label="ID" value="<auto-generated>" mono />
                <PreviewRow label="Title" value={title || "(not set)"} />
                {parentId && <PreviewRow label="Parent" value={parentId} mono />}
                <PreviewRow label="State" value="draft" />

                <div className="pt-2 border-t border-gray-200 dark:border-gray-700">
                  <div className="text-xs text-gray-500 mb-2">Metadata:</div>
                  {metadata.filter(([k]) => k.trim()).map(([k, v]) => (
                    <PreviewRow key={k} label={k} value={v || "(empty)"} />
                  ))}
                  {metadata.filter(([k]) => k.trim()).length === 0 && (
                    <div className="text-xs text-gray-400 italic">No metadata fields</div>
                  )}
                </div>
              </div>

              <div className="bg-blue-50 dark:bg-blue-900/20 rounded-xl p-4 text-sm text-blue-700 dark:text-blue-400">
                <strong>What will be created:</strong>
                <ul className="list-disc ml-4 mt-1 text-xs space-y-0.5">
                  <li>Directory: <code>objects/&lt;id&gt;/</code></li>
                  <li>File: <code>&lt;id&gt;.md</code> with YAML frontmatter</li>
                  <li>File: <code>history.jsonl</code> with created event</li>
                  <li>Index updated in SQLite</li>
                  {parentId && <li>Auto relation: <code>contains</code> from parent</li>}
                </ul>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-6 py-4 border-t border-gray-200 dark:border-gray-700">
          <button onClick={onCancel} className="px-4 py-2 text-sm text-gray-500 hover:text-gray-700">
            Cancel
          </button>
          <div className="flex gap-3">
            {step > 0 && (
              <button
                onClick={() => setStep(step - 1)}
                className="flex items-center gap-1 px-4 py-2 text-sm bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 rounded-lg transition-colors"
              >
                <ChevronLeft size={16} /> Back
              </button>
            )}
            {step < 2 ? (
              <button
                onClick={() => setStep(step + 1)}
                disabled={!canNext()}
                className="flex items-center gap-1 px-6 py-2 text-sm font-medium bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white rounded-lg transition-colors"
              >
                Next <ChevronRight size={16} />
              </button>
            ) : (
              <button
                onClick={handleCreate}
                disabled={!title.trim()}
                className="flex items-center gap-1 px-6 py-2 text-sm font-medium bg-green-600 hover:bg-green-700 disabled:bg-gray-300 text-white rounded-lg transition-colors"
              >
                <Check size={16} /> Create Object
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function PreviewRow({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex gap-3 text-sm">
      <span className="text-gray-400 w-20 shrink-0">{label}:</span>
      <span className={mono ? "font-mono text-xs text-blue-600" : ""}>{value}</span>
    </div>
  );
}
