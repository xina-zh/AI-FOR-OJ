export class ApiError extends Error {
  readonly status: number;
  readonly details?: unknown;

  constructor(status: number, message: string, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.details = details;
  }
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...(init.headers ?? {}),
    },
  });

  const text = await response.text();
  const data = text ? safeParseJSON(text) : undefined;

  if (!response.ok) {
    const message = errorMessage(data) ?? response.statusText ?? 'request failed';
    throw new ApiError(response.status, message, data);
  }

  return data as T;
}

export function jsonBody(value: unknown): string {
  return JSON.stringify(value);
}

function safeParseJSON(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

function errorMessage(data: unknown): string | undefined {
  if (data && typeof data === 'object' && 'error' in data) {
    const value = (data as { error?: unknown }).error;
    return typeof value === 'string' ? value : undefined;
  }
  return undefined;
}
