import Link from "next/link";

import { getTrucks } from "../../lib/api";
import type { Truck, TruckStatus } from "../../lib/types";

function statusTone(status: TruckStatus) {
  switch (status) {
    case "ENROUTE":
      return "green";
    case "SHOP":
      return "red";
    case "STOP":
      return "muted";
    default:
      return "amber";
  }
}

function formatDate(value: string | null) {
  if (!value) {
    return "-";
  }

  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(parsed);
}

function formatMakeYear(truck: Truck) {
  const parts = [truck.make, truck.year].filter(Boolean);
  return parts.length > 0 ? parts.join(" ") : "-";
}

function formatOdometer(value: number | null) {
  if (value === null) {
    return "-";
  }

  return new Intl.NumberFormat("en-US").format(value);
}

export default async function TrucksPage() {
  let trucks: Truck[] = [];
  let errorMessage = "";

  try {
    trucks = await getTrucks();
  } catch (error) {
    errorMessage =
      error instanceof Error ? error.message : "Failed to load trucks";
  }

  const activeCount = trucks.filter((truck) => truck.active).length;
  const shopCount = trucks.filter((truck) => truck.status === "SHOP").length;
  const unavailableCount = trucks.filter(
    (truck) => truck.status === "UNAVAILABLE",
  ).length;

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Fleet registry</p>
          <h1 className="hero-title">Trucks</h1>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Active</span>
            <strong className="metric-value mono">{activeCount}</strong>
          </div>
          <div className="hero-metric">
            <span className="metric-label">In shop</span>
            <strong className="metric-value mono">{shopCount}</strong>
          </div>
          <div className="hero-metric">
            <span className="metric-label">Unavailable</span>
            <strong className="metric-value mono">{unavailableCount}</strong>
          </div>
        </div>
      </section>

      {errorMessage ? (
        <section className="panel">
          <h2>Unable to load trucks</h2>
          <p>{errorMessage}</p>
        </section>
      ) : (
        <section className="panel">
          <div className="panel-header">
            <h2>All trucks</h2>
            <Link className="panel-link" href="/dashboard">
              Back to dashboard
            </Link>
          </div>

          <div className="table-wrap">
            <table className="dense-table trucks-table">
              <thead>
                <tr>
                  <th>Unit</th>
                  <th>Make / year</th>
                  <th>Company</th>
                  <th>Status</th>
                  <th>DOT expires</th>
                  <th>PM target</th>
                </tr>
              </thead>
              <tbody>
                {trucks.map((truck) => (
                  <tr key={truck.id} className="truck-row">
                    <td className="mono">
                      <Link className="truck-link" href={`/trucks/${truck.id}`}>
                        {truck.unit_number}
                      </Link>
                    </td>
                    <td>{formatMakeYear(truck)}</td>
                    <td>{truck.company ?? "-"}</td>
                    <td>
                      <span
                        className={`status-pill ${statusTone(truck.status)}`}
                      >
                        {truck.status}
                      </span>
                    </td>
                    <td>{formatDate(truck.dot_inspection_expires_at)}</td>
                    <td className="mono">
                      {formatOdometer(truck.next_pm_odometer)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
    </main>
  );
}
