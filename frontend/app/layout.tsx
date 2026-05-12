import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";

import SidebarNav from "../components/sidebar-nav";
import "../styles/globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Fixy Fleet Maintenance",
  description: "Internal fleet maintenance platform for a trucking operation",
};

type RootLayoutProps = {
  children: ReactNode;
};

export default function RootLayout({ children }: RootLayoutProps) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <div className="app-shell">
          <SidebarNav />
          <div className="app-main">{children}</div>
        </div>
      </body>
    </html>
  );
}
