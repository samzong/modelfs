export type SSEHandler = (resource: string, action: string, payload: any) => void;

export function attachSSE(ns: string, onEvent: SSEHandler) {
  if (typeof window === "undefined" || !(window as any).EventSource) return () => {};
  const es = new EventSource(`/api/sse?namespace=${encodeURIComponent(ns)}`);
  es.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data);
      onEvent(data.resource, data.action, data.payload);
    } catch {}
  };
  ["added", "modified", "deleted"].forEach((evt) => {
    es.addEventListener(evt, (e: any) => {
      try {
        const data = JSON.parse(e.data);
        onEvent(data.resource, data.action, data.payload);
      } catch {}
    });
  });
  return () => es.close();
}
