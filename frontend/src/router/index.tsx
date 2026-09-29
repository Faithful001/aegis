import React from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
  useParams,
} from '@tanstack/react-router';

// Import route components
import HomePage from '../pages/index';
import LoginPage from '../pages/auth/login/index';
import RegisterPage from '../pages/auth/register/index';
import ChatPage from '../pages/chat/index';
import DynamicChatThreadPage from '../pages/chat/[id]/index';
import ProfilePage from '../pages/profile/index';
import AnalyticsPage from '../pages/analytics/index';
import SettingsPage from '../pages/settings/index';

// Root Route
const rootRoute = createRootRoute({
  component: () => <Outlet />,
});

// Index Route (/)
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
});

// Auth Routes
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/auth/login',
  component: LoginPage,
});

const registerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/auth/register',
  component: RegisterPage,
});

// Chat Routes
const chatRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/chat',
  component: ChatPage,
});

const chatDynamicRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/chat/$id',
  component: () => {
    const { id } = useParams({ from: '/chat/$id' });
    return <DynamicChatThreadPage id={id} />;
  },
});

// Profile Route (/profile)
const profileRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/profile',
  component: ProfilePage,
});

// Analytics Route
const analyticsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/analytics',
  component: AnalyticsPage,
});

// Settings & Vault Route
const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: SettingsPage,
});

// Route Tree Assembly
const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  registerRoute,
  chatRoute,
  chatDynamicRoute,
  profileRoute,
  analyticsRoute,
  settingsRoute,
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

export const AppRouter: React.FC = () => {
  return <RouterProvider router={router} />;
};
