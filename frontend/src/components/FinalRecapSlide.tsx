import type { Archetype, ArchetypeCode, Cta, Slide } from '../api/types';
import { recapActionRedirectUrl } from '../api/client';
import styles from './FinalRecapSlide.module.css';

type FinalSlide = Extract<Slide, { type: 'final' }>;
type FavoriteCategorySlide = Extract<Slide, { type: 'favorite_category' }>;

const HEADLINES: Record<ArchetypeCode, string> = {
  collector: 'Год, в котором важные находки оставались рядом',
  dealmaker: 'Год, в котором вещи находили новых владельцев',
  explorer: 'Год, в котором вы не переставали искать новое',
  negotiator: 'Год, в котором вы всегда находили общий язык',
};

const STAT_ART: Record<string, string> = {
  active_days: 'active_days',
  views: 'views',
  favorites: 'favorites',
  messages: 'messages',
  seasons: 'season-spring',
};

function artUrl(name: string): string {
  return `/art/${name}.webp`;
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat('ru-RU').format(value);
}

function selectProductActions(actions: Cta[] | undefined, archetype: ArchetypeCode): Cta[] {
  if (!actions?.length) {
    return [];
  }

  const productActions = actions.filter(
    (action) =>
      action.action === 'open_category' ||
      action.action === 'open_favorites' ||
      action.action === 'create_listing',
  );
  const preferredAction: Partial<Record<ArchetypeCode, Cta['action']>> = {
    collector: 'open_favorites',
    dealmaker: 'create_listing',
    explorer: 'open_category',
    negotiator: 'open_category',
  };

  const primary = productActions.find((action) => action.action === preferredAction[archetype]);
  const ordered = primary
    ? [primary, ...productActions.filter((action) => action !== primary)]
    : productActions;

  return ordered.slice(0, 2);
}

interface FinalRecapSlideProps {
  slide: FinalSlide;
  recapId: string;
  year: number;
  archetype: Archetype;
  favoriteCategory?: FavoriteCategorySlide;
  interestSummary?: string;
  onShare: () => void;
  shareDisabled: boolean;
  shareLabel?: string;
  shareFeedback?: { message: string; failed: boolean };
  shareUrl?: string;
  onExit: () => void;
}

export function FinalRecapSlide({
  slide,
  recapId,
  year,
  archetype,
  favoriteCategory,
  interestSummary,
  onShare,
  shareDisabled,
  shareLabel,
  shareFeedback,
  shareUrl,
  onExit,
}: FinalRecapSlideProps) {
  const productActions = selectProductActions(slide.actions, archetype.code);
  const shareAction = slide.actions?.find((action) => action.action === 'share_recap');

  return (
    <section className={styles.slide} aria-labelledby="final-title">
      <div className={styles.cover}>
        <div className={styles.content}>
          <header className={styles.header}>
            <p className={styles.label}>ВАШИ ИТОГИ {year}</p>
            <h1 id="final-title">{HEADLINES[archetype.code]}</h1>
            <p className={styles.subtitle}>{archetype.description}</p>
          </header>

          {slide.stats?.length ? (
            <ul className={styles.stats} data-count={Math.min(slide.stats.length, 4)}>
              {slide.stats.slice(0, 4).map((stat) => (
                <li key={stat.code}>
                  <div className={styles.statArt} aria-hidden="true">
                    <img src={artUrl(STAT_ART[stat.code] ?? 'final')} alt="" />
                  </div>
                  <strong>{formatNumber(stat.value)}</strong>
                  <span>{stat.label}</span>
                </li>
              ))}
            </ul>
          ) : null}

          <div className={styles.actions}>
            <div className={styles.actionButtons}>
              {shareAction ? (
                <button type="button" onClick={onShare} disabled={shareDisabled}>
                  {shareLabel ?? 'Поделиться итогами ↗'}
                </button>
              ) : null}

              {productActions.map((action, actionIndex) =>
                action.action === 'open_category' ||
                action.action === 'open_favorites' ||
                action.action === 'create_listing' ? (
                  <a
                    key={action.action}
                    className={
                      actionIndex === 0
                        ? styles.productAction
                        : styles.secondaryProductAction
                    }
                    href={recapActionRedirectUrl(recapId, action.action)}
                  >
                    {action.title}
                  </a>
                ) : null,
              )}
            </div>

            <button type="button" className={styles.exit} onClick={onExit}>
              Другой профиль
            </button>

            <p
              className={shareFeedback?.failed ? styles.feedbackError : styles.feedback}
              role="status"
              aria-live="polite"
              aria-atomic="true"
            >
              {shareFeedback?.message ?? '\u00a0'}
            </p>

            {shareUrl ? (
              <a className={styles.publicLink} href={shareUrl}>
                Открыть публичную карточку
              </a>
            ) : null}
          </div>
        </div>

        <aside className={styles.visualPanel} aria-label="Архетип и главный интерес года">
          <div className={styles.visualHeading}>
            <span>АРХЕТИП ГОДА</span>
            <strong aria-hidden="true">
              {archetype.title.charAt(0).toLocaleUpperCase('ru-RU')}
            </strong>
          </div>

          <figure className={styles.archetypeVisual}>
            <img src={artUrl(`archetype-${archetype.code}`)} alt="" aria-hidden="true" />
            <figcaption>
              <span>Ваш архетип</span>
              <strong>{archetype.title}</strong>
            </figcaption>
          </figure>

          {favoriteCategory ? (
            <div className={styles.interestCard}>
              <strong>Главный интерес — {favoriteCategory.category.title}</strong>
              <span>{interestSummary ?? favoriteCategory.subcategory?.title}</span>
            </div>
          ) : null}

          <div className={styles.ring} aria-hidden="true" />
        </aside>
      </div>
    </section>
  );
}
