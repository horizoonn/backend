import type {
  ApiErrorBody,
  CreateRecapResponse,
  ErrorCode,
  Profile,
  ProfileList,
  Recap,
  SharedRecap,
  SharedRecapLink,
} from './types';

const BASE_URL: string = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

/**
 * ApiError несёт `code` контракта — по нему выбирается текст для пользователя.
 * Серверный `message` остаётся в консоли: он написан для логов, не для UI.
 */
export class ApiError extends Error {
  readonly code: ErrorCode;
  readonly requestId?: string;
  readonly status: number;

  constructor(status: number, body: Partial<ApiErrorBody>) {
    super(body.message ?? `HTTP ${status}`);
    this.name = 'ApiError';
    this.status = status;
    this.code = body.code ?? 'internal_error';
    this.requestId = body.requestId;
  }
}

const USER_MESSAGES: Record<ErrorCode, string> = {
  bad_request: 'Что-то не так с запросом. Обновите страницу и попробуйте снова.',
  profile_not_found: 'Такого профиля больше нет. Выберите другой.',
  recap_not_found: 'Итоги не найдены. Похоже, ссылка устарела.',
  shared_recap_not_found: 'Публичные итоги не найдены. Проверьте ссылку.',
  not_enough_activity: 'За этот год слишком мало активности, чтобы собрать итоги.',
  recap_not_ready: 'Итоги ещё готовятся. Попробуйте через минуту.',
  rate_limited: 'Слишком много запросов. Подождите немного.',
  internal_error: 'Что-то сломалось на нашей стороне. Мы уже разбираемся.',
};

export function userMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return USER_MESSAGES[error.code];
  }

  return 'Не удалось связаться с сервером. Проверьте соединение.';
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as Partial<ApiErrorBody>;

    throw new ApiError(response.status, body);
  }

  return (await response.json()) as T;
}

export async function listProfiles(): Promise<Profile[]> {
  const { items } = await request<ProfileList>('/profiles');

  return items;
}

/**
 * Генерация идемпотентна по паре (профиль, год): 201 на первый вызов, 200 если
 * итоги уже собраны. Фронту разница не важна — id один и тот же.
 */
export async function createRecap(profileId: string, year: number): Promise<string> {
  const { id } = await request<CreateRecapResponse>('/recaps', {
    method: 'POST',
    body: JSON.stringify({ profileId, year }),
  });

  return id;
}

export async function getRecap(recapId: string): Promise<Recap> {
  return request<Recap>(`/recaps/${recapId}`);
}

export async function shareRecap(recapId: string): Promise<SharedRecapLink> {
  return request<SharedRecapLink>(`/recaps/${recapId}/share`, { method: 'POST' });
}

export async function getSharedRecap(token: string): Promise<SharedRecap> {
  return request<SharedRecap>(`/shared-recaps/${encodeURIComponent(token)}`);
}
