export type AppNavItem = {
  label: string;
  href: string;
  icon: string;
};

export const appNavItems: AppNavItem[] = [
  { label: "Overview", href: "/app/dashboard", icon: "dashboard" },
  {
    label: "Agent Settings",
    href: "/app/agent-settings",
    icon: "settings_input_component",
  },
  { label: "Portfolio", href: "/app/portfolio", icon: "pie_chart" },
  { label: "Activity", href: "/app/activity", icon: "history" },
];
