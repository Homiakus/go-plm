import { useState, useEffect } from "react";
import type { ObjectDTO, ViewMode, ProjectInfo } from "../types";
import {
  Plus,
  RefreshCw,
  Search,
  FileText,
  Table,
  FolderTree,
  Sun,
  Moon,
  Tag,
  Nut,
  GitBranch,
} from "lucide-react";

interface ToolbarProps {
  project: ProjectInfo | null;
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  selectedObject: ObjectDTO | null;
  onCreateObject: () => void;
  onRefresh: () => void;
  onSearch: (query: string) => void;
}

export default function Toolbar({
  project,
  viewMode,
  onViewModeChange,
  selectedObject,
  onCreateObject,
  onRefresh,
  onSearch,
}: ToolbarProps) {
  const [dark, setDark] = useState(() =>
    document.documentElement.classList.contains("dark")
  );

  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    localStorage.setItem("go-plm-dark", String(dark));
  }, [dark]);

  return (
    <header className="border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
      {/* Main toolbar */}
      <div className="flex items-center gap-2 px-4 py-2">
        {/* Project title */}
        <div className="flex items-center gap-2 mr-4">
          <FolderTree size={20} className="text-blue-500" />
          <span className="font-semibold text-sm">
            {project?.title || "go-plm"}
          </span>
        </div>

        {/* View modes */}
        <div className="flex items-center gap-1 bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
          {(["browse", "bom", "search", "release", "stdparts", "gitchanges"] as ViewMode[]).map((mode) => (
            <button
              key={mode}
              onClick={() => onViewModeChange(mode)}
              disabled={mode === "release" && !selectedObject}
              className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
                viewMode === mode
                  ? "bg-white dark:bg-gray-700 shadow-sm text-blue-600 dark:text-blue-400"
                  : mode === "release" && !selectedObject
                  ? "text-gray-300 cursor-not-allowed"
                  : "text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
              }`}
            >
              {mode === "browse" && <FileText size={14} className="inline mr-1" />}
              {mode === "bom" && <Table size={14} className="inline mr-1" />}
              {mode === "search" && <Search size={14} className="inline mr-1" />}
              {mode === "release" && <Tag size={14} className="inline mr-1" />}
              {mode === "stdparts" && <Nut size={14} className="inline mr-1" />}
              {mode === "gitchanges" && <GitBranch size={14} className="inline mr-1" />}
              {mode === "browse"
                ? "Browse"
                : mode === "bom"
                ? "BOM"
                : mode === "search"
                ? "Search"
                : mode === "release"
                ? "Release"
                : mode === "stdparts"
                ? "Std Parts"
                : "Changes"}
            </button>
          ))}
        </div>

        <div className="flex-1" />

        {/* Search */}
        <div className="relative">
          <Search
            size={16}
            className="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400"
          />
          <input
            type="text"
            placeholder="Search (FTS5)..."
            onChange={(e) => onSearch(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                onViewModeChange("search");
              }
            }}
            className="w-56 pl-8 pr-3 py-1.5 text-sm rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* Create button — opens wizard */}
        <button
          onClick={onCreateObject}
          className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors"
        >
          <Plus size={16} />
          New
        </button>

        {/* Dark mode */}
        <button
          onClick={() => setDark(!dark)}
          className="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          title={dark ? "Light mode" : "Dark mode"}
        >
          {dark ? <Sun size={16} /> : <Moon size={16} />}
        </button>

        {/* Refresh */}
        <button
          onClick={onRefresh}
          className="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          title="Refresh"
        >
          <RefreshCw size={16} />
        </button>
      </div>
    </header>
  );
}
