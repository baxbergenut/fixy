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
    href: "/trailers",
    label: "Trailers",
    description: "Trailer assignments",
  },
  {
    href: "/transponders",
    label: "Transponders",
    description: "Toll device assignments",
  },
  {
    href: "/tablets",
    label: "Tablets",
    description: "Device assignments",
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
    </aside>
  );
}
