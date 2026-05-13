import type {
    InvoiceParseResult,
    MaintenanceCreateRequest,
    MaintenanceLog,
    Tablet,
    TabletUpsertRequest,
    Trailer,
    TrailerUpsertRequest,
    Transponder,
    TransponderUpsertRequest,
    Truck,
} from "./types";

function resolveApiBaseUrl(): string {
    const configuredBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL?.trim();

    if (configuredBaseUrl) {
        return configuredBaseUrl.replace(/\/+$/, "");
    }

    if (typeof window !== "undefined") {
        return `${window.location.protocol}//${window.location.hostname}:8080`;
    }

    return "http://localhost:8080";
}

const apiBaseUrl = resolveApiBaseUrl();

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

export async function getTrailers(): Promise<Trailer[]> {
    return fetchJson<Trailer[]>("/api/trailers");
}

export async function getTransponders(): Promise<Transponder[]> {
    return fetchJson<Transponder[]>("/api/transponders");
}

export async function getTablets(): Promise<Tablet[]> {
    return fetchJson<Tablet[]>("/api/tablets");
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

async function writeJson<T>(path: string, method: "POST" | "PATCH", payload: object): Promise<T> {
    const response = await fetch(`${apiBaseUrl}${path}`, {
        method,
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
    });

    if (!response.ok) {
        throw new Error(await readResponseError(response));
    }

    return response.json() as Promise<T>;
}

export async function createTrailer(payload: TrailerUpsertRequest): Promise<Trailer> {
    return writeJson<Trailer>("/api/trailers", "POST", payload);
}

export async function updateTrailer(
    id: string,
    payload: TrailerUpsertRequest,
): Promise<Trailer> {
    return writeJson<Trailer>(`/api/trailers/${encodeURIComponent(id)}`, "PATCH", payload);
}

export async function createTransponder(
    payload: TransponderUpsertRequest,
): Promise<Transponder> {
    return writeJson<Transponder>("/api/transponders", "POST", payload);
}

export async function updateTransponder(
    id: string,
    payload: TransponderUpsertRequest,
): Promise<Transponder> {
    return writeJson<Transponder>(`/api/transponders/${encodeURIComponent(id)}`, "PATCH", payload);
}

export async function createTablet(payload: TabletUpsertRequest): Promise<Tablet> {
    return writeJson<Tablet>("/api/tablets", "POST", payload);
}

export async function updateTablet(
    id: string,
    payload: TabletUpsertRequest,
): Promise<Tablet> {
    return writeJson<Tablet>(`/api/tablets/${encodeURIComponent(id)}`, "PATCH", payload);
}
