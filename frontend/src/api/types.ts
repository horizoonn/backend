/**
 * Типы контракта из backend/recap/api/recap/v1/openapi.yaml.
 *
 * Написаны руками, потому что фронт использует малую часть схемы. Если контракт
 * разойдётся с этим файлом, ломаться будет разбор ответа, а не сборка, поэтому
 * при правке спеки сюда надо заглядывать.
 */

export type Uuid = string;

export interface Profile {
  id: Uuid;
  name: string;
  surname: string;
  avatarUrl?: string | null;
  registeredAt?: string;
  hint?: string;
}

export interface ProfileList {
  items: Profile[];
}

export interface CreateRecapRequest {
  profileId: Uuid;
  year: number;
}

export interface CreateRecapResponse {
  id: Uuid;
}

export type ArchetypeCode = 'collector' | 'dealmaker' | 'negotiator' | 'explorer';

export type MetricCode =
  | 'active_days'
  | 'views'
  | 'favorites'
  | 'purchases'
  | 'sales'
  | 'messages'
  | 'categories'
  | 'listings';

export interface ArchetypeReason {
  metric: MetricCode;
  value: string;
  explanation: string;
}

export interface Archetype {
  code: ArchetypeCode;
  title: string;
  description: string;
  reasons: ArchetypeReason[];
}

export interface CategoryRef {
  id: Uuid;
  title: string;
}

export interface ListingRef {
  id: Uuid;
  title: string;
  price?: number | null;
  imageUrl?: string | null;
  categoryId?: Uuid;
  addedAt?: string;
}

export interface AmountRange {
  min: number;
  max: number | null;
  currency: 'RUB';
  label: string;
}

export interface Badge {
  code: string;
  title: string;
  description?: string;
  level?: 'bronze' | 'silver' | 'gold';
}

export type CtaAction =
  'open_listing' | 'open_category' | 'open_favorites' | 'create_listing' | 'share_recap';

export interface Cta {
  action: CtaAction;
  title: string;
  listingId?: Uuid | null;
  categoryId?: Uuid | null;
  url?: string | null;
}

export type Season = 'winter' | 'spring' | 'summer' | 'autumn';

export interface PeriodInterest {
  period: Season;
  category: CategoryRef;
  subcategory?: CategoryRef | null;
  weight?: number;
}

export interface StatTile {
  code: 'active_days' | 'views' | 'favorites' | 'messages' | 'seasons';
  value: number;
  label: string;
}

interface SlideBase {
  title: string;
  subtitle?: string;
  cta?: Cta | null;
}

export type Slide =
  | (SlideBase & { type: 'intro'; year: number })
  | (SlideBase & { type: 'active_days'; activeDays: number })
  | (SlideBase & { type: 'views'; views: number })
  | (SlideBase & {
      type: 'favorites';
      favorites: number;
      oldestFavorite?: ListingRef | null;
      stillAvailable?: number;
    })
  | (SlideBase & {
      type: 'favorite_category';
      category: CategoryRef;
      subcategory?: CategoryRef | null;
      share?: number;
      recommendations?: ListingRef[];
    })
  | (SlideBase & { type: 'purchases'; purchases: number; badge?: Badge | null })
  | (SlideBase & {
      type: 'sales';
      sales: number;
      amountRange?: AmountRange | null;
      badge?: Badge | null;
    })
  | (SlideBase & { type: 'messages'; messages: number })
  | (SlideBase & { type: 'interests'; periods: PeriodInterest[]; shiftSummary?: string })
  | (SlideBase & { type: 'archetype'; archetype: Archetype })
  | (SlideBase & { type: 'final'; stats?: StatTile[]; actions?: Cta[] });

export type SlideType = Slide['type'];

export interface Recap {
  id: Uuid;
  profileId: Uuid;
  year: number;
  status: 'generating' | 'ready' | 'failed';
  archetype: Archetype;
  slides: Slide[];
  generatedAt: string;
}

export type ErrorCode =
  | 'bad_request'
  | 'profile_not_found'
  | 'recap_not_found'
  | 'shared_recap_not_found'
  | 'not_enough_activity'
  | 'recap_not_ready'
  | 'rate_limited'
  | 'internal_error';

/** Тело ошибки контракта. `message` — для логов, пользователю его не показываем. */
export interface ApiErrorBody {
  code: ErrorCode;
  message: string;
  requestId?: Uuid;
}

export type BadgeLevel = 'bronze' | 'silver' | 'gold';

export interface SharedRecapLink {
  token: string;
  url: string;
  createdAt: string;
}

export interface SharedRecap {
  year: number;
  displayName: string;
  archetype: { code: ArchetypeCode; title: string; description: string };
  activeDays: number;
  views?: number;
  topCategory?: { categoryTitle: string; subcategoryTitle?: string | null };
  interestSummary?: string;
  badges: Array<{
    code: string;
    title: string;
    description: string;
    level: BadgeLevel;
    iconUrl?: string;
  }>;
}
