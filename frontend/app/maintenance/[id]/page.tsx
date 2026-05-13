import Link from "next/link";
import { notFound } from "next/navigation";

import { getMaintenanceLog } from "../../../lib/api";
import type { MaintenanceLog } from "../../../lib/types";

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

function formatDateTime(value: string | null) {
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
    hour: "2-digit",
    minute: "2-digit",
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

function formatField(value: string | number | boolean | null) {
  if (value === null || value === "") {
    return "-";
  }

  return String(value);
}

function formatVerification(log: MaintenanceLog) {
  if (log.manager_verified && log.accounting_verified) {
    return "Manager + Accounting";
  }

  if (log.manager_verified) {
    return "Manager";
  }

  if (log.accounting_verified) {
    return "Accounting";
  }

  return "Pending";
}

function InvoiceLink({ url }: { url: string | null }) {
  if (!url) {
    return "-";
  }

  return (
    <a href={url} rel="noreferrer" target="_blank">
      Open invoice
    </a>
  );
}

function DetailField({
  label,
  value,
}: {
  label: string;
  value: React.ReactNode;
}) {
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

type MaintenancePageParams = {
  params: Promise<{
    id: string;
  }>;
};

export default async function MaintenanceDetailPage({
  params,
}: MaintenancePageParams) {
  const { id } = await params;

  let log: MaintenanceLog;

  try {
    log = await getMaintenanceLog(id);
  } catch {
    notFound();
  }

  const unitLabel = formatUnit(log);
  const heroTitle =
    unitLabel === "-" ? "Maintenance entry" : `Unit ${unitLabel}`;

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Transaction detail</p>
          <h1 className="hero-title">{heroTitle}</h1>
          <p className="hero-copy">
            Full maintenance transaction record with payment, verification, and
            source details.
          </p>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Amount</span>
            <strong className="metric-value mono">
              {formatCurrency(log.amount)}
            </strong>
          </div>
          <div className="hero-metric">
            <span className="metric-label">Verification</span>
            <strong className="metric-value mono">
              {formatVerification(log)}
            </strong>
          </div>
        </div>
      </section>

      <div className="detail-grid">
        <DetailCard title="Entry">
          <div className="detail-fields-grid">
            <DetailField label="Date" value={formatDate(log.expense_date)} />
            <DetailField label="Unit" value={unitLabel} />
            <DetailField label="Category" value={log.category} />
            <DetailField label="Driver" value={formatField(log.driver_name)} />
            <DetailField
              label="Description"
              value={formatField(log.description)}
            />
            <DetailField label="Week" value={formatField(log.week_label)} />
          </div>
        </DetailCard>

        <DetailCard title="Payment">
          <div className="detail-fields-grid">
            <DetailField label="Amount" value={formatCurrency(log.amount)} />
            <DetailField
              label="Payment type"
              value={formatField(log.payment_type)}
            />
            <DetailField label="Paid by" value={formatField(log.paid_by)} />
            <DetailField
              label="Who covers"
              value={formatField(log.who_covers)}
            />
            <DetailField
              label="Reference"
              value={formatField(log.reference_number)}
            />
          </div>
        </DetailCard>

        <DetailCard title="Verification">
          <div className="detail-fields-grid">
            <DetailField
              label="Manager verified"
              value={log.manager_verified ? "Yes" : "No"}
            />
            <DetailField
              label="Accounting verified"
              value={log.accounting_verified ? "Yes" : "No"}
            />
            <DetailField
              label="Telegram message"
              value={formatField(log.telegram_message)}
            />
          </div>
        </DetailCard>

        <DetailCard title="Files & system">
          <div className="detail-fields-grid">
            <DetailField
              label="Invoice file"
              value={<InvoiceLink url={log.invoice_file_url} />}
            />
            <DetailField label="Record ID" value={log.id} />
            <DetailField label="Truck ID" value={formatField(log.truck_id)} />
            <DetailField
              label="Trailer ID"
              value={formatField(log.trailer_id)}
            />
            <DetailField
              label="Created"
              value={formatDateTime(log.created_at)}
            />
            <DetailField
              label="Updated"
              value={formatDateTime(log.updated_at)}
            />
          </div>
        </DetailCard>
      </div>

      <section className="panel detail-footer">
        <div className="panel-header">
          <span className="panel-kicker">Navigation</span>
          <Link className="panel-link" href="/maintenance">
            Back to maintenance logs
          </Link>
        </div>
      </section>
    </main>
  );
}
