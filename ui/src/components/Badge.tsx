import type { Phase } from "@api/types";

export default function Badge({ phase }: { phase: Phase }) {
  const map: Record<Phase, string> = {
    READY: "bg-green-100 text-green-800",
    FAILED: "bg-red-100 text-red-800",
    PROCESSING: "bg-yellow-100 text-yellow-800",
    PENDING: "bg-gray-100 text-gray-800",
    UNKNOWN: "bg-gray-100 text-gray-800",
  };
  return <span className={`px-2 py-1 rounded-lg text-sm ${map[phase]}`}>{phase}</span>;
}

