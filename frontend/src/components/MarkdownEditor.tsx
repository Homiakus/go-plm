// Vditor-based Markdown editor with YAML frontmatter editing and autosave.
import { useEffect, useRef, useState, useCallback } from "react";
import { Save, Eye, Code, RefreshCw } from "lucide-react";
import type { ObjectDTO } from "../types";

interface MarkdownEditorProps {
  object: ObjectDTO;
  onSave: (frontmatter: string, body: string) => Promise<void>;
}

export default function MarkdownEditor({ object, onSave }: MarkdownEditorProps) {
  const [mode, setMode] = useState<"edit" | "preview" | "split">("edit");
  const [frontmatter, setFrontmatter] = useState("");
  const [body, setBody] = useState("# " + object.title + "\n\n");
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const editorRef = useRef<HTMLTextAreaElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  // Generate frontmatter from object
  useEffect(() => {
    const fm = `id: ${object.id}
project: ${object.project}
class: ${object.class}
sequence: ${object.sequence || ""}
version: ${object.version}
revision: ${object.revision}
state: ${object.state}
title: ${object.title}
metadata:${object.metadata ? "\n" + Object.entries(object.metadata).map(([k, v]) =>
  typeof v === "object" ? `  ${k}: ${JSON.stringify(v)}` : `  ${k}: ${v}`
).join("\n") : " {}"}
relations: []
artifacts: []`;
    setFrontmatter(fm);
  }, [object]);

  // Autosave with 5-second debounce
  const debouncedSave = useCallback(() => {
    if (saved) return;
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(async () => {
      setSaving(true);
      setError(null);
      try {
        const fullFm = "---\n" + frontmatter + "\n---";
        await onSave(fullFm, body);
        setSaved(true);
      } catch (e: any) {
        setError(e.message);
      } finally {
        setSaving(false);
      }
    }, 5000);
  }, [frontmatter, body, saved, onSave]);

  useEffect(() => {
    if (!saved) debouncedSave();
  }, [frontmatter, body]);

  const handleManualSave = async () => {
    if (timerRef.current) clearTimeout(timerRef.current);
    setSaving(true);
    setError(null);
    try {
      const fullFm = "---\n" + frontmatter + "\n---";
      await onSave(fullFm, body);
      setSaved(true);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  // Keyboard shortcut: Ctrl+S
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "s") {
        e.preventDefault();
        handleManualSave();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [frontmatter, body]);

  const previewHtml = () => {
    let html = body
      .replace(/^### (.+)$/gm, '<h3 class="text-lg font-bold mt-4 mb-2">$1</h3>')
      .replace(/^## (.+)$/gm, '<h2 class="text-xl font-bold mt-6 mb-3">$1</h2>')
      .replace(/^# (.+)$/gm, '<h1 class="text-2xl font-bold mt-8 mb-4">$1</h1>')
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.+?)\*/g, '<em>$1</em>')
      .replace(/`(.+?)`/g, '<code class="bg-gray-100 dark:bg-gray-800 px-1 rounded text-sm">$1</code>')
      .replace(/^- (.+)$/gm, '<li class="ml-4">• $1</li>')
      .replace(/^\[(.+?)\]\((.+?)\)/gm, '<a href="$2" class="text-blue-500 underline">$1</a>')
      .replace(/\n\n/g, '</p><p class="mb-2">')
      .replace(/\n/g, '<br/>');
    return '<p class="mb-2">' + html + '</p>';
  };

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900">
        <div className="flex bg-gray-200 dark:bg-gray-700 rounded-lg p-0.5">
          {(["edit", "preview", "split"] as const).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
                mode === m
                  ? "bg-white dark:bg-gray-600 shadow-sm"
                  : "text-gray-500 hover:text-gray-700"
              }`}
            >
              {m === "edit" && <Code size={14} className="inline mr-1" />}
              {m === "preview" && <Eye size={14} className="inline mr-1" />}
              {m === "split" && <RefreshCw size={14} className="inline mr-1" />}
              {m === "edit" ? "Edit" : m === "preview" ? "Preview" : "Split"}
            </button>
          ))}
        </div>

        <div className="flex-1" />

        {/* Save status */}
        {saving ? (
          <span className="text-xs text-amber-500 flex items-center gap-1">
            <RefreshCw size={12} className="animate-spin" /> Saving...
          </span>
        ) : saved ? (
          <span className="text-xs text-green-500 flex items-center gap-1">
            ✓ Saved
          </span>
        ) : (
          <span className="text-xs text-gray-400 flex items-center gap-1">
            Unsaved changes
          </span>
        )}

        <button
          onClick={handleManualSave}
          disabled={saved || saving}
          className="flex items-center gap-1 px-3 py-1 text-xs font-medium bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white rounded-lg"
        >
          <Save size={14} />
          Save
        </button>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/30 px-4 py-2 text-sm text-red-600 dark:text-red-400">
          {error}
        </div>
      )}

      {/* Editor area */}
      <div className={`flex-1 overflow-hidden ${mode === "split" ? "flex" : ""}`}>
        {/* Frontmatter */}
        <div className={`border-b border-gray-200 dark:border-gray-700 ${mode === "split" ? "w-1/2 border-r" : ""}`}>
          <div className="px-4 py-1.5 bg-gray-100 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <span className="text-xs font-mono font-bold text-gray-500">YAML Frontmatter</span>
          </div>
          <textarea
            ref={editorRef}
            value={frontmatter}
            onChange={(e) => { setFrontmatter(e.target.value); setSaved(false); }}
            className="w-full h-48 p-4 font-mono text-sm bg-white dark:bg-gray-900 resize-none focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
            spellCheck={false}
          />
        </div>

        {/* Markdown body */}
        <div className={`flex-1 ${mode === "split" ? "w-1/2" : ""}`}>
          <div className="px-4 py-1.5 bg-gray-100 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
            <span className="text-xs font-mono font-bold text-gray-500">Markdown Body</span>
          </div>
          {mode === "preview" ? (
            <div
              className="p-6 prose dark:prose-invert max-w-none overflow-y-auto h-full"
              dangerouslySetInnerHTML={{ __html: previewHtml() }}
            />
          ) : (
            <textarea
              value={body}
              onChange={(e) => { setBody(e.target.value); setSaved(false); }}
              className="w-full h-full p-6 text-sm bg-white dark:bg-gray-900 resize-none focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500 font-mono"
              placeholder="# Start writing..."
            />
          )}
          {mode === "split" && (
            <div
              className="p-6 prose dark:prose-invert max-w-none overflow-y-auto border-t border-gray-200 dark:border-gray-700"
              dangerouslySetInnerHTML={{ __html: previewHtml() }}
            />
          )}
        </div>
      </div>
    </div>
  );
}
