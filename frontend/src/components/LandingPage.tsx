import { useState, useEffect } from "react";
import { FolderOpen, PlusCircle, Clock, Settings } from "lucide-react";

interface LandingPageProps {
  onCreateProject: (name: string, title: string) => void;
  onOpenProject: (path: string) => void;
}

interface RecentProject {
  path: string;
  name: string;
  title: string;
  lastOpened: string;
}

export default function LandingPage({ onCreateProject, onOpenProject }: LandingPageProps) {
  const [newName, setNewName] = useState("");
  const [newTitle, setNewTitle] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [openPath, setOpenPath] = useState("");
  const [recent, setRecent] = useState<RecentProject[]>([]);

  useEffect(() => {
    // Load recent projects from localStorage
    try {
      const stored = localStorage.getItem("go-plm-recent");
      if (stored) setRecent(JSON.parse(stored));
    } catch {}
  }, []);

  const handleCreate = () => {
    if (newName.trim()) {
      onCreateProject(newName.trim(), newTitle.trim() || newName.trim());
      setShowCreate(false);
      setNewName("");
      setNewTitle("");
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-blue-50 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center p-8">
      <div className="max-w-lg w-full">
        {/* Logo */}
        <div className="text-center mb-10">
          <div className="text-7xl mb-4">📦</div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            go-plm
          </h1>
          <p className="text-gray-500 mt-2 text-lg">
            Local Git-native PLM/PDM System
          </p>
          <p className="text-gray-400 text-sm mt-1">
            No server. No database. No cloud.
          </p>
        </div>

        {/* Actions */}
        <div className="space-y-4">
          {/* Create project */}
          {showCreate ? (
            <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-6 shadow-sm">
              <h2 className="font-semibold mb-4">Create New Project</h2>
              <input
                type="text"
                placeholder="Project name (e.g., my-project)"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                className="w-full px-4 py-2.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 mb-3 focus:outline-none focus:ring-2 focus:ring-blue-500"
                autoFocus
              />
              <input
                type="text"
                placeholder="Project title (optional)"
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                className="w-full px-4 py-2.5 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-900 mb-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <div className="flex gap-3">
                <button
                  onClick={handleCreate}
                  disabled={!newName.trim()}
                  className="flex-1 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white font-medium rounded-lg transition-colors"
                >
                  Create Project
                </button>
                <button
                  onClick={() => setShowCreate(false)}
                  className="px-4 py-2.5 text-gray-500 hover:text-gray-700 rounded-lg"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <button
              onClick={() => setShowCreate(true)}
              className="w-full flex items-center gap-3 px-6 py-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-xl shadow-sm transition-all hover:shadow-md"
            >
              <PlusCircle size={24} />
              <div className="text-left">
                <div className="text-lg">Create New Project</div>
                <div className="text-sm opacity-80">Start a new PLM workspace</div>
              </div>
            </button>
          )}

          {/* Open project */}
          <div className="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 shadow-sm">
            <div className="flex items-center gap-3">
              <FolderOpen size={20} className="text-gray-400 shrink-0" />
              <input
                type="text"
                placeholder="Path to project (e.g., ~/projects/a320)"
                value={openPath}
                onChange={(e) => setOpenPath(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && onOpenProject(openPath || ".")}
                className="flex-1 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              />
              <button
                onClick={() => onOpenProject(openPath || ".")}
                className="px-4 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg text-sm font-medium transition-colors"
              >
                Open
              </button>
            </div>
          </div>
        </div>

        {/* Recent projects */}
        {recent.length > 0 && (
          <div className="mt-8">
            <h3 className="text-sm font-medium text-gray-500 mb-3 flex items-center gap-2">
              <Clock size={14} /> Recent Projects
            </h3>
            <div className="space-y-2">
              {recent.map((p) => (
                <button
                  key={p.path}
                  onClick={() => onOpenProject(p.path)}
                  className="w-full flex items-center gap-3 px-4 py-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-600 transition-colors text-left"
                >
                  <FolderOpen size={18} className="text-blue-500 shrink-0" />
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-sm truncate">{p.title}</div>
                    <div className="text-xs text-gray-400 truncate">{p.path}</div>
                  </div>
                  <span className="text-xs text-gray-400">{p.lastOpened}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Philosophy */}
        <div className="mt-12 text-center">
          <p className="text-xs text-gray-400">
            Just files. Just Git. Just Markdown.
          </p>
          <div className="flex justify-center gap-6 mt-3 text-xs text-gray-400">
            <span>🔒 Local-first</span>
            <span>📝 Markdown+YAML</span>
            <span>🔄 Git-native</span>
            <span>⚡ FSM-driven</span>
          </div>
        </div>
      </div>
    </div>
  );
}
