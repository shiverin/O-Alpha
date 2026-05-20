import type { Metadata } from "next";
// Ignore missing type declarations for CSS side-effect import
// @ts-ignore
import "./globals.css";

export const metadata: Metadata = {
  title: "O(Alpha) | Autonomous PMS",
  description: "Algorithmic portfolio management and backtesting",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
