// ── Tailwind-safe class name helpers ──
// Tailwind's JIT compiler cannot resolve dynamic class names like `bg-${color}-500`.
// Use full static class name lookups instead.

export const stateBadgeClasses: Record<string, string> = {
  gray: "bg-gray-100 dark:bg-gray-900/30 text-gray-700 dark:text-gray-400",
  amber: "bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400",
  blue: "bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400",
  green: "bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400",
  red: "bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400",
  slate: "bg-slate-100 dark:bg-slate-900/30 text-slate-700 dark:text-slate-400",
};

export const stateDotClasses: Record<string, string> = {
  gray: "bg-gray-400",
  amber: "bg-amber-500",
  blue: "bg-blue-500",
  green: "bg-green-500",
  red: "bg-red-500",
  slate: "bg-slate-500",
};

export const makeBuyClasses: Record<string, string> = {
  make: "bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400",
  buy: "bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-400",
};
