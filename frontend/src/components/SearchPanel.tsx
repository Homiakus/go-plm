import { useState, useEffect } from "react";
import type { SearchResult } from "../types";
import { Search, FileText, X } from "lucide-react";
import api from "../bindings/api";
import { CLASS_LABELS, STATE_LABELS, STATE_COLORS } from "../types";
import { stateBadgeClasses } from "../utils";

interface SearchPanelProps {
  query: string;
  onSelect: (id: string) => void;
  onClose: () => void;
}

export default function SearchPanel({ query, onSelect, onClose }: SearchPanelProps) {
  const [searchText, setSearchText] = useState(query);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  useEffect(() => {
    if (query) {
      setSearchText(query);
      doSearch(query);
    }
  }, [query]);

  const doSearch = async (q: string) => {
    if (!q.trim()) return;
    try {
      setLoading(true);
      const res = await api.searchObjects({ query: q, limit: 50 });
      setResults(res);
      setSearched(true);
    } catch (err) {
      console.error("Search failed:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      doSearch(searchText);
    }
  };

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Search bar */}
      <div className="border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="relative flex-1">
            <Search
              size={18}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
            />
            <input
              type="text"
              value={searchText}
              onChange={(e) => setSearchText(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Search objects (FTS5 full-text search)..."
              className="w-full pl-10 pr-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500 text-lg"
              autoFocus
            />
          </div>
          <button
            onClick={() => doSearch(searchText)}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
          >
            Search
          </button>
          <button
            onClick={onClose}
            className="p-2 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800"
          >
            <X size={20} />
          </button>
        </div>
        <p className="text-xs text-gray-400 mt-2">
          FTS5 full-text search across ID, title, class, state, metadata, and body content.
        </p>
      </div>

      {/* Results */}
      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
          </div>
        ) : !searched ? (
          <div className="flex items-center justify-center py-20 text-gray-400">
            <div className="text-center">
              <Search size={48} className="mx-auto mb-4 opacity-50" />
              <p>Enter a search query to find objects</p>
            </div>
          </div>
        ) : results.length === 0 ? (
          <div className="flex items-center justify-center py-20 text-gray-400">
            <div className="text-center">
              <FileText size={48} className="mx-auto mb-4 opacity-50" />
              <p>No results found for "{searchText}"</p>
            </div>
          </div>
        ) : (
          <div>
            <div className="px-6 py-2 text-sm text-gray-500 border-b border-gray-100 dark:border-gray-800">
              {results.length} result{results.length !== 1 ? "s" : ""}
            </div>
            {results.map((r) => {
              const stateColor = STATE_COLORS[r.state] || "gray";
              return (
                <div
                  key={r.object_id}
                  onClick={() => onSelect(r.object_id)}
                  className="px-6 py-3 border-b border-gray-100 dark:border-gray-800 hover:bg-blue-50 dark:hover:bg-blue-900/10 cursor-pointer transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <FileText size={16} className="text-gray-400 shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-sm truncate">
                        {r.title}
                      </div>
                      <div className="flex items-center gap-2 mt-0.5">
                        <code className="text-xs text-gray-400 font-mono">
                          {r.object_id}
                        </code>
                        <span className="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800">
                          {CLASS_LABELS[r.class] || r.class}
                        </span>
                        <span
                          className={`text-xs px-1.5 py-0.5 rounded ${stateBadgeClasses[stateColor] || stateBadgeClasses.gray}`}
                        >
                          {STATE_LABELS[r.state] || r.state}
                        </span>
                      </div>
                      {r.excerpt && (
                        <p className="text-xs text-gray-500 mt-1 truncate">
                          {r.excerpt}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
