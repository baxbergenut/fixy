import Link from "next/link";

import TrailerManager from "../../components/trailers/trailer-manager";
import { getTrailers } from "../../lib/api";
import type { Trailer } from "../../lib/types";

export default async function TrailersPage() {
  let trailers: Trailer[] = [];
  let errorMessage = "";

  try {
    trailers = await getTrailers();
  } catch (error) {
    errorMessage =
      error instanceof Error ? error.message : "Failed to load trailers";
  }

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Trailer registry</p>
          <h1 className="hero-title">Trailers</h1>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Trailers</span>
            <strong className="metric-value mono">{trailers.length}</strong>
          </div>
        </div>
      </section>

      {errorMessage ? (
        <section className="panel">
          <h2>Unable to load trailers</h2>
          <p>{errorMessage}</p>
        </section>
      ) : (
        <>
          <TrailerManager trailers={trailers} />
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
