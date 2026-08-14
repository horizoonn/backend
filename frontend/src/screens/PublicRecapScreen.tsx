import { useState } from 'react';

import type { SharedRecap } from '../api/types';
import { BadgeCard } from '../components/BadgeCard';
import styles from './PublicRecapScreen.module.css';

function artUrl(name: string): string {
  return `/art/${name}.webp`;
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat('ru-RU').format(value);
}

function publicBadgeIcon(code: string): string {
  return artUrl(code.startsWith('seller') ? 'badge-sales' : 'badge');
}

export function PublicRecapScreen({ recap }: { recap: SharedRecap }) {
  const [shareState, setShareState] = useState<
    | { status: 'idle' }
    | { status: 'loading' }
    | { status: 'success'; message: string }
  >({ status: 'idle' });
  const url = window.location.href;

  const share = async () => {
    if (shareState.status === 'loading') {
      return;
    }

    setShareState({ status: 'loading' });

    if (navigator.share) {
      try {
        await navigator.share({
          title: `Итоги ${recap.year} — Авито`,
          text: `${recap.displayName} — ${recap.archetype.title}`,
          url,
        });
        setShareState({ status: 'success', message: 'Ссылка отправлена' });
        return;
      } catch (cause: unknown) {
        if (cause instanceof DOMException && cause.name === 'AbortError') {
          setShareState({ status: 'idle' });
          return;
        }
        console.warn(cause);
      }
    }

    try {
      await navigator.clipboard.writeText(url);
      setShareState({ status: 'success', message: 'Ссылка скопирована' });
    } catch (cause: unknown) {
      console.warn(cause);
      window.prompt('Скопируйте ссылку на публичные итоги', url);
      setShareState({ status: 'success', message: 'Ссылка готова' });
    }
  };

  return (
    <main className={styles.page}>
      <header className={styles.top}>
        <a href="/" aria-label="Авито — создать свои итоги">
          <img src="/brand/avito-logo.svg" alt="" />
        </a>
        <span>Публичные итоги · {recap.year}</span>
      </header>

      <section className={styles.poster} aria-labelledby="public-recap-title">
        <div className={styles.content}>
          <header className={styles.identity}>
            <p className={styles.label}>
              ИТОГИ {recap.displayName.toLocaleUpperCase('ru-RU')} · {recap.year}
            </p>
            <h1 id="public-recap-title">
              <span className={styles.name}>{recap.displayName} —</span>
              <span className={styles.archetype}>{recap.archetype.title}</span>
            </h1>
            <p className={styles.description}>{recap.archetype.description}</p>
            <p className={styles.privacy}>Публичная версия · без приватных деталей</p>
          </header>

          <div className={styles.facts}>
            <article>
              <img src={artUrl('active_days')} alt="" aria-hidden="true" />
              <div>
                <strong>{formatNumber(recap.activeDays)}</strong>
                <span>активных дней</span>
              </div>
            </article>
            {typeof recap.views === 'number' ? (
              <article>
                <img src={artUrl('views')} alt="" aria-hidden="true" />
                <div>
                  <strong>{formatNumber(recap.views)}</strong>
                  <span>просмотров</span>
                </div>
              </article>
            ) : null}
            {recap.topCategory ? (
              <article className={styles.categoryFact}>
                <img src={artUrl('interests')} alt="" aria-hidden="true" />
                <div>
                  <strong>{recap.topCategory.categoryTitle}</strong>
                  <span>главный интерес</span>
                  {recap.topCategory.subcategoryTitle ? (
                    <small>{recap.topCategory.subcategoryTitle}</small>
                  ) : null}
                </div>
              </article>
            ) : null}
          </div>

          <div className={styles.footer}>
            {recap.interestSummary ? (
              <article className={styles.interest}>
                <span aria-hidden="true">“</span>
                <div>
                  <strong>ИНТЕРЕСЫ ГОДА</strong>
                  <p>{recap.interestSummary}</p>
                </div>
              </article>
            ) : null}

            <div className={styles.actions}>
              <a href="/">Создать свои итоги</a>
              <button
                type="button"
                onClick={() => void share()}
                disabled={shareState.status === 'loading'}
              >
                {shareState.status === 'loading'
                  ? 'Подготавливаем ссылку…'
                  : 'Поделиться карточкой ↗'}
              </button>
              <p role="status">{shareState.status === 'success' ? shareState.message : '\u00a0'}</p>
            </div>
          </div>
        </div>

        <aside
          className={styles.visualPanel}
          aria-label="Архетип и достижения года"
        >
          <div className={styles.visualHeading}>
            <span>АРХЕТИП ГОДА</span>
            <strong aria-hidden="true">{recap.archetype.title.charAt(0).toLocaleUpperCase('ru-RU')}</strong>
          </div>

          <div className={styles.visual}>
            <img src={artUrl(`archetype-${recap.archetype.code}`)} alt="" aria-hidden="true" />
          </div>

          {recap.badges.length ? (
            <section className={styles.badges} aria-label="Достижения года">
              {recap.badges.map((badge) => (
                <BadgeCard
                  key={badge.code}
                  title={badge.title}
                  description={badge.description}
                  level={badge.level}
                  iconUrl={badge.iconUrl ?? publicBadgeIcon(badge.code)}
                />
              ))}
            </section>
          ) : (
            <p className={styles.noBadges}>Весь год — уже достижение</p>
          )}
        </aside>
      </section>
    </main>
  );
}
