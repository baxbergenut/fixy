import Link from "next/link";

import MaintenanceEntryForm from "../../../components/maintenance/maintenance-entry-form";
import { getTrucks } from "../../../lib/api";
import type { Truck } from "../../../lib/types";

export default async function NewMaintenanceEntryPage() {
  let trucks: Truck[] = [];
  let errorMessage = "";

  try {
    trucks = await getTrucks();
  } catch (error) {
    errorMessage =
      error instanceof Error ? error.message : "Failed to load trucks";
  }

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Operations entry</p>
          <h1 className="hero-title">Add maintenance entry</h1>
          <p className="hero-copy">
            Upload an invoice, let Groq prefill the fields, then confirm and
            save the log.
          </p>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Trucks loaded</span>
            <strong className="metric-value mono">{trucks.length}</strong>
          </div>
          <div className="hero-metric hero-action-card">
            <Link className="hero-action-button" href="/maintenance">
              Back to logs
            </Link>
          </div>
        </div>
      </section>

      {errorMessage ? (
        <section className="panel">
          <h2>Unable to load trucks</h2>
          <p>{errorMessage}</p>
        </section>
      ) : (
        <MaintenanceEntryForm trucks={trucks} />
      )}
    </main>
  );
}
