import { useState, useEffect, useCallback } from "react";
import api from "./bindings/api";
import type {
  Diagnostic,
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
import BOMViewer from "./components/BOMViewer";
import SearchPanel from "./components/SearchPanel";
import StatusBar from "./components/StatusBar";
import CreateWizard from "./components/CreateWizard";
import CommandPalette from "./components/CommandPalette";
import { getStandardCommands } from "./components/CommandPalette";
import SafeDeleteDialog from "./components/SafeDeleteDialog";
import StandardPartsCenter from "./components/StandardPartsCenter";
import GitChangesPanel from "./components/GitChangesPanel";
import ReleaseDashboard from "./components/ReleaseDashboard";

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

  const confirmDelete = async (objectId: string, force: boolean) => {
    try {
      await api.deleteObject(objectId, force);
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
    metadata?: Record<string, unknown>,
    parentId?: string
  ) => {
    try {
      const resp = await api.createObject({
        project_path: project?.root || ".",
        class: classType,
        title,
        parent_object_id: parentId,
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
        onCreateProject={async (path, code, title) => {
          setLoading(true);
          try {
            await api.createProject({ path, code, title });
            await loadProject();
          } catch (e: any) {
            setError(e.message || "Failed to create project");
            setLoading(false);
          }
        }}
        onOpenProject={async (path) => {
          setLoading(true);
          try {
            await api.openProject(path);
            await loadProject();
          } catch (e: any) {
            setError(e.message || "Failed to open project");
            setLoading(false);
          }
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
            await handleCreateObject(cls, title, meta, parentId);
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
          onGitChanges: () => setViewMode("gitchanges"),
          onRelease: () => selectedObject && setViewMode("release"),
          onStandardParts: () => setViewMode("stdparts"),
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
          ) : viewMode === "release" && selectedObject ? (
            <div className="flex-1 overflow-y-auto p-6">
              <ReleaseDashboard object={selectedObject} />
            </div>
          ) : viewMode === "bom" && selectedId ? (
            <BOMViewer rootId={selectedId} onSelect={selectObject} />
          ) : selectedObject ? (
            <ObjectEditor
              object={selectedObject}
              onTransition={(t) => handleTransition(selectedObject!.id, t)}
              onViewBOM={() => handleViewBOM(selectedObject!.id)}
              onObjectUpdated={async () => {
                await refreshTree();
                if (selectedId) await selectObject(selectedId);
              }}
            />
          ) : (
            <WelcomeScreen
              project={project}
              stats={treeNodes.length}
              onRefresh={refreshTree}
            />
          )}
        </main>

        <RightInspector object={selectedObject} />
      </div>

      {/* Status bar */}
      <StatusBar project={project} selectedObject={selectedObject} />
    </div>
  );
}

function RightInspector({ object }: { object: ObjectDTO | null }) {
  const [diagnostics, setDiagnostics] = useState<Diagnostic[] | null>(null);

  useEffect(() => {
    setDiagnostics(null);
    if (!object) return;
    api.validateObject(object.id).then(setDiagnostics).catch(() => setDiagnostics(null));
  }, [object?.id]);

  return (
    <aside className="w-80 border-l border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-950 overflow-y-auto">
      <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
        <h2 className="text-sm font-semibold">Inspector</h2>
      </div>
      {!object ? (
        <div className="p-4 text-sm text-gray-400">Select an object to inspect properties and readiness.</div>
      ) : (
        <div className="p-4 space-y-5 text-sm">
          <section>
            <div className="text-xs font-semibold uppercase text-gray-400 mb-2">Object</div>
            <div className="font-medium truncate">{object.title}</div>
            <code className="block mt-1 text-xs text-blue-600 dark:text-blue-400 break-all">{object.id}</code>
            <div className="mt-2 grid grid-cols-2 gap-2 text-xs">
              <span className="text-gray-400">Class</span><span>{object.class}</span>
              <span className="text-gray-400">State</span><span>{object.state}</span>
              <span className="text-gray-400">Revision</span><span>{object.revision}</span>
            </div>
          </section>

          <section>
            <div className="text-xs font-semibold uppercase text-gray-400 mb-2">Readiness</div>
            {diagnostics == null ? (
              <div className="text-gray-400">Checking...</div>
            ) : diagnostics.length === 0 ? (
              <div className="text-green-600 dark:text-green-400">Ready</div>
            ) : (
              <div className="text-amber-600 dark:text-amber-400">
                {diagnostics.filter((d) => d.severity === "blocker").length} blockers, {diagnostics.filter((d) => d.severity !== "blocker").length} warnings
              </div>
            )}
          </section>

          <section>
            <div className="text-xs font-semibold uppercase text-gray-400 mb-2">Relations</div>
            <div>{object.relations?.length || 0} outgoing</div>
          </section>

          <section>
            <div className="text-xs font-semibold uppercase text-gray-400 mb-2">Files</div>
            <div>{object.artifacts?.length || 0} attached</div>
          </section>

          <section>
            <div className="text-xs font-semibold uppercase text-gray-400 mb-2">Path</div>
            <code className="text-xs break-all">objects/{object.id}/{object.id}.md</code>
          </section>
        </div>
      )}
    </aside>
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
