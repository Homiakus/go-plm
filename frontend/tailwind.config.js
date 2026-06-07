/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        plm: {
          draft: "#9ca3af",
          review: "#f59e0b",
          approved: "#3b82f6",
          released: "#22c55e",
          blocked: "#ef4444",
          obsolete: "#64748b",
          archived: "#475569",
        },
      },
    },
  },
  plugins: [],
};
