// Validation panel showing diagnostics grouped by severity with suggested actions
import { useState, useEffect } from "react";
import { AlertTriangle, XCircle, Info, AlertOctagon, Loader2 } from "lucide-react";
import type { Diagnostic, ObjectDTO } from "../types";
import api from "../bindings/api";

interface ValidationPanelProps {
  object: ObjectDTO;
}

export default function ValidationPanel({ object }: ValidationPanelProps) {
  const [diags, setDiags] = useState<Diagnostic[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      api.validateBOM(object.id).catch(() => []),
      api.detectBOMCycles(object.id).catch(() => []),
    ]).then(([bomDiags, cycleDiags]) => {
      setDiags([...bomDiags, ...cycleDiags]);
      setLoading(false);
    }).catch((e) => {
      setError(e.message);
      setLoading(false);
    });
  }, [object.id]);

  const blockers = diags.filter((d) => d.severity === "blocker");
  const warnings = diags.filter((d) => d.severity === "warning");
  const errors = diags.filter((d) => d.severity === "error");
  const infos = diags.filter((d) => d.severity === "info");

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 size={24} className="animate-spin text-gray-400" />
        <span className="ml-2 text-sm text-gray-400">Running validation...</span>
      </div>
    );
  }

  if (error) {
    return <div className="text-red-500 text-sm p-4">{error}</div>;
  }

  if (diags.length === 0) {
    return (
      <div className="text-center py-12 text-gray-400">
        <CheckCircle size={48} className="mx-auto mb-4 text-green-400 opacity-50" />
        <p className="font-medium">All checks passed</p>
        <p className="text-sm mt-1">No diagnostics found</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="flex gap-4">
        {blockers.length > 0 && (
          <div className="flex items-center gap-2 px-3 py-2 bg-red-50 dark:bg-red-900/20 rounded-lg">
            <XCircle size={16} className="text-red-500" />
            <span className="text-sm font-medium text-red-700 dark:text-red-400">{blockers.length} blocker{blockers.length !== 1 ? "s" : ""}</span>
          </div>
        )}
        {errors.length > 0 && (
          <div className="flex items-center gap-2 px-3 py-2 bg-red-50 dark:bg-red-900/20 rounded-lg">
            <AlertOctagon size={16} className="text-red-500" />
            <span className="text-sm font-medium text-red-700">{errors.length} error{errors.length !== 1 ? "s" : ""}</span>
          </div>
        )}
        {warnings.length > 0 && (
          <div className="flex items-center gap-2 px-3 py-2 bg-amber-50 dark:bg-amber-900/20 rounded-lg">
            <AlertTriangle size={16} className="text-amber-500" />
            <span className="text-sm font-medium text-amber-700 dark:text-amber-400">{warnings.length} warning{warnings.length !== 1 ? "s" : ""}</span>
          </div>
        )}
        {infos.length > 0 && (
          <div className="flex items-center gap-2 px-3 py-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <Info size={16} className="text-blue-500" />
            <span className="text-sm font-medium text-blue-700 dark:text-blue-400">{infos.length} info</span>
          </div>
        )}
      </div>

      {/* Diagnostics list */}
      <div className="space-y-2">
        {diags.map((d) => (
          <div
            key={d.id}
            className={`p-3 rounded-lg border-l-4 ${
              d.severity === "blocker"
                ? "border-l-red-500 bg-red-50/50 dark:bg-red-900/10"
                : d.severity === "error"
                ? "border-l-red-400 bg-red-50/30 dark:bg-red-900/5"
                : d.severity === "warning"
                ? "border-l-amber-500 bg-amber-50/50 dark:bg-amber-900/10"
                : "border-l-blue-500 bg-blue-50/50 dark:bg-blue-900/10"
            }`}
          >
            <div className="flex items-start gap-2">
              {d.severity === "blocker" ? (
                <XCircle size={16} className="text-red-500 mt-0.5 shrink-0" />
              ) : d.severity === "warning" ? (
                <AlertTriangle size={16} className="text-amber-500 mt-0.5 shrink-0" />
              ) : (
                <Info size={16} className="text-blue-500 mt-0.5 shrink-0" />
              )}
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-0.5">
                  <code className="text-xs bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded font-mono">{d.code}</code>
                  {d.path && <span className="text-xs text-gray-400 font-mono">{d.path}</span>}
                </div>
                <p className="text-sm">{d.message}</p>
                {d.suggested_actions && d.suggested_actions.length > 0 && (
                  <div className="flex gap-2 mt-2">
                    {d.suggested_actions.map((action) => (
                      <span key={action} className="text-xs px-2 py-0.5 bg-gray-100 dark:bg-gray-800 rounded-full text-blue-600 dark:text-blue-400 font-mono">
                        {action}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function CheckCircle({ size, className }: { size: number; className?: string }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className={className}>
      <circle cx="12" cy="12" r="10" />
      <path d="M9 12l2 2 4-4" />
    </svg>
  );
}
