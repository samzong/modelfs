import type { Phase } from "@api/types";

export default function StatusPill({ phase }: { phase: Phase }) {
  const color = phase === "READY" ? "bg-green-100 text-green-800" : phase === "FAILED" ? "bg-red-100 text-red-800" : phase === "PROCESSING" ? "bg-yellow-100 text-yellow-800" : "bg-gray-100 text-gray-800";
  return <span className={`px-2 py-1 rounded text-sm ${color}`}>{phase}</span>;
}

