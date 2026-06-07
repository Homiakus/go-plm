// Right-click context menu for Project Tree items
import { useEffect, useRef } from "react";
import { Eye, GitBranch, Trash2, Link2, FilePlus, Copy, ExternalLink } from "lucide-react";

export interface MenuItem {
  label?: string;
  icon?: string;
  action?: () => void;
  separator?: boolean;
  disabled?: boolean;
  danger?: boolean;
}

interface ContextMenuProps {
  x: number;
  y: number;
  items: MenuItem[];
  onClose: () => void;
}

const iconMap: Record<string, React.ReactNode> = {
  eye: <Eye size={14} />,
  "git-branch": <GitBranch size={14} />,
  trash: <Trash2 size={14} />,
  "link-2": <Link2 size={14} />,
  "file-plus": <FilePlus size={14} />,
  copy: <Copy size={14} />,
  "external-link": <ExternalLink size={14} />,
};

export function getNodeMenuItems(
  objectId: string,
  objectClass: string,
  objectState: string,
  callbacks: {
    onOpen: (id: string) => void;
    onCreateChild: (parentId: string) => void;
    onDelete: (id: string) => void;
    onCopyId: (id: string) => void;
    onViewBOM: (id: string) => void;
    onViewWhereUsed: (id: string) => void;
    onTransition: (id: string, transition: string) => void;
  }
): MenuItem[] {
  const items: MenuItem[] = [
    { label: "Open", icon: "eye", action: () => callbacks.onOpen(objectId) },
    { label: "Copy ID", icon: "copy", action: () => callbacks.onCopyId(objectId) },
    { separator: true },
  ];

  // Assembly-specific
  if (objectClass === "asm") {
    items.push({ label: "Add Child to BOM...", icon: "file-plus", action: () => callbacks.onCreateChild(objectId) });
    items.push({ label: "View BOM", icon: "git-branch", action: () => callbacks.onViewBOM(objectId) });
  }

  items.push({ label: "Where Used", icon: "link-2", action: () => callbacks.onViewWhereUsed(objectId) });

  // Lifecycle actions based on state
  if (objectState === "draft") {
    items.push({ separator: true });
    items.push({ label: "Submit for Review", icon: "git-branch", action: () => callbacks.onTransition(objectId, "submit_review") });
  } else if (objectState === "in_review") {
    items.push({ separator: true });
    items.push({ label: "Approve", icon: "git-branch", action: () => callbacks.onTransition(objectId, "approve") });
    items.push({ label: "Reject", action: () => callbacks.onTransition(objectId, "reject") });
  } else if (objectState === "approved") {
    items.push({ separator: true });
    items.push({ label: "Release", icon: "git-branch", action: () => callbacks.onTransition(objectId, "release") });
  } else if (objectState === "released") {
    items.push({ separator: true });
    items.push({ label: "Revise", icon: "git-branch", action: () => callbacks.onTransition(objectId, "revise") });
  }

  items.push({ separator: true });
  items.push({ label: "Delete", icon: "trash", action: () => callbacks.onDelete(objectId), danger: true });

  return items;
}

export default function ContextMenu({ x, y, items, onClose }: ContextMenuProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    };
    const keyHandler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("mousedown", handler);
    document.addEventListener("keydown", keyHandler);
    return () => {
      document.removeEventListener("mousedown", handler);
      document.removeEventListener("keydown", keyHandler);
    };
  }, [onClose]);

  // Adjust position to stay within viewport
  const adjustedX = Math.min(x, window.innerWidth - 220);
  const adjustedY = Math.min(y, window.innerHeight - items.length * 36 - 20);

  return (
    <div
      ref={ref}
      className="fixed z-50 bg-white dark:bg-gray-850 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl py-1 min-w-[200px]"
      style={{ left: adjustedX, top: adjustedY }}
    >
      {items.map((item, i) =>
        item.separator ? (
          <div key={i} className="border-t border-gray-100 dark:border-gray-700 my-1" />
        ) : (
          <button
            key={i}
            onClick={() => { item.action?.(); onClose(); }}
            disabled={item.disabled}
            className={`w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left transition-colors ${
              item.danger
                ? "text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
                : item.disabled
                ? "text-gray-300 cursor-not-allowed"
                : "text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
            }`}
          >
            {item.icon && <span className="shrink-0">{iconMap[item.icon]}</span>}
            {item.label}
          </button>
        )
      )}
    </div>
  );
}
