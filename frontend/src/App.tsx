import { useState, useEffect, useCallback } from "react";
import api from "./bindings/api";
import type {
  ObjectDTO,
  TreeNodeDTO,
  TreeRequest,
  ViewMode,
  ProjectInfo,
} from "./types";
import LandingPage from "./components/LandingPage";
import Sidebar from "./components/Sidebar";
import Toolbar from "./components/Toolbar";
import ObjectEditor from "./components/ObjectEditor";
import MarkdownEditor from "./components/MarkdownEditor";
import BOMViewer from "./components/BOMViewer";
import SearchPanel from "./components/SearchPanel";
import StatusBar from "./components/StatusBar";
import CreateWizard from "./components/CreateWizard";
import CommandPalette from "./components/CommandPalette";
import { getStandardCommands } from "./components/CommandPalette";
import SafeDeleteDialog from "./components/SafeDeleteDialog";
import StandardPartsCenter from "./components/StandardPartsCenter";
import GitChangesPanel from "./components/GitChangesPanel";

export default function App() {
  const [project, setProject] = useState<ProjectInfo | null>(null);
  const [objects, setObjects] = useState<ObjectDTO[]>([]);
  const [treeNodes, setTreeNodes] = useState<TreeNodeDTO[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [selectedObject, setSelectedObject] = useState<ObjectDTO | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>("browse");
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showWizard, setShowWizard] = useState(false);
  const [showCommandPalette, setShowCommandPalette] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<ObjectDTO | null>(null);

  // Load project data on mount
  useEffect(() => {
    loadProject();
  }, []);

  const loadProject = async () => {
    try {
      setLoading(true);
      setError(null);
      const info = await api.projectInfo();
      setProject(info);
      const nodes = await api.getTree({ offset: 0, limit: 200 });
      setTreeNodes(nodes || []);
      const objs = await api.listObjects();
      setObjects(objs || []);
    } catch (err: any) {
      setError(err.message || "Failed to load project");
    } finally {
      setLoading(false);
    }
  };

  // Select an object
  const selectObject = useCallback(
    async (id: string) => {
      setSelectedId(id);
      setViewMode("browse");
      try {
        const obj = await api.getObject(id);
        setSelectedObject(obj);
      } catch {
        setSelectedObject(null);
      }
    },
    []
  );

  // Refresh tree after mutations
  const refreshTree = useCallback(async () => {
    const nodes = await api.getTree({ offset: 0, limit: 200 });
    setTreeNodes(nodes || []);
    const objs = await api.listObjects();
    setObjects(objs || []);
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "k") {
        e.preventDefault();
        setShowCommandPalette((v) => !v);
      }
      if ((e.ctrlKey || e.metaKey) && e.key === "n") {
        e.preventDefault();
        setShowWizard(true);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  // Handle delete
  const handleDeleteObject = async (objectId: string) => {
    try {
      const obj = await api.getObject(objectId);
      if (obj) setDeleteTarget(obj);
    } catch (e: any) {
      setError(e.message);
    }
  };

  const confirmDelete = async (objectId: string, _force: boolean) => {
    try {
      await api.deleteObject(objectId);
      setDeleteTarget(null);
      await refreshTree();
      if (selectedId === objectId) {
        setSelectedId(null);
        setSelectedObject(null);
      }
    } catch (e: any) {
      setError(e.message);
    }
  };

  // Handle creating a new object
  const handleCreateObject = async (
    classType: string,
    title: string,
    metadata?: Record<string, unknown>
  ) => {
    try {
      const resp = await api.createObject({
        project_path: project?.root || ".",
        class: classType,
        title,
        metadata,
      });
      await refreshTree();
      if (resp.object_id) {
        selectObject(resp.object_id);
      }
    } catch (err: any) {
      setError(err.message || "Failed to create object");
    }
  };

  // Handle lifecycle transition
  const handleTransition = async (objectId: string, transition: string) => {
    try {
      await api.runTransition({ object_id: objectId, transition });
      await refreshTree();
      if (selectedId === objectId) {
        selectObject(objectId);
      }
    } catch (err: any) {
      setError(err.message || "Transition failed");
    }
  };

  const handleViewBOM = (id: string) => {
    setSelectedId(id);
    setViewMode("bom");
  };

  // Landing page when no project
  if (!project && !loading) {
    return (
      <LandingPage
        onCreateProject={async (name, title) => {
          setLoading(true);
          try {
            await api.projectInfo(); // will fail, so init via API
            await loadProject();
          } catch {
            // Project needs CLI init — reload
            window.location.reload();
          }
        }}
        onOpenProject={async (path) => {
          setLoading(true);
          window.location.href = `/?path=${encodeURIComponent(path)}`;
        }}
      />
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full bg-gray-50 dark:bg-gray-900">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4" />
          <p className="text-gray-500">Loading project...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-900">
      {/* Toolbar */}
      <Toolbar
        project={project}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        selectedObject={selectedObject}
        onCreateObject={() => setShowWizard(true)}
        onRefresh={refreshTree}
        onSearch={setSearchQuery}
      />

      {/* Create Wizard modal */}
      {showWizard && (
        <CreateWizard
          onCreate={async (cls, title, meta, parentId) => {
            await handleCreateObject(cls, title, meta);
            setShowWizard(false);
          }}
          onCancel={() => setShowWizard(false)}
          suggestedParentId={selectedId || undefined}
        />
      )}

      {/* Safe Delete dialog */}
      {deleteTarget && (
        <SafeDeleteDialog
          object={deleteTarget}
          onDelete={confirmDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}

      {/* Command Palette */}
      <CommandPalette
        isOpen={showCommandPalette}
        onClose={() => setShowCommandPalette(false)}
        commands={getStandardCommands({
          onCreateObject: () => setShowWizard(true),
          onSearch: () => setViewMode("search"),
          onViewBOM: () => selectedId && setViewMode("bom"),
          onRefresh: refreshTree,
          onRebuildIndex: async () => { await api.rebuildIndex(); refreshTree(); },
          onGitCheckpoint: async () => {
            const msg = prompt("Checkpoint message:");
            if (msg) await api.createCheckpoint(msg);
          },
          onCreateRelease: () => {},
        })}
      />

      {/* Error banner */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/30 border-b border-red-200 dark:border-red-800 px-4 py-2 text-sm text-red-700 dark:text-red-400 flex items-center justify-between">
          <span>{error}</span>
          <button onClick={() => setError(null)} className="font-bold ml-4">
            ×
          </button>
        </div>
      )}

      {/* Main content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <Sidebar
          treeNodes={treeNodes}
          selectedId={selectedId}
          onSelect={selectObject}
          onViewBOM={handleViewBOM}
          onDelete={handleDeleteObject}
          onCreateChild={(parentId) => { setSelectedId(parentId); setShowWizard(true); }}
          onTransition={handleTransition}
          searchQuery={searchQuery}
        />

        {/* Content area */}
        <main className="flex-1 flex flex-col overflow-hidden">
          {viewMode === "search" ? (
            <SearchPanel
              query={searchQuery}
              onSelect={selectObject}
              onClose={() => setViewMode("browse")}
            />
          ) : viewMode === "stdparts" ? (
            <StandardPartsCenter />
          ) : viewMode === "gitchanges" ? (
            <GitChangesPanel />
          ) : viewMode === "bom" && selectedId ? (
            <BOMViewer rootId={selectedId} onSelect={selectObject} />
          ) : selectedObject ? (
            <ObjectEditor
              object={selectedObject}
              onTransition={(t) => handleTransition(selectedObject!.id, t)}
              onViewBOM={() => handleViewBOM(selectedObject!.id)}
            />
          ) : (
            <WelcomeScreen
              project={project}
              stats={treeNodes.length}
              onRefresh={refreshTree}
            />
          )}
        </main>
      </div>

      {/* Status bar */}
      <StatusBar project={project} selectedObject={selectedObject} />
    </div>
  );
}

function WelcomeScreen({
  project,
  stats,
  onRefresh,
}: {
  project: ProjectInfo | null;
  stats: number;
  onRefresh: () => void;
}) {
  return (
    <div className="flex-1 flex items-center justify-center p-8">
      <div className="text-center max-w-lg">
        <div className="text-6xl mb-6">📦</div>
        <h1 className="text-2xl font-bold mb-2">
          {project?.title || "go-plm"}
        </h1>
        <p className="text-gray-500 mb-8">
          Local Git-native PLM/PDM system for engineering documentation
        </p>
        <div className="grid grid-cols-3 gap-4 mb-8">
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
            <div className="text-2xl font-bold text-blue-600">{stats}</div>
            <div className="text-xs text-gray-500">Objects</div>
          </div>
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
            <div className="text-2xl font-bold text-green-600">📝</div>
            <div className="text-xs text-gray-500">Markdown+YAML</div>
          </div>
          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
            <div className="text-2xl font-bold text-purple-600">🔀</div>
            <div className="text-xs text-gray-500">Git-native</div>
          </div>
        </div>
        <p className="text-sm text-gray-400">
          Select an object from the sidebar or create a new one to get started.
        </p>
      </div>
    </div>
  );
}
