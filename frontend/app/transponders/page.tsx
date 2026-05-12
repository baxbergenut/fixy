import Link from "next/link";

import TransponderManager from "../../components/transponders/transponder-manager";
import { getTransponders, getTrucks } from "../../lib/api";
import type { Transponder, Truck } from "../../lib/types";

export default async function TranspondersPage() {
  let transponders: Transponder[] = [];
  let trucks: Truck[] = [];
  let errorMessage = "";

  try {
    [transponders, trucks] = await Promise.all([
      getTransponders(),
      getTrucks(),
    ]);
  } catch (error) {
    errorMessage =
      error instanceof Error ? error.message : "Failed to load transponders";
  }

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Toll devices</p>
          <h1 className="hero-title">Transponders</h1>
          <p className="hero-copy">
            Assign transponders to trucks and keep the active device list
            current.
          </p>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Transponders</span>
            <strong className="metric-value mono">{transponders.length}</strong>
          </div>
          <div className="hero-metric">
            <span className="metric-label">Assigned</span>
            <strong className="metric-value mono">
              {transponders.filter((item) => item.truck_id).length}
            </strong>
          </div>
        </div>
      </section>

      {errorMessage ? (
        <section className="panel">
          <h2>Unable to load transponders</h2>
          <p>{errorMessage}</p>
        </section>
      ) : (
        <>
          <TransponderManager transponders={transponders} trucks={trucks} />
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
