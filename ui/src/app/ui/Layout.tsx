import { Link, Outlet } from "@tanstack/react-router";
import NamespaceSelector from "@components/NamespaceSelector";
import "../../index.css";

export default function Layout() {
  return (
    <div className="min-h-screen flex bg-muted">
      <aside className="w-64 p-4 bg-white border-r">
        <div className="text-2xl font-semibold mb-4">modelfs</div>
        <nav className="flex flex-col gap-1">
          <Link to="/models" className="px-3 py-2 rounded-lg hover:bg-muted">Models</Link>
          <Link to="/modelsources" className="px-3 py-2 rounded-lg hover:bg-muted">ModelSources</Link>
        </nav>
      </aside>
      <main className="flex-1">
        <header className="flex items-center border-b bg-white p-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">Namespace</span>
            <NamespaceSelector />
          </div>
        </header>
        <section className="page-container">
          <Outlet />
        </section>
      </main>
    </div>
  );
}
