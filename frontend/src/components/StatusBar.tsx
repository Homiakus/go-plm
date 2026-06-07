import { useState, useEffect } from "react";
import type { ObjectDTO, ProjectInfo } from "../types";
import { GitBranch, Database, HardDrive } from "lucide-react";
import api from "../bindings/api";

interface StatusBarProps {
  project: ProjectInfo | null;
  selectedObject: ObjectDTO | null;
}

export default function StatusBar({ project, selectedObject }: StatusBarProps) {
  const [stats, setStats] = useState<{
    objects: number;
    relations: number;
    artifacts: number;
  } | null>(null);
  const [gitClean, setGitClean] = useState(true);

  useEffect(() => {
    api.stats().then(setStats).catch(() => {});
    api.gitStatus().then((s) => setGitClean(s.clean)).catch(() => {});
  }, []);

  return (
    <footer className="border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-950 px-4 py-1.5 flex items-center gap-4 text-xs text-gray-400">
      {/* Git status */}
      <div className="flex items-center gap-1.5">
        <GitBranch size={12} />
        <span className={gitClean ? "text-green-500" : "text-amber-500"}>
          {gitClean ? "clean" : "modified"}
        </span>
      </div>

      {/* Index stats */}
      {stats && (
        <div className="flex items-center gap-1.5">
          <Database size={12} />
          <span>
            {stats.objects} objects · {stats.relations} rels · {stats.artifacts}{" "}
            artifacts
          </span>
        </div>
      )}

      <div className="flex-1" />

      {/* Selected object path */}
      {selectedObject && (
        <div className="flex items-center gap-1.5">
          <HardDrive size={12} />
          <span className="font-mono">
            objects/{selectedObject.id}/{selectedObject.id}.md
          </span>
        </div>
      )}

      {/* Project */}
      {project && (
        <span className="text-gray-500 font-medium">{project.code}</span>
      )}
    </footer>
  );
}
