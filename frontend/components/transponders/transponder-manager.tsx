"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { createTransponder, updateTransponder } from "../../lib/api";
import type {
  Transponder,
  TransponderUpsertRequest,
  Truck,
} from "../../lib/types";

type TransponderFormState = {
  truckId: string;
  transponderNumber: string;
  oldTransponderNumber: string;
  mcCompany: string;
  status: string;
  notes: string;
};

type TransponderManagerProps = {
  transponders: Transponder[];
  trucks: Truck[];
};

function emptyState(): TransponderFormState {
  return {
    truckId: "",
    transponderNumber: "",
    oldTransponderNumber: "",
    mcCompany: "",
    status: "Active",
    notes: "",
  };
}

function fromTransponder(item: Transponder): TransponderFormState {
  return {
    truckId: item.truck_id ?? "",
    transponderNumber: item.transponder_number ?? "",
    oldTransponderNumber: item.old_transponder_number ?? "",
    mcCompany: item.mc_company ?? "",
    status: item.status ?? "Active",
    notes: item.notes ?? "",
  };
}

export default function TransponderManager({
  transponders,
  trucks,
}: TransponderManagerProps) {
  const router = useRouter();
  const [selectedId, setSelectedId] = useState("");
  const [state, setState] = useState<TransponderFormState>(() => emptyState());
  const [errorMessage, setErrorMessage] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const selectedItem = useMemo(
    () => transponders.find((item) => item.id === selectedId) ?? null,
    [selectedId, transponders],
  );

  function beginEdit(item: Transponder) {
    setSelectedId(item.id);
    setState(fromTransponder(item));
    setErrorMessage("");
  }

  function beginNew() {
    setSelectedId("");
    setState(emptyState());
    setErrorMessage("");
  }

  function updateField<K extends keyof TransponderFormState>(
    key: K,
    value: TransponderFormState[K],
  ) {
    setState((current) => ({
      ...current,
      [key]: value,
    }));
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");

    if (state.transponderNumber.trim() === "") {
      setErrorMessage("Transponder number is required.");
      return;
    }

    const payload: TransponderUpsertRequest = {
      truck_id: state.truckId.trim() || null,
      transponder_number: state.transponderNumber.trim(),
      old_transponder_number: state.oldTransponderNumber.trim() || null,
      mc_company: state.mcCompany.trim() || null,
      status: state.status.trim() || null,
      notes: state.notes.trim() || null,
    };

    setIsSaving(true);
    try {
      if (selectedId) {
        await updateTransponder(selectedId, payload);
      } else {
        await createTransponder(payload);
      }

      beginNew();
      router.refresh();
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "Failed to save transponder",
      );
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="entry-layout">
      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>{selectedId ? "Edit transponder" : "Add transponder"}</h2>
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
              <span>Transponder number</span>
              <input
                value={state.transponderNumber}
                onChange={(event) =>
                  updateField("transponderNumber", event.target.value)
                }
                placeholder="Transponder number"
              />
            </label>

            <label className="form-field">
              <span>Old transponder number</span>
              <input
                value={state.oldTransponderNumber}
                onChange={(event) =>
                  updateField("oldTransponderNumber", event.target.value)
                }
                placeholder="Old number"
              />
            </label>

            <label className="form-field">
              <span>MC company</span>
              <input
                value={state.mcCompany}
                onChange={(event) =>
                  updateField("mcCompany", event.target.value)
                }
                placeholder="MC company"
              />
            </label>

            <label className="form-field">
              <span>Status</span>
              <select
                value={state.status}
                onChange={(event) => updateField("status", event.target.value)}
              >
                <option value="Active">Active</option>
                <option value="Inactive">Inactive</option>
              </select>
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
              Editing transponder{" "}
              {selectedItem.transponder_number ?? selectedItem.id}.
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
                  ? "Update transponder"
                  : "Save transponder"}
            </button>
          </div>
        </form>
      </section>

      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>Transponder assignments</h2>
          <span className="panel-kicker">{transponders.length} records</span>
        </div>

        <div className="table-wrap">
          <table className="dense-table">
            <thead>
              <tr>
                <th>Truck</th>
                <th>Transponder</th>
                <th>Old number</th>
                <th>MC company</th>
                <th>Status</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {transponders.map((item) => (
                <tr key={item.id}>
                  <td className="mono">{item.truck_unit_number ?? "-"}</td>
                  <td className="mono">{item.transponder_number ?? "-"}</td>
                  <td className="mono">{item.old_transponder_number ?? "-"}</td>
                  <td>{item.mc_company ?? "-"}</td>
                  <td>{item.status}</td>
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
