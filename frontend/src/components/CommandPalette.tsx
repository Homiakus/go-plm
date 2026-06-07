// Command Palette (Ctrl+K) — keyboard-first command interface
import { useState, useEffect, useRef, useMemo } from "react";
import { Search, Plus, Table, Tag, GitBranch, RefreshCw, Database, Nut } from "lucide-react";

interface Command {
  id: string;
  label: string;
  icon: React.ReactNode;
  shortcut?: string;
  action: () => void;
  category: string;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  commands: Command[];
}

export default function CommandPalette({ isOpen, onClose, commands }: CommandPaletteProps) {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen) {
      setQuery("");
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  }, [isOpen]);

  const filtered = useMemo(() => {
    if (!query.trim()) return commands;
    const q = query.toLowerCase();
    return commands.filter(
      (c) => c.label.toLowerCase().includes(q) || c.category.toLowerCase().includes(q)
    );
  }, [commands, query]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      onClose();
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && filtered[selectedIndex]) {
      filtered[selectedIndex].action();
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]" onClick={onClose}>
      <div
        className="bg-white dark:bg-gray-900 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 w-full max-w-lg overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Search */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <Search size={18} className="text-gray-400 shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => { setQuery(e.target.value); setSelectedIndex(0); }}
            onKeyDown={handleKeyDown}
            placeholder="Type a command..."
            className="flex-1 bg-transparent text-lg focus:outline-none"
          />
          <kbd className="text-xs text-gray-400 bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded font-mono">
            esc
          </kbd>
        </div>

        {/* Results */}
        <div className="max-h-80 overflow-y-auto p-2">
          {filtered.length === 0 ? (
            <div className="text-center py-8 text-gray-400 text-sm">No commands found</div>
          ) : (
            filtered.map((cmd, i) => (
              <button
                key={cmd.id}
                className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                  i === selectedIndex
                    ? "bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400"
                    : "text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                }`}
                onClick={() => { cmd.action(); onClose(); }}
                onMouseEnter={() => setSelectedIndex(i)}
              >
                <span className="shrink-0 text-gray-400">{cmd.icon}</span>
                <span className="flex-1 text-left">{cmd.label}</span>
                <span className="text-xs text-gray-400">{cmd.category}</span>
                {cmd.shortcut && (
                  <kbd className="text-xs text-gray-400 bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded font-mono">
                    {cmd.shortcut}
                  </kbd>
                )}
              </button>
            ))
          )}
        </div>

        {/* Footer */}
        <div className="px-4 py-2 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-400 flex gap-4">
          <span>↑↓ Navigate</span>
          <span>↵ Execute</span>
          <span>Esc Close</span>
        </div>
      </div>
    </div>
  );
}

// Standard command palette items factory
export function getStandardCommands(callbacks: {
  onCreateObject: () => void;
  onSearch: () => void;
  onViewBOM: () => void;
  onRefresh: () => void;
  onRebuildIndex: () => void;
  onGitCheckpoint: () => void;
  onGitChanges: () => void;
  onRelease: () => void;
  onStandardParts: () => void;
}): Command[] {
  return [
    { id: "create-object", label: "Create Object...", icon: <Plus size={16} />, shortcut: "Ctrl+N", category: "Objects", action: callbacks.onCreateObject },
    { id: "search", label: "Search Objects", icon: <Search size={16} />, shortcut: "Ctrl+P", category: "Objects", action: callbacks.onSearch },
    { id: "view-bom", label: "View BOM", icon: <Table size={16} />, shortcut: "Ctrl+B", category: "BOM", action: callbacks.onViewBOM },
    { id: "rebuild-index", label: "Rebuild Search Index", icon: <RefreshCw size={16} />, category: "Index", action: callbacks.onRebuildIndex },
    { id: "refresh", label: "Refresh Project", icon: <RefreshCw size={16} />, shortcut: "F5", category: "Project", action: callbacks.onRefresh },
    { id: "git-checkpoint", label: "Create Git Checkpoint", icon: <GitBranch size={16} />, category: "Git", action: callbacks.onGitCheckpoint },
    { id: "git-changes", label: "Show Changes", icon: <GitBranch size={16} />, category: "Git", action: callbacks.onGitChanges },
    { id: "release-dashboard", label: "Open Release Dashboard", icon: <Tag size={16} />, category: "Release", action: callbacks.onRelease },
    { id: "standard-parts", label: "Open Standard Parts", icon: <Nut size={16} />, category: "Standard Parts", action: callbacks.onStandardParts },
    { id: "rebuild", label: "Rebuild Index from Files", icon: <Database size={16} />, category: "Index", action: callbacks.onRebuildIndex },
  ];
}
