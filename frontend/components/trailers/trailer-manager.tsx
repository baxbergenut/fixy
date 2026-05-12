"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { createTrailer, updateTrailer } from "../../lib/api";
import type { Trailer, TrailerUpsertRequest } from "../../lib/types";

type TrailerFormState = {
  unitNumber: string;
  vin: string;
  plateNumber: string;
  year: string;
  make: string;
  usageType: string;
  location: string;
  availability: string;
  notes: string;
};

type TrailerManagerProps = {
  trailers: Trailer[];
};

function emptyState(): TrailerFormState {
  return {
    unitNumber: "",
    vin: "",
    plateNumber: "",
    year: "",
    make: "",
    usageType: "",
    location: "",
    availability: "Ready",
    notes: "",
  };
}

function fromTrailer(trailer: Trailer): TrailerFormState {
  return {
    unitNumber: trailer.unit_number,
    vin: trailer.vin ?? "",
    plateNumber: trailer.plate_number ?? "",
    year: trailer.year === null ? "" : String(trailer.year),
    make: trailer.make ?? "",
    usageType: trailer.usage_type ?? "",
    location: trailer.location ?? "",
    availability: trailer.availability ?? "",
    notes: trailer.notes ?? "",
  };
}

export default function TrailerManager({ trailers }: TrailerManagerProps) {
  const router = useRouter();
  const [selectedId, setSelectedId] = useState("");
  const [state, setState] = useState<TrailerFormState>(() => emptyState());
  const [errorMessage, setErrorMessage] = useState("");
  const [isSaving, setIsSaving] = useState(false);

  const selectedTrailer = useMemo(
    () => trailers.find((trailer) => trailer.id === selectedId) ?? null,
    [selectedId, trailers],
  );

  function beginEdit(trailer: Trailer) {
    setSelectedId(trailer.id);
    setState(fromTrailer(trailer));
    setErrorMessage("");
  }

  function beginNew() {
    setSelectedId("");
    setState(emptyState());
    setErrorMessage("");
  }

  function updateField<K extends keyof TrailerFormState>(
    key: K,
    value: TrailerFormState[K],
  ) {
    setState((current) => ({
      ...current,
      [key]: value,
    }));
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");

    if (state.unitNumber.trim() === "") {
      setErrorMessage("Trailer unit number is required.");
      return;
    }

    const payload: TrailerUpsertRequest = {
      unit_number: state.unitNumber.trim(),
      vin: state.vin.trim() || null,
      plate_number: state.plateNumber.trim() || null,
      year: state.year.trim() === "" ? null : Number(state.year),
      make: state.make.trim() || null,
      usage_type: state.usageType.trim() || null,
      location: state.location.trim() || null,
      availability: state.availability.trim() || null,
      notes: state.notes.trim() || null,
    };

    if (payload.year !== null && Number.isNaN(payload.year)) {
      setErrorMessage("Enter a valid trailer year.");
      return;
    }

    setIsSaving(true);
    try {
      if (selectedId) {
        await updateTrailer(selectedId, payload);
      } else {
        await createTrailer(payload);
      }

      beginNew();
      router.refresh();
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "Failed to save trailer",
      );
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="entry-layout">
      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>{selectedId ? "Edit trailer" : "Add trailer"}</h2>
          <button className="panel-link" onClick={beginNew} type="button">
            Clear form
          </button>
        </div>

        <form className="entry-form" onSubmit={handleSubmit}>
          <div className="entry-grid">
            <label className="form-field">
              <span>Unit number</span>
              <input
                value={state.unitNumber}
                onChange={(event) =>
                  updateField("unitNumber", event.target.value)
                }
                placeholder="Trailer unit"
              />
            </label>

            <label className="form-field">
              <span>Plate number</span>
              <input
                value={state.plateNumber}
                onChange={(event) =>
                  updateField("plateNumber", event.target.value)
                }
                placeholder="Plate"
              />
            </label>

            <label className="form-field">
              <span>VIN</span>
              <input
                value={state.vin}
                onChange={(event) => updateField("vin", event.target.value)}
                placeholder="VIN"
              />
            </label>

            <label className="form-field">
              <span>Year</span>
              <input
                inputMode="numeric"
                value={state.year}
                onChange={(event) => updateField("year", event.target.value)}
                placeholder="Year"
              />
            </label>

            <label className="form-field">
              <span>Make</span>
              <input
                value={state.make}
                onChange={(event) => updateField("make", event.target.value)}
                placeholder="Make"
              />
            </label>

            <label className="form-field">
              <span>Usage type</span>
              <input
                value={state.usageType}
                onChange={(event) =>
                  updateField("usageType", event.target.value)
                }
                placeholder="Usage type"
              />
            </label>

            <label className="form-field">
              <span>Location</span>
              <input
                value={state.location}
                onChange={(event) =>
                  updateField("location", event.target.value)
                }
                placeholder="Location"
              />
            </label>

            <label className="form-field">
              <span>Availability</span>
              <input
                value={state.availability}
                onChange={(event) =>
                  updateField("availability", event.target.value)
                }
                placeholder="Ready, N/A, Returned, SALE"
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

          {selectedTrailer ? (
            <p className="helper-text">
              Editing trailer {selectedTrailer.unit_number}.
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
                  ? "Update trailer"
                  : "Save trailer"}
            </button>
          </div>
        </form>
      </section>

      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>Trailer registry</h2>
          <span className="panel-kicker">{trailers.length} trailers</span>
        </div>

        <div className="table-wrap">
          <table className="dense-table">
            <thead>
              <tr>
                <th>Unit</th>
                <th>Plate</th>
                <th>Make / year</th>
                <th>Location</th>
                <th>Availability</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {trailers.map((trailer) => (
                <tr key={trailer.id}>
                  <td className="mono">{trailer.unit_number}</td>
                  <td>{trailer.plate_number ?? "-"}</td>
                  <td>
                    {[trailer.make, trailer.year].filter(Boolean).join(" ") ||
                      "-"}
                  </td>
                  <td>{trailer.location ?? "-"}</td>
                  <td>{trailer.availability ?? "-"}</td>
                  <td>
                    <button
                      className="panel-link"
                      onClick={() => beginEdit(trailer)}
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
