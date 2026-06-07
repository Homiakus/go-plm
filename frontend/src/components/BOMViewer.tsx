import { useState, useEffect } from "react";
import type { BOMRowDTO } from "../types";
import { Table, AlertTriangle, CheckCircle, ListTree, List, Download } from "lucide-react";
import { makeBuyClasses } from "../utils";
import api from "../bindings/api";

interface BOMViewerProps {
  rootId: string;
  onSelect: (id: string) => void;
}

export default function BOMViewer({ rootId, onSelect }: BOMViewerProps) {
  const [rows, setRows] = useState<BOMRowDTO[]>([]);
  const [flat, setFlat] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadBOM();
  }, [rootId, flat]);

  const loadBOM = async () => {
    try {
      setLoading(true);
      setError(null);
      const resp = await api.getBOM({ root_id: rootId, flat });
      setRows(resp.rows);
    } catch (err: any) {
      setError(err.message || "Failed to load BOM");
    } finally {
      setLoading(false);
    }
  };

  const totalQty = rows
    .filter((r) => r.child_class === "prt")
    .reduce((sum, r) => sum + r.quantity, 0);
  const uniqueParts = new Set(
    rows.filter((r) => r.child_class === "prt").map((r) => r.child_id)
  ).size;

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-lg font-bold flex items-center gap-2">
              <Table size={20} className="text-purple-500" />
              Bill of Materials
            </h2>
            <p className="text-sm text-gray-500 mt-1">
              Root: <code className="text-xs">{rootId}</code>
            </p>
          </div>
          <div className="flex items-center gap-3">
            {/* Flat/Structured toggle */}
            <div className="flex bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
              <button
                onClick={() => setFlat(false)}
                className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
                  !flat
                    ? "bg-white dark:bg-gray-700 shadow-sm text-purple-600"
                    : "text-gray-500"
                }`}
              >
                <ListTree size={14} className="inline mr-1" />
                Structured
              </button>
              <button
                onClick={() => setFlat(true)}
                className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
                  flat
                    ? "bg-white dark:bg-gray-700 shadow-sm text-purple-600"
                    : "text-gray-500"
                }`}
              >
                <List size={14} className="inline mr-1" />
                Flat
              </button>
            </div>

            {/* Summary */}
            <div className="flex items-center gap-4 text-xs text-gray-500">
              <span>{rows.length} rows</span>
              <span>{uniqueParts} unique parts</span>
              {!flat && <span>Qty: {totalQty.toFixed(0)}</span>}
            </div>

            {/* Export button */}
            <button
              onClick={async () => {
                try {
                  const csv = await api.exportBOM(rootId, "csv");
                  const blob = new Blob([csv], { type: "text/csv" });
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement("a");
                  a.href = url; a.download = `bom-${rootId}.csv`;
                  a.click(); URL.revokeObjectURL(url);
                } catch (e: any) {
                  setError(e.message);
                }
              }}
              className="flex items-center gap-1 px-3 py-1.5 text-xs font-medium bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors"
            >
              <Download size={14} />
              Export CSV
            </button>
          </div>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/30 px-6 py-2 text-sm text-red-700 dark:text-red-400">
          {error}
        </div>
      )}

      {/* Loading */}
      {loading ? (
        <div className="flex-1 flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-purple-500" />
        </div>
      ) : rows.length === 0 ? (
        <div className="flex-1 flex items-center justify-center text-gray-400">
          <div className="text-center">
            <Table size={48} className="mx-auto mb-4 opacity-50" />
            <p>No BOM rows found</p>
            <p className="text-sm mt-1">
              Add 'contains' relations to build the BOM structure
            </p>
          </div>
        </div>
      ) : (
        /* BOM Table */
        <div className="flex-1 overflow-auto">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-gray-50 dark:bg-gray-950">
              <tr className="border-b border-gray-200 dark:border-gray-700">
                {!flat && (
                  <th className="text-left py-2 px-4 font-medium text-gray-500 w-20">
                    Level
                  </th>
                )}
                <th className="text-left py-2 px-4 font-medium text-gray-500">
                  Position
                </th>
                <th className="text-left py-2 px-4 font-medium text-gray-500">
                  Part ID
                </th>
                <th className="text-left py-2 px-4 font-medium text-gray-500">
                  Class
                </th>
                <th className="text-right py-2 px-4 font-medium text-gray-500">
                  Qty
                </th>
                <th className="text-left py-2 px-4 font-medium text-gray-500">
                  Unit
                </th>
                <th className="text-left py-2 px-4 font-medium text-gray-500">
                  Make/Buy
                </th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr
                  key={row.row_id}
                  onClick={() => onSelect(row.child_id)}
                  className="border-b border-gray-100 dark:border-gray-800 hover:bg-blue-50 dark:hover:bg-blue-900/10 cursor-pointer transition-colors"
                >
                  {!flat && (
                    <td className="py-2 px-4">
                      <span
                        className="inline-block ml-0"
                        style={{ marginLeft: `${row.level * 16}px` }}
                      >
                        {row.level === 0 ? "—" : `└ L${row.level}`}
                      </span>
                    </td>
                  )}
                  <td className="py-2 px-4 font-mono text-xs text-gray-400">
                    {row.position || "—"}
                  </td>
                  <td className="py-2 px-4 font-mono text-xs text-blue-600 dark:text-blue-400 hover:underline">
                    {row.child_id}
                  </td>
                  <td className="py-2 px-4">
                    <span className="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800">
                      {row.child_class}
                    </span>
                  </td>
                  <td className="py-2 px-4 text-right font-medium">
                    {row.quantity}
                  </td>
                  <td className="py-2 px-4 text-gray-500">{row.unit || "—"}</td>
                  <td className="py-2 px-4">
                    {row.make_buy ? (
                      <span
                        className={`text-xs px-1.5 py-0.5 rounded ${makeBuyClasses[row.make_buy] || "bg-gray-100 dark:bg-gray-900/30 text-gray-700"}`}
                      >
                        {row.make_buy}
                      </span>
                    ) : (
                      <span className="text-gray-300">—</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
