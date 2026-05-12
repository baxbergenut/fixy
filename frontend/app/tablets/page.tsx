import Link from "next/link";

import TabletManager from "../../components/tablets/tablet-manager";
import { getTablets, getTrucks } from "../../lib/api";
import type { Tablet, Truck } from "../../lib/types";

export default async function TabletsPage() {
  let tablets: Tablet[] = [];
  let trucks: Truck[] = [];
  let errorMessage = "";

  try {
    [tablets, trucks] = await Promise.all([getTablets(), getTrucks()]);
  } catch (error) {
    errorMessage =
      error instanceof Error ? error.message : "Failed to load tablets";
  }

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Device inventory</p>
          <h1 className="hero-title">Tablets</h1>
          <p className="hero-copy">
            Track tablet assignments, contract dates, and device details by
            truck.
          </p>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Tablets</span>
            <strong className="metric-value mono">{tablets.length}</strong>
          </div>
          <div className="hero-metric">
            <span className="metric-label">Assigned</span>
            <strong className="metric-value mono">
              {tablets.filter((item) => item.truck_id).length}
            </strong>
          </div>
        </div>
      </section>

      {errorMessage ? (
        <section className="panel">
          <h2>Unable to load tablets</h2>
          <p>{errorMessage}</p>
        </section>
      ) : (
        <>
          <TabletManager tablets={tablets} trucks={trucks} />
          <section className="panel detail-footer">
            <div className="panel-header">
              <span className="panel-kicker">Navigation</span>
              <Link className="panel-link" href="/dashboard">
                Back to dashboard
              </Link>
            </div>
          </section>
        </>
      )}
    </main>
  );
}
