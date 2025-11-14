import { createRootRoute, createRoute, createRouter, RouterProvider } from "@tanstack/react-router";
import Layout from "@app/ui/Layout";
import ModelsPage from "@pages/Models";
import ModelDetailPage from "@pages/ModelDetail";
import ModelSourcesPage from "@pages/ModelSources";
import ModelWizardPage from "@pages/ModelWizard";
import ModelSourceWizardPage from "@pages/ModelSourceWizard";
import ModelSourceDetailPage from "@pages/ModelSourceDetail";

const rootRoute = createRootRoute({ component: () => <Layout /> });
const modelsRoute = createRoute({ getParentRoute: () => rootRoute, path: "/models", component: () => <ModelsPage /> });
const modelDetailRoute = createRoute({ getParentRoute: () => rootRoute, path: "/models/$ns/$name", component: () => <ModelDetailPage /> });
const sourcesRoute = createRoute({ getParentRoute: () => rootRoute, path: "/modelsources", component: () => <ModelSourcesPage /> });
const wizardRoute = createRoute({ getParentRoute: () => rootRoute, path: "/models/wizard", component: () => <ModelWizardPage /> });
const wizardEditRoute = createRoute({ getParentRoute: () => rootRoute, path: "/models/$ns/$name/edit", component: () => <ModelWizardPage /> });
const msWizardRoute = createRoute({ getParentRoute: () => rootRoute, path: "/modelsources/new", component: () => <ModelSourceWizardPage /> });
const msDetailRoute = createRoute({ getParentRoute: () => rootRoute, path: "/modelsources/$ns/$name", component: () => <ModelSourceDetailPage /> });
const msEditRoute = createRoute({ getParentRoute: () => rootRoute, path: "/modelsources/$ns/$name/edit", component: () => <ModelSourceWizardPage /> });

const routeTree = rootRoute.addChildren([modelsRoute, modelDetailRoute, sourcesRoute, wizardRoute, wizardEditRoute, msWizardRoute, msDetailRoute, msEditRoute]);
const router = createRouter({ routeTree });

export default function AppRouter() {
  return <RouterProvider router={router} />;
}
