export type AppNavItem = {
  label: string;
  href: string;
  icon: string;
  tourId: string;
};

export const appNavItems: AppNavItem[] = [
  {
    label: "Overview",
    href: "/app/dashboard",
    icon: "dashboard",
    tourId: "overview",
  },
  {
    label: "Agent Settings",
    href: "/app/agent-settings",
    icon: "settings_input_component",
    tourId: "agent-settings",
  },
  {
    label: "Portfolio",
    href: "/app/portfolio",
    icon: "pie_chart",
    tourId: "portfolio",
  },
  {
    label: "Activity",
    href: "/app/activity",
    icon: "history",
    tourId: "activity",
  },
];
