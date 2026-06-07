// Safe Delete confirmation dialog with where-used impact analysis
import { useState, useEffect } from "react";
import { AlertTriangle, Trash2, Loader2, Link2 } from "lucide-react";
import type { ObjectDTO } from "../types";
import api from "../bindings/api";

interface SafeDeleteDialogProps {
  object: ObjectDTO;
  onDelete: (objectId: string, force: boolean) => Promise<void>;
  onCancel: () => void;
}

export default function SafeDeleteDialog({ object, onDelete, onCancel }: SafeDeleteDialogProps) {
  const [usedBy, setUsedBy] = useState<ObjectDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    api.whereUsed(object.id)
      .then(setUsedBy)
      .catch(() => setUsedBy([]))
      .finally(() => setLoading(false));
  }, [object.id]);

  const handleDelete = async (force: boolean) => {
    setDeleting(true);
    try {
      await onDelete(object.id, force);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onCancel}>
      <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-md p-6" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
            <Trash2 size={20} className="text-red-500" />
          </div>
          <div>
            <h3 className="text-lg font-bold">Delete Object</h3>
            <p className="text-sm text-gray-500">{object.id}</p>
          </div>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-6">
            <Loader2 size={20} className="animate-spin text-gray-400" />
            <span className="ml-2 text-sm text-gray-400">Checking references...</span>
          </div>
        ) : usedBy.length > 0 ? (
          <div>
            <div className="flex items-center gap-2 p-3 bg-amber-50 dark:bg-amber-900/20 rounded-lg mb-4">
              <AlertTriangle size={18} className="text-amber-500 shrink-0" />
              <div>
                <p className="text-sm font-medium text-amber-700 dark:text-amber-400">Object is referenced by {usedBy.length} other object{usedBy.length !== 1 ? "s" : ""}</p>
                <p className="text-xs text-amber-600 dark:text-amber-500 mt-0.5">Deleting may break BOM structures and relations</p>
              </div>
            </div>

            <div className="max-h-40 overflow-y-auto mb-4 space-y-1">
              {usedBy.map((obj) => (
                <div key={obj.id} className="flex items-center gap-2 text-sm px-3 py-1.5 bg-gray-50 dark:bg-gray-800 rounded">
                  <Link2 size={14} className="text-gray-400 shrink-0" />
                  <span className="font-mono text-xs text-blue-600">{obj.id}</span>
                  <span className="text-gray-500 truncate">{obj.title}</span>
                </div>
              ))}
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => handleDelete(true)}
                disabled={deleting}
                className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-300 text-white rounded-lg text-sm font-medium transition-colors"
              >
                {deleting ? "Deleting..." : "Force Delete Anyway"}
              </button>
              <button
                onClick={onCancel}
                className="flex-1 px-4 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg text-sm transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <div>
            <p className="text-sm text-gray-500 mb-4">
              This object is not referenced by any other objects. It can be safely deleted.
            </p>
            <p className="text-xs text-gray-400 mb-4">
              This action cannot be undone. The object directory and all files will be removed.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => handleDelete(false)}
                disabled={deleting}
                className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-300 text-white rounded-lg text-sm font-medium transition-colors"
              >
                {deleting ? "Deleting..." : "Delete"}
              </button>
              <button
                onClick={onCancel}
                className="flex-1 px-4 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg text-sm transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
