import { useState, useMemo, useCallback } from "react";
import type { TreeNodeDTO } from "../types";
import { stateDotClasses } from "../utils";
import ContextMenu, { getNodeMenuItems, type MenuItem } from "./ContextMenu";
import {
  Package,
  Component,
  Ruler,
  FileText,
  Nut,
  BrickWall,
  Scissors,
  CornerDownRight,
  Cpu,
  SearchCheck,
  ClipboardList,
  ListChecks,
  Table,
  Tag,
  GitPullRequest,
  File,
  ChevronRight,
  ChevronDown,
} from "lucide-react";

interface SidebarProps {
  treeNodes: TreeNodeDTO[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onViewBOM: (id: string) => void;
  onDelete: (id: string) => void;
  onCreateChild: (parentId: string) => void;
  onTransition: (id: string, transition: string) => void;
  searchQuery: string;
}

const iconMap: Record<string, React.ComponentType<any>> = {
  package: Package,
  component: Component,
  ruler: Ruler,
  "file-text": FileText,
  nut: Nut,
  "brick-wall": BrickWall,
  scissors: Scissors,
  "corner-down-right": CornerDownRight,
  cpu: Cpu,
  "search-check": SearchCheck,
  "clipboard-list": ClipboardList,
  "list-checks": ListChecks,
  table: Table,
  tag: Tag,
  "git-pull-request": GitPullRequest,
};

function getIcon(iconName: string) {
  const Icon = iconMap[iconName] || File;
  return Icon;
}

export default function Sidebar({
  treeNodes,
  selectedId,
  onSelect,
  onViewBOM,
  onDelete,
  onCreateChild,
  onTransition,
  searchQuery,
}: SidebarProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [filter, setFilter] = useState("");
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; items: MenuItem[] } | null>(null);

  const handleContextMenu = useCallback(
    (e: React.MouseEvent, node: TreeNodeDTO) => {
      e.preventDefault();
      const items = getNodeMenuItems(node.id, node.class, node.state, {
        onOpen: onSelect,
        onCreateChild,
        onDelete,
        onCopyId: (id) => navigator.clipboard.writeText(id),
        onViewBOM,
        onViewWhereUsed: onSelect,
        onTransition,
      });
      setContextMenu({ x: e.clientX, y: e.clientY, items });
    },
    [onSelect, onCreateChild, onDelete, onViewBOM, onTransition]
  );

  const filtered = useMemo(() => {
    if (!filter && !searchQuery) return treeNodes;
    const q = (searchQuery || filter).toLowerCase();
    return treeNodes.filter(
      (n) =>
        n.title.toLowerCase().includes(q) ||
        n.id.toLowerCase().includes(q) ||
        n.class.includes(q)
    );
  }, [treeNodes, filter, searchQuery]);

  const toggle = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  return (
    <aside className="w-72 border-r border-gray-200 dark:border-gray-700 flex flex-col bg-gray-50 dark:bg-gray-950 overflow-hidden">
      {/* Filter */}
      <div className="p-3 border-b border-gray-200 dark:border-gray-700">
        <input
          type="text"
          placeholder="Filter objects..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="w-full px-3 py-1.5 text-sm rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-y-auto p-2">
        {filtered.length === 0 ? (
          <div className="text-center text-gray-400 text-sm py-8">
            No objects found
          </div>
        ) : (
          filtered.map((node) => {
            const Icon = getIcon(node.icon);
            const isSelected = selectedId === node.id;
            const isExpanded = expanded.has(node.id);
            const dotColor = stateDotClasses[node.status_color] || stateDotClasses.gray;

            return (
              <div key={node.id}>
                <div
                  className={`flex items-center gap-2 px-2 py-1.5 rounded-md cursor-pointer text-sm group transition-colors ${
                    isSelected
                      ? "bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-300"
                      : "hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300"
                  }`}
                  onClick={() => onSelect(node.id)}
                  onContextMenu={(e) => handleContextMenu(e, node)}
                >
                  {/* Expand toggle for assemblies */}
                  {node.has_children && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        toggle(node.id);
                      }}
                      className="shrink-0 p-0.5 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
                    >
                      {isExpanded ? (
                        <ChevronDown size={14} />
                      ) : (
                        <ChevronRight size={14} />
                      )}
                    </button>
                  )}
                  {!node.has_children && <div className="w-5" />}

                  <Icon size={16} className="shrink-0 text-gray-400" />
                  <span className="flex-1 truncate">{node.title}</span>

                  {/* Status dot */}
                  <span
                    className={`inline-block w-2.5 h-2.5 rounded-full shrink-0 ${dotColor}`}
                    title={node.state}
                  />

                  {/* BOM button for assemblies */}
                  {node.class === "asm" && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onViewBOM(node.id);
                      }}
                      className="opacity-0 group-hover:opacity-100 text-xs px-1.5 py-0.5 rounded bg-blue-100 dark:bg-blue-900/50 text-blue-700 dark:text-blue-400 hover:bg-blue-200 transition-opacity"
                      title="View BOM"
                    >
                      BOM
                    </button>
                  )}
                </div>

                {/* Child count and warnings */}
                {node.has_children && isExpanded && (
                  <div className="ml-8 text-xs text-gray-400 py-1">
                    {node.children_count} children
                    {node.warnings > 0 && (
                      <span className="text-amber-500 ml-2">
                        {node.warnings} ⚠
                      </span>
                    )}
                    {node.blockers > 0 && (
                      <span className="text-red-500 ml-2">
                        {node.blockers} ✗
                      </span>
                    )}
                  </div>
                )}

                {/* Thumbnail preview */}
                {node.thumbnail_url && (
                  <div className="ml-11 mb-1">
                    <img
                      src={node.thumbnail_url}
                      alt={node.title}
                      className="w-16 h-12 object-cover rounded border border-gray-200 dark:border-gray-700"
                      onError={(e) => {
                        (e.target as HTMLImageElement).style.display = "none";
                      }}
                    />
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Stats footer */}
      <div className="px-3 py-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-400 flex justify-between">
        <span>{filtered.length} objects</span>
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          items={contextMenu.items}
          onClose={() => setContextMenu(null)}
        />
      )}
    </aside>
  );
}
