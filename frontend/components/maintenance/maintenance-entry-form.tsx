"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { createMaintenanceLog, parseInvoice } from "../../lib/api";
import type {
  InvoiceParseResult,
  MaintenanceCreateRequest,
  Truck,
} from "../../lib/types";

const categoryOptions = [
  "PM Service",
  "Oil change",
  "Tire issue",
  "Engine issue",
  "Towing",
  "Road Service",
  "Body work",
  "Leakage",
  "Kris Shop",
  "Truck Wash/Detailing",
  "Electrical issue",
  "Fluids/Truck Parts",
  "Brakes/Drums/Rotors",
  "Scale",
  "Other",
] as const;

type MaintenanceFormState = {
  truckId: string;
  expenseDate: string;
  driverName: string;
  amount: string;
  category: string;
  description: string;
  referenceNumber: string;
  paymentType: string;
  whoCovers: string;
  paidBy: string;
  managerVerified: boolean;
  accountingVerified: boolean;
};

type MaintenanceEntryFormProps = {
  trucks: Truck[];
};

function todayIsoDate() {
  return new Date().toISOString().slice(0, 10);
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(value);
}

function buildInitialState(): MaintenanceFormState {
  return {
    truckId: "",
    expenseDate: todayIsoDate(),
    driverName: "",
    amount: "",
    category: "Other",
    description: "",
    referenceNumber: "",
    paymentType: "",
    whoCovers: "Company",
    paidBy: "",
    managerVerified: false,
    accountingVerified: false,
  };
}

function resolveTruckId(trucks: Truck[], unitNumber: string | null) {
  if (!unitNumber) {
    return "";
  }

  const normalizedUnit = unitNumber.trim();
  return trucks.find((truck) => truck.unit_number === normalizedUnit)?.id ?? "";
}

function valueOrEmpty(value: string | null | undefined) {
  return value ?? "";
}

export default function MaintenanceEntryForm({
  trucks,
}: MaintenanceEntryFormProps) {
  const router = useRouter();
  const [state, setState] = useState<MaintenanceFormState>(() =>
    buildInitialState(),
  );
  const [fileName, setFileName] = useState("");
  const [parsedInvoice, setParsedInvoice] = useState<InvoiceParseResult | null>(
    null,
  );
  const [isParsing, setIsParsing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  const selectedTruck = useMemo(
    () => trucks.find((truck) => truck.id === state.truckId) ?? null,
    [state.truckId, trucks],
  );

  function applyParsedInvoice(result: InvoiceParseResult) {
    const matchingTruckId = resolveTruckId(trucks, result.truck_unit_number);

    setState((current) => ({
      ...current,
      truckId: matchingTruckId || current.truckId,
      expenseDate: result.expense_date ?? current.expenseDate,
      driverName: valueOrEmpty(result.driver_name),
      amount:
        result.amount === null || Number.isNaN(result.amount)
          ? current.amount
          : String(result.amount),
      category: result.category ?? current.category,
      description: valueOrEmpty(result.description),
      referenceNumber: valueOrEmpty(result.reference_number),
      whoCovers: current.whoCovers || "Company",
    }));

    setParsedInvoice(result);
  }

  async function handleInvoiceUpload(
    event: React.ChangeEvent<HTMLInputElement>,
  ) {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    setFileName(file.name);
    setIsParsing(true);
    setErrorMessage("");

    try {
      const result = await parseInvoice(file);
      applyParsedInvoice(result);
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "Failed to parse invoice",
      );
    } finally {
      setIsParsing(false);
      event.target.value = "";
    }
  }

  function updateField<K extends keyof MaintenanceFormState>(
    key: K,
    value: MaintenanceFormState[K],
  ) {
    setState((current) => ({
      ...current,
      [key]: value,
    }));
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");

    if (!state.truckId) {
      setErrorMessage("Choose a truck before saving the entry.");
      return;
    }

    const amountValue = Number(state.amount);
    if (Number.isNaN(amountValue)) {
      setErrorMessage("Enter a valid amount.");
      return;
    }

    const payload: MaintenanceCreateRequest = {
      truck_id: state.truckId,
      trailer_id: null,
      expense_date: state.expenseDate,
      week_label: null,
      driver_name: state.driverName || null,
      amount: amountValue,
      category: state.category,
      payment_type: state.paymentType || null,
      description: state.description || null,
      reference_number: state.referenceNumber || null,
      who_covers: state.whoCovers || null,
      paid_by: state.paidBy || null,
      manager_verified: state.managerVerified,
      accounting_verified: state.accountingVerified,
      invoice_file_url: null,
    };

    setIsSaving(true);
    try {
      await createMaintenanceLog(payload);
      router.push("/maintenance");
      router.refresh();
    } catch (error) {
      setErrorMessage(
        error instanceof Error
          ? error.message
          : "Failed to save maintenance log",
      );
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <div className="entry-layout">
      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>Invoice upload</h2>
          <span className="panel-kicker">Groq prefill</span>
        </div>

        <label className="upload-dropzone">
          <input
            accept="application/pdf,image/*"
            className="file-input"
            onChange={handleInvoiceUpload}
            type="file"
          />
          <strong>
            {isParsing ? "Parsing invoice..." : "Choose invoice file"}
          </strong>
          <span>
            Upload a PDF or image. We send it to Groq, then prefill the
            maintenance fields with the extracted JSON.
          </span>
          {fileName ? (
            <span className="mono upload-file-name">{fileName}</span>
          ) : null}
        </label>

        {parsedInvoice ? (
          <div className="preview-panel">
            <p className="eyebrow">Extracted invoice data</p>
            <div className="preview-grid">
              <div>
                <span className="preview-label">Vendor</span>
                <strong>{parsedInvoice.vendor ?? "-"}</strong>
              </div>
              <div>
                <span className="preview-label">Truck unit</span>
                <strong>{parsedInvoice.truck_unit_number ?? "-"}</strong>
              </div>
              <div>
                <span className="preview-label">Date</span>
                <strong>{parsedInvoice.expense_date ?? "-"}</strong>
              </div>
              <div>
                <span className="preview-label">Amount</span>
                <strong>
                  {parsedInvoice.amount === null
                    ? "-"
                    : formatCurrency(parsedInvoice.amount)}
                </strong>
              </div>
            </div>
          </div>
        ) : null}
      </section>

      <section className="panel entry-panel">
        <div className="panel-header">
          <h2>Maintenance details</h2>
          <span className="panel-kicker">Confirm and save</span>
        </div>

        <form className="entry-form" onSubmit={handleSubmit}>
          <div className="entry-grid">
            <label className="form-field form-field-wide">
              <span>Truck</span>
              <select
                value={state.truckId}
                onChange={(event) => updateField("truckId", event.target.value)}
              >
                <option value="">Select a truck</option>
                {trucks.map((truck) => (
                  <option key={truck.id} value={truck.id}>
                    Unit {truck.unit_number}
                    {truck.company ? ` - ${truck.company}` : ""}
                  </option>
                ))}
              </select>
            </label>

            <label className="form-field">
              <span>Date</span>
              <input
                type="date"
                value={state.expenseDate}
                onChange={(event) =>
                  updateField("expenseDate", event.target.value)
                }
              />
            </label>

            <label className="form-field">
              <span>Category</span>
              <select
                value={state.category}
                onChange={(event) =>
                  updateField("category", event.target.value)
                }
              >
                {categoryOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            </label>

            <label className="form-field">
              <span>Amount</span>
              <input
                inputMode="decimal"
                placeholder="0.00"
                step="0.01"
                type="number"
                value={state.amount}
                onChange={(event) => updateField("amount", event.target.value)}
              />
            </label>

            <label className="form-field">
              <span>Driver name</span>
              <input
                value={state.driverName}
                onChange={(event) =>
                  updateField("driverName", event.target.value)
                }
                placeholder="Driver name"
              />
            </label>

            <label className="form-field">
              <span>Reference number</span>
              <input
                value={state.referenceNumber}
                onChange={(event) =>
                  updateField("referenceNumber", event.target.value)
                }
                placeholder="Invoice or transaction number"
              />
            </label>

            <label className="form-field">
              <span>Payment type</span>
              <input
                value={state.paymentType}
                onChange={(event) =>
                  updateField("paymentType", event.target.value)
                }
                placeholder="Card, EFS, cash, etc."
              />
            </label>

            <label className="form-field">
              <span>Who covers</span>
              <input
                value={state.whoCovers}
                onChange={(event) =>
                  updateField("whoCovers", event.target.value)
                }
                placeholder="Company, driver, vendor"
              />
            </label>

            <label className="form-field">
              <span>Paid by</span>
              <input
                value={state.paidBy}
                onChange={(event) => updateField("paidBy", event.target.value)}
                placeholder="Optional"
              />
            </label>

            <label className="form-field form-field-wide">
              <span>Description</span>
              <textarea
                rows={4}
                value={state.description}
                onChange={(event) =>
                  updateField("description", event.target.value)
                }
                placeholder="Short description of the work performed"
              />
            </label>

            <label className="checkbox-field">
              <input
                checked={state.managerVerified}
                onChange={(event) =>
                  updateField("managerVerified", event.target.checked)
                }
                type="checkbox"
              />
              <span>Manager verified</span>
            </label>

            <label className="checkbox-field">
              <input
                checked={state.accountingVerified}
                onChange={(event) =>
                  updateField("accountingVerified", event.target.checked)
                }
                type="checkbox"
              />
              <span>Accounting verified</span>
            </label>
          </div>

          {selectedTruck ? (
            <p className="helper-text">
              Saving for unit {selectedTruck.unit_number}.
            </p>
          ) : null}

          {errorMessage ? <p className="form-error">{errorMessage}</p> : null}

          <div className="form-actions">
            <button
              className="primary-button"
              disabled={isSaving || isParsing}
              type="submit"
            >
              {isSaving ? "Saving..." : "Save maintenance entry"}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
