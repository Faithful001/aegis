import React from "react";
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
  useParams,
  useLocation,
} from "@tanstack/react-router";

// Import route components
import HomePage from "../pages/index";
import LoginPage from "../pages/auth/login/index";
import RegisterPage from "../pages/auth/register/index";
import ForgotPasswordPage from "../pages/auth/forgot-password/index";
import VerifyEmailPage from "../pages/auth/verify-email/index";
import VerifyOTPPage from "../pages/auth/verify-otp/index";
import ResetPasswordPage from "../pages/auth/reset-password/index";
import ChatPage from "../pages/chat/index";
import DynamicChatThreadPage from "../pages/chat/[id]/index";
import ProfilePage from "../pages/profile/index";
import AnalyticsPage from "../pages/analytics/index";
import SettingsPage, { SettingsTab } from "../pages/settings/index";

import { MainLayout } from "../components/layout/MainLayout";

// Root Route with persistent MainLayout for all app pages
const rootRoute = createRootRoute({
  component: () => {
    const { pathname } = useLocation();
    const isAuth = pathname.startsWith("/auth");

    if (isAuth) {
      return <Outlet />;
    }

    return <MainLayout />;
  },
});

// Index Route (/)
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: HomePage,
});

// Auth Routes
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/login",
  component: LoginPage,
});

const registerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/register",
  component: RegisterPage,
});

const forgotPasswordRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/forgot-password",
  component: ForgotPasswordPage,
});

const verifyEmailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/verify-email",
  component: VerifyEmailPage,
});

const verifyOTPRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/verify-otp",
  component: VerifyOTPPage,
});

const resetPasswordRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/auth/reset-password",
  component: ResetPasswordPage,
});

// Chat Routes
const chatRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/chat",
  component: ChatPage,
});

const chatDynamicRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/chat/$id",
  component: () => {
    const { id } = useParams({ from: "/chat/$id" });
    return <DynamicChatThreadPage id={id} />;
  },
});

// Thread Route (/c/:id)
const threadRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/c/$id",
  component: () => {
    const { id } = useParams({ from: "/c/$id" });
    return <DynamicChatThreadPage id={id} />;
  },
});

// Profile Route (/profile)
const profileRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/profile",
  component: ProfilePage,
});

// Analytics Route
const analyticsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/analytics",
  component: AnalyticsPage,
});

export type SettingsSearch = {
  tab?: SettingsTab;
};

// Settings & Vault Route
const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings",
  validateSearch: (search: Record<string, unknown>): SettingsSearch => {
    return {
      tab: (search.tab as SettingsTab) || "profile",
    };
  },
  component: SettingsPage,
});

// Route Tree Assembly
const routeTree = rootRoute.addChildren([
  indexRoute,
  chatRoute,
  chatDynamicRoute,
  threadRoute,
  profileRoute,
  analyticsRoute,
  settingsRoute,
  loginRoute,
  registerRoute,
  forgotPasswordRoute,
  verifyEmailRoute,
  verifyOTPRoute,
  resetPasswordRoute,
]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

export const AppRouter: React.FC = () => {
  return <RouterProvider router={router} />;
};
