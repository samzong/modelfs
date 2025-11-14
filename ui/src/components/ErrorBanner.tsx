import type { ErrorBanner as EB } from "@api/types";

export default function ErrorBanner({ items }: { items?: EB[] }) {
  const list = Array.isArray(items) ? items : [];
  if (list.length === 0) return null;
  return (
    <div className="mb-3">
      {list.map((e, i) => (
        <div key={i} className="border bg-red-50 text-red-700 rounded-lg p-3 mb-2">
          <div className="font-medium">{e.reason}</div>
          <div className="text-sm">{e.message}</div>
        </div>
      ))}
    </div>
  );
}
