const stats = [
  { label: "Active trucks", value: "38", tone: "green" },
  { label: "In shop", value: "4", tone: "amber" },
  { label: "Overdue PM", value: "7", tone: "red" },
  { label: "DOT expiring < 30 days", value: "3", tone: "muted" },
] as const;

const alerts = [
  {
    unit: "071",
    title: "DOT expires in 12 days",
    detail: "Inspection paperwork pending",
    tone: "amber",
  },
  {
    unit: "102",
    title: "PM overdue by 1,240 miles",
    detail: "Schedule service today",
    tone: "red",
  },
  {
    unit: "088",
    title: "Oil change due next week",
    detail: "Projected at 412 miles remaining",
    tone: "muted",
  },
] as const;

const trucks = [
  {
    unit: "071",
    make: "Freightliner Cascadia",
    status: "ENROUTE",
    pmDue: "1,420",
    expense: "$1,240.00",
  },
  {
    unit: "102",
    make: "Peterbilt 579",
    status: "SHOP",
    pmDue: "-240",
    expense: "$3,880.00",
  },
  {
    unit: "088",
    make: "Kenworth T680",
    status: "ENROUTE",
    pmDue: "412",
    expense: "$860.00",
  },
  {
    unit: "063",
    make: "Volvo VNL",
    status: "STOP",
    pmDue: "2,980",
    expense: "$0.00",
  },
] as const;

function statusTone(status: string) {
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

export default function DashboardPage() {
  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Fleet control</p>
          <h1 className="hero-title">Maintenance dashboard</h1>
          <p className="hero-copy">
            A dense, at-a-glance view of fleet status, maintenance pressure, and
            the trucks that need attention now.
          </p>
        </div>
        <div className="hero-metrics">
          <div className="hero-metric">
            <span className="metric-label">Month spend</span>
            <strong className="metric-value mono">$28,410.00</strong>
          </div>
          <div className="hero-metric">
            <span className="metric-label">Last month</span>
            <strong className="metric-value mono">$24,965.00</strong>
          </div>
        </div>
      </section>

      <section className="stats-grid">
        {stats.map((stat) => (
          <article className="stat-card" key={stat.label}>
            <span className="stat-label">{stat.label}</span>
            <strong className="stat-value mono">{stat.value}</strong>
            <span className={`status-pill ${stat.tone}`}>{stat.label}</span>
          </article>
        ))}
      </section>

      <section className="content-grid">
        <article className="panel panel-span-2">
          <div className="panel-header">
            <h2>Trucks to watch</h2>
            <span className="panel-kicker">sorted by PM urgency</span>
          </div>

          <div className="table-wrap">
            <table className="dense-table">
              <thead>
                <tr>
                  <th>Unit</th>
                  <th>Make</th>
                  <th>Status</th>
                  <th>PM due</th>
                  <th>Last expense</th>
                </tr>
              </thead>
              <tbody>
                {trucks.map((truck) => (
                  <tr key={truck.unit}>
                    <td className="mono">{truck.unit}</td>
                    <td>{truck.make}</td>
                    <td>
                      <span
                        className={`status-pill ${statusTone(truck.status)}`}
                      >
                        {truck.status}
                      </span>
                    </td>
                    <td className="mono">{truck.pmDue}</td>
                    <td className="mono">{truck.expense}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </article>

        <article className="panel">
          <div className="panel-header">
            <h2>Alerts</h2>
            <span className="panel-kicker">action needed</span>
          </div>

          <div className="alert-list">
            {alerts.map((alert) => (
              <div className="alert-item" key={alert.unit}>
                <div>
                  <p className="alert-title">Unit {alert.unit}</p>
                  <p className="alert-detail">{alert.title}</p>
                </div>
                <span className={`status-pill ${alert.tone}`}>
                  {alert.detail}
                </span>
              </div>
            ))}
          </div>
        </article>
      </section>
    </main>
  );
}
