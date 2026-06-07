// Release Dashboard — readiness check with blocker visualization and Explain Blockers
import { useState, useEffect } from "react";
import { CheckCircle, XCircle, AlertTriangle, Loader2, Tag, List } from "lucide-react";
import type { Diagnostic, ObjectDTO } from "../types";
import api from "../bindings/api";
import { STATE_COLORS, STATE_LABELS } from "../types";

interface ReleaseDashboardProps {
  object: ObjectDTO;
}

export default function ReleaseDashboard({ object }: ReleaseDashboardProps) {
  const [ready, setReady] = useState(false);
  const [blockers, setBlockers] = useState<Diagnostic[]>([]);
  const [scope, setScope] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [createdId, setCreatedId] = useState<string | null>(null);

  const handleCreateRelease = async () => {
    setCreating(true);
    setError(null);
    try {
      const id = await api.createReleasePackage(object.id, object.title + " Release");
      setCreatedId(id);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setCreating(false);
    }
  };

  useEffect(() => {
    setLoading(true);
    setError(null);
    Promise.all([
      api.checkReleaseReadiness(object.id),
      api.buildReleaseScope(object.id).catch(() => []),
    ]).then(([readiness, scopeIds]) => {
      setReady(readiness.ready);
      setBlockers(readiness.blockers || []);
      setScope(scopeIds);
      setLoading(false);
    }).catch((e) => {
      setError(e.message);
      setLoading(false);
    });
  }, [object.id]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 size={24} className="animate-spin text-gray-400" />
        <span className="ml-2 text-sm text-gray-400">Checking release readiness...</span>
      </div>
    );
  }

  if (error) {
    return <div className="text-red-500 text-sm p-4">{error}</div>;
  }

  return (
    <div className="space-y-6">
      {/* Status banner */}
      <div className={`p-4 rounded-xl border-2 ${ready ? "border-green-500 bg-green-50 dark:bg-green-900/20" : "border-red-500 bg-red-50 dark:bg-red-900/20"}`}>
        <div className="flex items-center gap-3">
          {ready ? (
            <CheckCircle size={32} className="text-green-500" />
          ) : (
            <XCircle size={32} className="text-red-500" />
          )}
          <div>
            <h3 className={`text-lg font-bold ${ready ? "text-green-700 dark:text-green-400" : "text-red-700 dark:text-red-400"}`}>
              {ready ? "Ready for Release" : "Not Ready for Release"}
            </h3>
            <p className="text-sm text-gray-500 mt-0.5">
              {ready
                ? "All checks passed. Release can proceed."
                : `${blockers.length} blocker${blockers.length !== 1 ? "s" : ""} must be resolved before release.`}
            </p>
          </div>
        </div>
      </div>

      {/* Blockers — Explain Blockers (Why Can't I) */}
      {blockers.length > 0 && (
        <div>
          <h4 className="text-sm font-semibold flex items-center gap-2 mb-3">
            <AlertTriangle size={16} className="text-amber-500" />
            Why Can't I Release?
          </h4>
          <div className="space-y-2">
            {blockers.map((b) => (
              <div key={b.id} className="bg-amber-50 dark:bg-amber-900/10 border border-amber-200 dark:border-amber-800 rounded-lg p-3">
                <div className="flex items-start gap-2">
                  <XCircle size={16} className="text-red-500 mt-0.5 shrink-0" />
                  <div>
                    <p className="text-sm font-medium">{b.message}</p>
                    <div className="flex items-center gap-2 mt-1">
                      <code className="text-xs bg-amber-100 dark:bg-amber-900/30 px-1.5 py-0.5 rounded">{b.code}</code>
                      {b.object_id && <span className="text-xs text-gray-400 font-mono">{b.object_id}</span>}
                    </div>
                    {b.suggested_actions && b.suggested_actions.length > 0 && (
                      <div className="flex gap-2 mt-2">
                        {b.suggested_actions.map((action) => (
                          <span key={action} className="text-xs px-2 py-0.5 bg-white dark:bg-gray-800 rounded-full border border-amber-200 dark:border-amber-700 text-amber-700 dark:text-amber-400">
                            {action.replace(/_/g, " ")}
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
      )}

      {/* Scope summary */}
      {scope.length > 0 && (
        <div>
          <h4 className="text-sm font-semibold flex items-center gap-2 mb-3">
            <List size={16} className="text-gray-500" />
            Release Scope ({scope.length} objects)
          </h4>
          <div className="text-xs text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg p-3 font-mono max-h-32 overflow-y-auto">
            {scope.slice(0, 10).map((id) => (
              <div key={id}>{id}</div>
            ))}
            {scope.length > 10 && <div className="text-gray-400 mt-1">... and {scope.length - 10} more</div>}
          </div>
        </div>
      )}

      {/* Ready actions */}
      {ready && (
        <div className="space-y-3">
          {createdId ? (
            <div className="p-4 bg-green-50 dark:bg-green-900/20 rounded-xl border border-green-200 dark:border-green-800">
              <CheckCircle size={24} className="text-green-500 mb-2" />
              <p className="font-medium text-green-700 dark:text-green-400">Release Package Created!</p>
              <code className="text-xs text-green-600 dark:text-green-500 font-mono mt-1 block">{createdId}</code>
            </div>
          ) : (
            <div className="flex gap-3">
              <button
                onClick={handleCreateRelease}
                disabled={creating}
                className="flex items-center gap-2 px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-300 text-white rounded-lg text-sm font-medium transition-colors"
              >
                <Tag size={16} />
                {creating ? "Creating..." : "Create Release Package"}
              </button>
            </div>
          )}
        </div>
      )}
      {error && <div className="text-red-500 text-sm">{error}</div>}
    </div>
  );
}
