"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const navigationItems = [
  {
    href: "/dashboard",
    label: "Dashboard",
    description: "Fleet overview",
  },
  {
    href: "/trucks",
    label: "Trucks",
    description: "Registry and unit details",
  },
  {
    href: "/maintenance",
    label: "Maintenance logs",
    description: "Operations history",
  },
] as const;

export default function SidebarNav() {
  const pathname = usePathname();

  return (
    <aside className="app-sidebar">
      <div className="sidebar-brand">
        <p className="eyebrow">Fixy Fleet</p>
        <h1>Fleet board</h1>
        <p>Static navigation for the core operations views.</p>
      </div>

      <nav className="sidebar-nav" aria-label="Primary">
        {navigationItems.map((item) => {
          const isActive =
            pathname === item.href || pathname.startsWith(`${item.href}/`);

          return (
            <Link
              key={item.href}
              aria-current={isActive ? "page" : undefined}
              className={`sidebar-link${isActive ? " active" : ""}`}
              href={item.href}
            >
              <span>{item.description}</span>
              <strong>{item.label}</strong>
            </Link>
          );
        })}
      </nav>

      <div className="sidebar-footer">
        <p className="sidebar-footer-label">Live sections</p>
        <p className="sidebar-footer-copy">
          Dashboard, trucks, and maintenance logs share this shell.
        </p>
      </div>
    </aside>
  );
}
