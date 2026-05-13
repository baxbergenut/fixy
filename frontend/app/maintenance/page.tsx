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

function TelegramLinkAction({ message }: { message: string | null }) {
  if (!message) {
    return <span className="maintenance-telegram-placeholder">-</span>;
  }

  return (
    <span
      className="telegram-link"
      aria-label="Telegram message"
      title={message}
    >
      <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
        <path d="M21.8 4.2 18.4 20c-.2 1-.8 1.3-1.7.8l-4.7-3.5-2.3 2.2c-.3.3-.6.5-1 .5l.3-5.1 9.3-8.4c.4-.4-.1-.6-.7-.2L6.2 12.4 1.2 10.8c-1-.3-1-1 .2-1.4L20.5 3c.9-.3 1.7.2 1.3 1.2Z" />
      </svg>
    </span>
  );
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
                    <th className="maintenance-description-header">
                      Description
                    </th>
                    <th>Amount</th>
                    <th>Payment</th>
                    <th>Paid by</th>
                    <th>Reference</th>
                    <th>Verified</th>
                    <th>tg</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => (
                    <tr key={log.id}>
                      <td>
                        <Link
                          className="truck-link"
                          href={`/maintenance/${log.id}`}
                        >
                          {formatDate(log.expense_date)}
                        </Link>
                      </td>
                      <td className="mono">{formatUnit(log)}</td>
                      <td>{log.category}</td>
                      <td className="maintenance-description">
                        {log.description ?? "-"}
                      </td>
                      <td className="mono">{formatCurrency(log.amount)}</td>
                      <td>{log.payment_type ?? "-"}</td>
                      <td>{log.paid_by ?? "-"}</td>
                      <td>{log.reference_number ?? "-"}</td>
                      <td>
                        {log.manager_verified ? "Manager" : "Pending"}
                        {log.accounting_verified ? " / Accounting" : ""}
                      </td>
                      <td className="maintenance-telegram-cell">
                        <TelegramLinkAction message={log.telegram_message} />
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
