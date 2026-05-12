import Link from "next/link";
import { notFound } from "next/navigation";

import { getMaintenanceLogs, getTruck, getTrucks } from "../../../lib/api";
import type { MaintenanceLog, Truck, TruckStatus } from "../../../lib/types";

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

function formatField(value: string | number | boolean | null) {
  if (value === null || value === "") {
    return "-";
  }

  return String(value);
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);
}

function formatTruckLabel(log: MaintenanceLog) {
  return log.truck_unit_number ?? log.trailer_unit_number ?? "-";
}

function DetailField({ label, value }: { label: string; value: string }) {
  return (
    <div className="detail-field">
      <span className="detail-label">{label}</span>
      <strong className="detail-value">{value}</strong>
    </div>
  );
}

function DetailCard({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="detail-card panel">
      <div className="panel-header">
        <h2>{title}</h2>
      </div>
      {children}
    </section>
  );
}

type TruckPageParams = {
  params: Promise<{
    id: string;
  }>;
};

export async function generateStaticParams() {
  try {
    const trucks = await getTrucks();
    return trucks.map((truck) => ({ id: truck.id }));
  } catch {
    return [];
  }
}

export default async function TruckDetailPage({ params }: TruckPageParams) {
  const { id } = await params;

  let truck: Truck;
  let history: MaintenanceLog[] = [];

  try {
    truck = await getTruck(id);
  } catch {
    notFound();
  }

  try {
    history = await getMaintenanceLogs(truck.id);
  } catch {
    history = [];
  }

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Truck detail</p>
          <h1 className="hero-title">Unit {truck.unit_number}</h1>
          <p className="hero-copy">
            Full truck record with identity, registration, and maintenance
            flags.
          </p>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Status</span>
            <strong className="metric-value">
              <span className={`status-pill ${statusTone(truck.status)}`}>
                {truck.status}
              </span>
            </strong>
          </div>
          <div className="hero-metric">
            <span className="metric-label">Active</span>
            <strong className="metric-value mono">
              {truck.active ? "Yes" : "No"}
            </strong>
          </div>
        </div>
      </section>

      <div className="detail-grid">
        <DetailCard title="Identity">
          <div className="detail-fields-grid">
            <DetailField label="Unit number" value={truck.unit_number} />
            <DetailField label="Make / year" value={formatMakeYear(truck)} />
            <DetailField label="VIN" value={truck.vin ?? "-"} />
            <DetailField label="Company" value={truck.company ?? "-"} />
            <DetailField label="Ownership" value={truck.ownership ?? "-"} />
            <DetailField label="Samsara ID" value={truck.samsara_id ?? "-"} />
          </div>
        </DetailCard>

        <DetailCard title="Registration">
          <div className="detail-fields-grid">
            <DetailField
              label="Plate number"
              value={truck.plate_number ?? "-"}
            />
            <DetailField label="Plate state" value={truck.plate_state ?? "-"} />
            <DetailField
              label="DOT expires"
              value={formatDate(truck.dot_inspection_expires_at)}
            />
            <DetailField
              label="DOT form"
              value={truck.dot_inspection_form_url ?? "-"}
            />
            <DetailField label="Created" value={formatDate(truck.created_at)} />
            <DetailField label="Updated" value={formatDate(truck.updated_at)} />
          </div>
        </DetailCard>

        <DetailCard title="Maintenance markers">
          <div className="detail-fields-grid">
            <DetailField
              label="Next PM odometer"
              value={formatField(truck.next_pm_odometer)}
            />
            <DetailField
              label="Next oil change"
              value={formatField(truck.next_oil_change_odometer)}
            />
            <DetailField
              label="Status changed"
              value={formatDate(truck.status_changed_at)}
            />
            <DetailField label="Status note" value={truck.status_note ?? "-"} />
            <DetailField label="Notes" value={truck.notes ?? "-"} />
            <DetailField label="Active" value={truck.active ? "Yes" : "No"} />
          </div>
        </DetailCard>
      </div>

      <section className="panel detail-footer">
        <div className="panel-header">
          <span className="panel-kicker">Navigation</span>
          <Link className="panel-link" href="/trucks">
            Back to all trucks
          </Link>
        </div>
      </section>

      <section className="panel detail-history">
        <div className="panel-header">
          <div>
            <span className="panel-kicker">Operations history</span>
            <h2>Maintenance activity on this truck</h2>
          </div>
          <Link className="panel-link" href="/maintenance">
            View all logs
          </Link>
        </div>

        {history.length === 0 ? (
          <div className="empty-state">
            <h2>No operations logged yet</h2>
            <p>
              Maintenance records for unit {truck.unit_number} will show up here
              once they are added.
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
                  <th>Who covers</th>
                </tr>
              </thead>
              <tbody>
                {history.map((log) => (
                  <tr key={log.id}>
                    <td>{formatDate(log.expense_date)}</td>
                    <td className="mono">{formatTruckLabel(log)}</td>
                    <td>{log.category}</td>
                    <td>{log.description ?? "-"}</td>
                    <td className="mono">{formatCurrency(log.amount)}</td>
                    <td>{log.who_covers ?? "-"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </main>
  );
}
