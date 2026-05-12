import Link from "next/link";

import { getMaintenanceLogs } from "../../lib/api";
import type { MaintenanceLog } from "../../lib/types";

function formatDate(value: string) {
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

function formatCurrency(value: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);
}

function formatUnit(log: MaintenanceLog) {
  return log.truck_unit_number ?? log.trailer_unit_number ?? "-";
}

export default async function MaintenanceLogsPage() {
  let logs: MaintenanceLog[] = [];
  let errorMessage = "";

  try {
    logs = await getMaintenanceLogs();
  } catch (error) {
    errorMessage =
      error instanceof Error
        ? error.message
        : "Failed to load maintenance logs";
  }

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Operations log</p>
          <h1 className="hero-title">Maintenance logs</h1>
          <p className="hero-copy">
            A running history of work performed on trucks and trailers.
          </p>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Entries</span>
            <strong className="metric-value mono">{logs.length}</strong>
          </div>
          <div className="hero-metric hero-action-card">
            <Link className="hero-action-button" href="/maintenance/new">
              Add new entry
            </Link>
          </div>
        </div>
      </section>

      {errorMessage ? (
        <section className="panel">
          <h2>Unable to load maintenance logs</h2>
          <p>{errorMessage}</p>
        </section>
      ) : (
        <section className="panel">
          <div className="panel-header">
            <h2>All operations</h2>
            <Link className="panel-link" href="/dashboard">
              Back to dashboard
            </Link>
          </div>

          {logs.length === 0 ? (
            <div className="empty-state">
              <h2>No maintenance logs yet</h2>
              <p>
                Once maintenance is recorded, it will appear here and inside the
                truck detail history section.
              </p>
            </div>
          ) : (
            <div className="table-wrap">
              <table className="dense-table maintenance-table">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Unit</th>
                    <th>Category</th>
                    <th>Description</th>
                    <th>Amount</th>
                    <th>Verified</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => (
                    <tr key={log.id}>
                      <td>{formatDate(log.expense_date)}</td>
                      <td className="mono">{formatUnit(log)}</td>
                      <td>{log.category}</td>
                      <td>{log.description ?? "-"}</td>
                      <td className="mono">{formatCurrency(log.amount)}</td>
                      <td>
                        {log.manager_verified ? "Manager" : "Pending"}
                        {log.accounting_verified ? " / Accounting" : ""}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </main>
  );
}
