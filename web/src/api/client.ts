export type ApiClient = {
  get<T>(path: string): Promise<T>;
  post<TResponse, TBody extends object>(path: string, body: TBody): Promise<TResponse>;
};

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

const defaultBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export function createApiClient(baseUrl = defaultBaseUrl): ApiClient {
  const normalizedBaseUrl = baseUrl.replace(/\/$/, '');

  async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const response = await fetch(`${normalizedBaseUrl}${path}`, {
      ...init,
      headers: {
        Accept: 'application/json',
        ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
        ...init?.headers,
      },
    });

    if (!response.ok) {
      throw new ApiError(response.status, await errorMessage(response));
    }

    return (await response.json()) as T;
  }

  return {
    get<T>(path: string) {
      return request<T>(path);
    },
    post<TResponse, TBody extends object>(path: string, body: TBody) {
      return request<TResponse>(path, {
        method: 'POST',
        body: JSON.stringify(body),
      });
    },
  };
}

async function errorMessage(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: string };
    return body.error || response.statusText || 'Request failed';
  } catch {
    return response.statusText || 'Request failed';
  }
}
