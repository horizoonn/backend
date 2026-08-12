import { useState } from 'react';

import type { SharedRecap } from '../api/types';
import { BadgeCard } from '../components/BadgeCard';

function artUrl(name: string): string {
  return `/art/${name}.png`;
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat('ru-RU').format(value);
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
    <main className="public-recap">
      <header className="public-poster__top">
        <a href="/" aria-label="Авито — создать свои итоги">
          Авито
        </a>
        <span>{recap.year}</span>
      </header>

      <section className="public-poster" aria-labelledby="public-recap-title">
        <header className="public-poster__identity">
          <p className="public-poster__label">ПУБЛИЧНЫЕ ИТОГИ {recap.year}</p>
          <h1 id="public-recap-title">
            {recap.displayName} — <span>{recap.archetype.title}</span>
          </h1>
          <p className="public-poster__description">{recap.archetype.description}</p>
          <p className="public-poster__privacy">
            Без приватных деталей — только безопасные моменты года.
          </p>
        </header>

        <div className="public-poster__visual" aria-hidden="true">
          <div>
            <img src={artUrl(`archetype-${recap.archetype.code}`)} alt="" />
          </div>
        </div>

        <div className="public-poster__facts">
          <article className="public-poster__fact--active">
            <img src={artUrl('active_days')} alt="" aria-hidden="true" />
            <strong>{formatNumber(recap.activeDays)}</strong>
            <span>активных дней</span>
          </article>
          {typeof recap.views === 'number' ? (
            <article className="public-poster__fact--views">
              <img src={artUrl('views')} alt="" aria-hidden="true" />
              <strong>{formatNumber(recap.views)}</strong>
              <span>просмотров</span>
            </article>
          ) : null}
          {recap.topCategory ? (
            <article>
              <strong>{recap.topCategory.categoryTitle}</strong>
              <span>главный интерес</span>
              {recap.topCategory.subcategoryTitle ? (
                <small>{recap.topCategory.subcategoryTitle}</small>
              ) : null}
            </article>
          ) : null}
        </div>

        {recap.interestSummary ? (
          <article className="public-poster__interest">
            <span>ИНТЕРЕСЫ ГОДА</span>
            <p>{recap.interestSummary}</p>
          </article>
        ) : null}

        {recap.badges.length ? (
          <section className="public-poster__badges" aria-labelledby="public-badges-title">
            <h2 id="public-badges-title">Достижения года</h2>
            <div>
              {recap.badges.map((badge) => (
                <BadgeCard
                  key={badge.code}
                  title={badge.title}
                  description={badge.description}
                  level={badge.level}
                  iconUrl={
                    badge.iconUrl ?? artUrl(badge.code.startsWith('seller') ? 'badge-sales' : 'badge')
                  }
                />
              ))}
            </div>
          </section>
        ) : null}

        <div className="public-poster__actions">
          <a href="/">Создать свои итоги</a>
          <button type="button" onClick={() => void share()} disabled={shareState.status === 'loading'}>
            {shareState.status === 'loading' ? 'Подготавливаем ссылку…' : 'Поделиться карточкой'}
          </button>
          <p role="status">{shareState.status === 'success' ? shareState.message : '\u00a0'}</p>
        </div>
      </section>

    </main>
  );
}
