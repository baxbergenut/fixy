"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { createTablet, updateTablet } from "../../lib/api";
import type { Tablet, TabletUpsertRequest, Truck } from "../../lib/types";

type TabletFormState = {
  truckId: string;
  imei: string;
  phoneNumber: string;
  deviceMake: string;
  deviceModel: string;
  contractType: string;
  contractStart: string;
  contractEnd: string;
  status: string;
  notes: string;
};

type TabletManagerProps = {
  tablets: Tablet[];
  trucks: Truck[];
};

function emptyState(): TabletFormState {
  return {
    truckId: "",
    imei: "",
    phoneNumber: "",
    deviceMake: "",
    deviceModel: "",
    contractType: "",
    contractStart: "",
    contractEnd: "",
    status: "",
    notes: "",
  };
}

function fromTablet(item: Tablet): TabletFormState {
  return {
    truckId: item.truck_id ?? "",
    imei: item.imei ?? "",
    phoneNumber: item.phone_number ?? "",
    deviceMake: item.device_make ?? "",
    deviceModel: item.device_model ?? "",
    contractType: item.contract_type ?? "",
    contractStart: item.contract_start ?? "",
    contractEnd: item.contract_end ?? "",
    status: item.status ?? "",
    notes: item.notes ?? "",
  };
}

export default function TabletManager({ tablets, trucks }: TabletManagerProps) {
  const router = useRouter();
  const [selectedId, setSelectedId] = useState("");
  const [state, setState] = useState<TabletFormState>(() => emptyState());
  const [errorMessage, setErrorMessage] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const selectedItem = useMemo(
    () => tablets.find((item) => item.id === selectedId) ?? null,
    [selectedId, tablets],
  );

  function beginEdit(item: Tablet) {
    setSelectedId(item.id);
    setState(fromTablet(item));
    setErrorMessage("");
  }

  function beginNew() {
    setSelectedId("");
    setState(emptyState());
    setErrorMessage("");
  }

  function updateField<K extends keyof TabletFormState>(
    key: K,
    value: TabletFormState[K],
  ) {
    setState((current) => ({
      ...current,
      [key]: value,
    }));
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");

    const payload: TabletUpsertRequest = {
      truck_id: state.truckId.trim() || null,
      imei: state.imei.trim() || null,
      phone_number: state.phoneNumber.trim() || null,
      device_make: state.deviceMake.trim() || null,
      device_model: state.deviceModel.trim() || null,
      contract_type: state.contractType.trim() || null,
      contract_start: state.contractStart.trim() || null,
      contract_end: state.contractEnd.trim() || null,
      status: state.status.trim() || null,
      notes: state.notes.trim() || null,
    };

    setIsSaving(true);
    try {
      if (selectedId) {
        await updateTablet(selectedId, payload);
      } else {
        await createTablet(payload);
      }

      beginNew();
      router.refresh();
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "Failed to save tablet",
      );
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="entry-layout">
      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>{selectedId ? "Edit tablet" : "Add tablet"}</h2>
          <button className="panel-link" onClick={beginNew} type="button">
            Clear form
          </button>
        </div>

        <form className="entry-form" onSubmit={handleSubmit}>
          <div className="entry-grid">
            <label className="form-field form-field-wide">
              <span>Assigned truck</span>
              <select
                value={state.truckId}
                onChange={(event) => updateField("truckId", event.target.value)}
              >
                <option value="">Unassigned</option>
                {trucks.map((truck) => (
                  <option key={truck.id} value={truck.id}>
                    Unit {truck.unit_number}
                    {truck.company ? ` - ${truck.company}` : ""}
                  </option>
                ))}
              </select>
            </label>

            <label className="form-field">
              <span>IMEI</span>
              <input
                value={state.imei}
                onChange={(event) => updateField("imei", event.target.value)}
                placeholder="IMEI"
              />
            </label>

            <label className="form-field">
              <span>Phone number</span>
              <input
                value={state.phoneNumber}
                onChange={(event) =>
                  updateField("phoneNumber", event.target.value)
                }
                placeholder="Phone number"
              />
            </label>

            <label className="form-field">
              <span>Device make</span>
              <input
                value={state.deviceMake}
                onChange={(event) =>
                  updateField("deviceMake", event.target.value)
                }
                placeholder="Device make"
              />
            </label>

            <label className="form-field">
              <span>Device model</span>
              <input
                value={state.deviceModel}
                onChange={(event) =>
                  updateField("deviceModel", event.target.value)
                }
                placeholder="Device model"
              />
            </label>

            <label className="form-field">
              <span>Contract type</span>
              <input
                value={state.contractType}
                onChange={(event) =>
                  updateField("contractType", event.target.value)
                }
                placeholder="Contract type"
              />
            </label>

            <label className="form-field">
              <span>Contract start</span>
              <input
                type="date"
                value={state.contractStart}
                onChange={(event) =>
                  updateField("contractStart", event.target.value)
                }
              />
            </label>

            <label className="form-field">
              <span>Contract end</span>
              <input
                type="date"
                value={state.contractEnd}
                onChange={(event) =>
                  updateField("contractEnd", event.target.value)
                }
              />
            </label>

            <label className="form-field">
              <span>Status</span>
              <input
                value={state.status}
                onChange={(event) => updateField("status", event.target.value)}
                placeholder="Status"
              />
            </label>

            <label className="form-field form-field-wide">
              <span>Notes</span>
              <textarea
                rows={4}
                value={state.notes}
                onChange={(event) => updateField("notes", event.target.value)}
                placeholder="Notes"
              />
            </label>
          </div>

          {selectedItem ? (
            <p className="helper-text">
              Editing tablet {selectedItem.imei ?? selectedItem.id}.
            </p>
          ) : null}

          {errorMessage ? <p className="form-error">{errorMessage}</p> : null}

          <div className="form-actions">
            <button
              className="primary-button"
              disabled={isSaving}
              type="submit"
            >
              {isSaving
                ? "Saving..."
                : selectedId
                  ? "Update tablet"
                  : "Save tablet"}
            </button>
          </div>
        </form>
      </section>

      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>Tablet inventory</h2>
          <span className="panel-kicker">{tablets.length} records</span>
        </div>

        <div className="table-wrap">
          <table className="dense-table">
            <thead>
              <tr>
                <th>Truck</th>
                <th>IMEI</th>
                <th>Phone</th>
                <th>Device</th>
                <th>Contract</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {tablets.map((item) => (
                <tr key={item.id}>
                  <td className="mono">{item.truck_unit_number ?? "-"}</td>
                  <td className="mono">{item.imei ?? "-"}</td>
                  <td className="mono">{item.phone_number ?? "-"}</td>
                  <td>
                    {[item.device_make, item.device_model]
                      .filter(Boolean)
                      .join(" ") || "-"}
                  </td>
                  <td>
                    {[item.contract_start, item.contract_end]
                      .filter(Boolean)
                      .join(" to ") || "-"}
                  </td>
                  <td>
                    <button
                      className="panel-link"
                      onClick={() => beginEdit(item)}
                      type="button"
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
