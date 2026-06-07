// Standard Parts Center — catalog, search, duplicate detection
import { useState, useEffect } from "react";
import { Search, Nut, AlertTriangle, CheckCircle2, Copy, Plus } from "lucide-react";
import type { ObjectDTO } from "../types";
import api from "../bindings/api";

export default function StandardPartsCenter() {
  const [parts, setParts] = useState<ObjectDTO[]>([]);
  const [filtered, setFiltered] = useState<ObjectDTO[]>([]);
  const [search, setSearch] = useState("");
  const [stdClass, setStdClass] = useState("all");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.listObjects().then((objs) => {
      const std = objs.filter((o) => o.class === "std");
      setParts(std);
      setFiltered(std);
      setLoading(false);
    }).catch(() => setLoading(false));
  }, []);

  useEffect(() => {
    let f = parts;
    if (stdClass !== "all") {
      f = f.filter((p) => p.metadata?.std_class === stdClass);
    }
    if (search.trim()) {
      const q = search.toLowerCase();
      f = f.filter((p) =>
        p.title.toLowerCase().includes(q) ||
        p.id.toLowerCase().includes(q) ||
        JSON.stringify(p.metadata).toLowerCase().includes(q)
      );
    }
    setFiltered(f);
  }, [search, stdClass, parts]);

  const stdClasses = [
    { code: "fst", label: "Fasteners" },
    { code: "brg", label: "Bearings" },
    { code: "elc", label: "Electrical" },
    { code: "pne", label: "Pneumatic" },
    { code: "hyd", label: "Hydraulic" },
    { code: "mat", label: "Materials" },
    { code: "cab", label: "Cables" },
    { code: "mot", label: "Motors" },
    { code: "sen", label: "Sensors" },
    { code: "prof", label: "Profiles" },
    { code: "seal", label: "Seals" },
  ];

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <h2 className="text-lg font-bold flex items-center gap-2">
          <Nut size={20} className="text-amber-500" />
          Standard Parts Center
        </h2>
        <p className="text-sm text-gray-500 mt-1">
          Browse, search, and detect duplicates in the standard parts library
        </p>

        {/* Filters */}
        <div className="flex items-center gap-3 mt-3">
          <div className="relative flex-1 max-w-sm">
            <Search size={16} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search parts..."
              className="w-full pl-8 pr-3 py-1.5 text-sm rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <select
            value={stdClass}
            onChange={(e) => setStdClass(e.target.value)}
            className="px-3 py-1.5 text-sm rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800"
          >
            <option value="all">All Classes</option>
            {stdClasses.map((c) => (
              <option key={c.code} value={c.code}>{c.label} ({c.code})</option>
            ))}
          </select>
        </div>
      </div>

      {/* Parts grid */}
      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-amber-500" />
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-12 text-gray-400">
            <Nut size={48} className="mx-auto mb-4 opacity-50" />
            <p>No standard parts found</p>
            <p className="text-sm mt-1">Create a 'std' class object to add parts to the library</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {filtered.map((part) => (
              <div
                key={part.id}
                className="p-4 rounded-xl border border-gray-200 dark:border-gray-700 hover:border-amber-300 dark:hover:border-amber-700 bg-white dark:bg-gray-800 transition-colors cursor-pointer group"
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Nut size={16} className="text-amber-500" />
                    <span className="font-medium text-sm">{part.title}</span>
                  </div>
                  <button
                    onClick={(e) => { e.stopPropagation(); navigator.clipboard.writeText(part.id); }}
                    className="opacity-0 group-hover:opacity-100 text-gray-400 hover:text-gray-600"
                    title="Copy ID"
                  >
                    <Copy size={14} />
                  </button>
                </div>

                <div className="space-y-1 text-xs">
                  <div className="font-mono text-blue-600 dark:text-blue-400">{part.id}</div>
                  {part.metadata?.std_class != null && (
                    <div className="text-gray-500">Class: <span className="font-medium">{String(part.metadata.std_class)}</span></div>
                  )}
                  {part.metadata?.standard != null && (
                    <div className="text-gray-500">Standard: <span className="font-medium">{String(part.metadata.standard)}</span></div>
                  )}
                  {part.metadata?.thread != null && (
                    <div className="text-gray-500">Thread: <span className="font-medium">{String(part.metadata.thread)}</span></div>
                  )}
                  {part.metadata?.material_grade != null && (
                    <div className="text-gray-500">Grade: <span className="font-medium">{String(part.metadata.material_grade)}</span></div>
                  )}
                </div>

                {/* Status badge */}
                <div className="flex items-center gap-2 mt-2 pt-2 border-t border-gray-100 dark:border-gray-700">
                  <span className={`inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full ${
                    part.state === "released" ? "bg-green-100 text-green-700" :
                    part.state === "approved" ? "bg-blue-100 text-blue-700" :
                    part.state === "draft" ? "bg-gray-100 text-gray-700" :
                    "bg-gray-100 text-gray-700"
                  }`}>
                    {part.state}
                  </span>
                  {part.metadata?.make_buy != null && (
                    <span className={`text-xs px-2 py-0.5 rounded-full ${
                      String(part.metadata.make_buy) === "buy" ? "bg-orange-100 text-orange-700" : "bg-green-100 text-green-700"
                    }`}>
                      {String(part.metadata.make_buy)}
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Stats */}
      <div className="px-4 py-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-400 flex justify-between">
        <span>{filtered.length} parts</span>
        <span>{parts.length} total</span>
      </div>
    </div>
  );
}
