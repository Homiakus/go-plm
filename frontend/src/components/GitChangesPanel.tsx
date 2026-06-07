// Git Changes Screen — engineering-friendly view of git status, diff, checkpoint
import { useState, useEffect } from "react";
import { GitBranch, GitCommit, Circle, Plus, Minus, Edit, RefreshCw, Check } from "lucide-react";
import api from "../bindings/api";

interface GitStatus {
  modified: string[];
  added: string[];
  deleted: string[];
  clean: boolean;
}

export default function GitChangesPanel() {
  const [status, setStatus] = useState<GitStatus | null>(null);
  const [message, setMessage] = useState("");
  const [committing, setCommitting] = useState(false);
  const [loading, setLoading] = useState(true);

  const loadStatus = async () => {
    setLoading(true);
    try {
      const s = await api.gitStatus();
      setStatus(s);
    } catch {}
    setLoading(false);
  };

  useEffect(() => { loadStatus(); }, []);

  const handleCheckpoint = async () => {
    if (!message.trim()) return;
    setCommitting(true);
    try {
      await api.createCheckpoint(message.trim());
      setMessage("");
      await loadStatus();
    } catch {}
    setCommitting(false);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <RefreshCw size={20} className="animate-spin text-gray-400" />
      </div>
    );
  }

  if (!status) {
    return <div className="text-center py-12 text-gray-400">Unable to read git status</div>;
  }

  const totalChanges = status.modified.length + status.added.length + status.deleted.length;

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <h2 className="text-lg font-bold flex items-center gap-2">
          <GitBranch size={20} className="text-orange-500" />
          Git Changes
        </h2>
        <p className="text-sm text-gray-500 mt-1">
          Engineering view of project changes — no git commands needed
        </p>
      </div>

      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {/* Clean/Modified banner */}
        <div className={`p-4 rounded-xl border-2 ${status.clean ? "border-green-500 bg-green-50 dark:bg-green-900/20" : "border-amber-500 bg-amber-50 dark:bg-amber-900/20"}`}>
          <div className="flex items-center gap-3">
            {status.clean ? (
              <Check size={24} className="text-green-500" />
            ) : (
              <Circle size={24} className="text-amber-500 fill-amber-500" />
            )}
            <div>
              <h3 className="font-bold text-lg">
                {status.clean ? "Working tree clean" : `${totalChanges} file(s) changed`}
              </h3>
              <p className="text-sm text-gray-500">
                {status.clean ? "No uncommitted changes" : "Create a checkpoint to save these changes"}
              </p>
            </div>
          </div>
        </div>

        {/* File lists */}
        {totalChanges > 0 && (
          <div className="space-y-4">
            {status.added.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold flex items-center gap-2 text-green-600 mb-2">
                  <Plus size={14} /> Added ({status.added.length})
                </h4>
                <div className="space-y-1">
                  {status.added.map((f) => (
                    <div key={f} className="text-sm font-mono text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-900/10 px-3 py-1 rounded">{f}</div>
                  ))}
                </div>
              </div>
            )}
            {status.modified.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold flex items-center gap-2 text-amber-600 mb-2">
                  <Edit size={14} /> Modified ({status.modified.length})
                </h4>
                <div className="space-y-1">
                  {status.modified.map((f) => (
                    <div key={f} className="text-sm font-mono text-amber-700 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/10 px-3 py-1 rounded">{f}</div>
                  ))}
                </div>
              </div>
            )}
            {status.deleted.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold flex items-center gap-2 text-red-600 mb-2">
                  <Minus size={14} /> Deleted ({status.deleted.length})
                </h4>
                <div className="space-y-1">
                  {status.deleted.map((f) => (
                    <div key={f} className="text-sm font-mono text-red-700 dark:text-red-400 bg-red-50 dark:bg-red-900/10 px-3 py-1 rounded">{f}</div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Create checkpoint */}
        {!status.clean && (
          <div className="bg-gray-50 dark:bg-gray-800 rounded-xl p-4">
            <h4 className="text-sm font-semibold flex items-center gap-2 mb-3">
              <GitCommit size={16} className="text-blue-500" />
              Create Checkpoint
            </h4>
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              placeholder="Describe what changed (e.g., 'Added bracket part, updated motor assembly BOM')"
              className="w-full px-3 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={2}
            />
            <div className="flex items-center justify-between mt-2">
              <span className="text-xs text-gray-400">This creates a git commit — a saved point you can return to</span>
              <button
                onClick={handleCheckpoint}
                disabled={!message.trim() || committing}
                className="px-4 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white text-sm font-medium rounded-lg transition-colors flex items-center gap-1"
              >
                <GitCommit size={14} />
                {committing ? "Creating..." : "Create Checkpoint"}
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="px-4 py-2 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <button onClick={loadStatus} className="flex items-center gap-1 text-xs text-gray-400 hover:text-gray-600">
          <RefreshCw size={12} /> Refresh
        </button>
        <span className="text-xs text-gray-400">
          {status.clean ? "Clean" : `${totalChanges} changes`}
        </span>
      </div>
    </div>
  );
}
