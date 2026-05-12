import type {
    InvoiceParseResult,
    MaintenanceCreateRequest,
    MaintenanceLog,
    Truck,
} from "./types";

const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

async function fetchJson<T>(path: string): Promise<T> {
    const response = await fetch(`${apiBaseUrl}${path}`, { cache: "no-store" });

    if (!response.ok) {
        throw new Error(await readResponseError(response));
    }

    return response.json() as Promise<T>;
}

async function readResponseError(response: Response): Promise<string> {
    const body = await response.text();

    if (!body) {
        return `Request failed: ${response.status} ${response.statusText}`;
    }

    try {
        const parsed = JSON.parse(body) as { error?: unknown };
        if (typeof parsed.error === "string" && parsed.error.trim() !== "") {
            return parsed.error;
        }
    } catch {
        // Fall back to the raw body below.
    }

    return body;
}

export async function getTrucks(): Promise<Truck[]> {
    return fetchJson<Truck[]>("/api/trucks");
}

export async function getTruck(id: string): Promise<Truck> {
    return fetchJson<Truck>(`/api/trucks/${encodeURIComponent(id)}`);
}

export async function getMaintenanceLogs(truckId?: string): Promise<MaintenanceLog[]> {
    const query = truckId ? `?truck_id=${encodeURIComponent(truckId)}` : "";
    return fetchJson<MaintenanceLog[]>(`/api/maintenance${query}`);
}

export async function parseInvoice(file: File): Promise<InvoiceParseResult> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch(`${apiBaseUrl}/api/invoice/parse`, {
        method: "POST",
        body: formData,
    });

    if (!response.ok) {
        throw new Error(await readResponseError(response));
    }

    return response.json() as Promise<InvoiceParseResult>;
}

export async function createMaintenanceLog(
    payload: MaintenanceCreateRequest,
): Promise<MaintenanceLog> {
    const response = await fetch(`${apiBaseUrl}/api/maintenance`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
    });

    if (!response.ok) {
        throw new Error(await readResponseError(response));
    }

    return response.json() as Promise<MaintenanceLog>;
}
