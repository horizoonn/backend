import {
  ArrowRight,
  Check,
  ChevronDown,
  LockKeyhole,
  RefreshCw,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

import { listProfiles, userMessage } from '../api/client';
import type { Profile } from '../api/types';

interface StartScreenProps {
  year: number;
  onStart: (profile: Profile) => void;
}

export function StartScreen({ year, onStart }: StartScreenProps) {
  const pickerRef = useRef<HTMLDivElement>(null);
  const pickerToggleRef = useRef<HTMLButtonElement>(null);
  const [profiles, setProfiles] = useState<Profile[]>([]);
  const [selected, setSelected] = useState<Profile | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadAttempt, setLoadAttempt] = useState(0);

  useEffect(() => {
    let cancelled = false;

    setLoading(true);
    setError(null);

    listProfiles()
      .then((items) => {
        if (cancelled) {
          return;
        }

        setProfiles(items);
        setSelected(items.find((profile) => profile.name === 'Илья') ?? items[0] ?? null);
      })
      .catch((cause: unknown) => {
        if (!cancelled) {
          console.error(cause);
          setProfiles([]);
          setSelected(null);
          setError(userMessage(cause));
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [loadAttempt]);

  useEffect(() => {
    if (!pickerOpen) {
      return;
    }

    const handlePointerDown = (event: PointerEvent) => {
      if (event.target instanceof Node && !pickerRef.current?.contains(event.target)) {
        setPickerOpen(false);
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') {
        return;
      }

      event.preventDefault();
      setPickerOpen(false);
      pickerToggleRef.current?.focus();
    };

    document.addEventListener('pointerdown', handlePointerDown);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('pointerdown', handlePointerDown);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [pickerOpen]);

  return (
    <main className="start">
      <header className="start__top">
        <a className="start__wordmark" href="/" aria-label="Авито, на главную">
          Авито
        </a>

        <p className="start__privacy">
          <LockKeyhole aria-hidden="true" size={16} />
          <span>
            Итоги видите только вы. Публичная карточка появится только после вашего действия.
          </span>
        </p>

        <div className="picker" ref={pickerRef}>
          <button
            ref={pickerToggleRef}
            type="button"
            className="picker__toggle"
            aria-expanded={pickerOpen}
            aria-controls="demo-profile-list"
            aria-haspopup="dialog"
            onClick={() => setPickerOpen((open) => !open)}
            disabled={loading || profiles.length === 0}
          >
            <ProfileAvatar profile={selected} />
            <span className="picker__toggle-copy">
              <small>Демо-сценарий</small>
              <strong>{selected?.name ?? 'Выберите профиль'}</strong>
            </span>
            <ChevronDown className={pickerOpen ? 'picker__chevron--open' : ''} size={18} />
          </button>

          {pickerOpen ? (
            <div
              className="picker__menu"
              id="demo-profile-list"
              role="dialog"
              aria-label="Выбор демо-сценария"
            >
              <p className="picker__caption">Демо-сценарии</p>
              <div className="picker__list" role="radiogroup" aria-label="Тестовые профили">
                {profiles.map((profile) => (
                  <button
                    type="button"
                    role="radio"
                    aria-checked={selected?.id === profile.id}
                    className={`picker__item${selected?.id === profile.id ? ' picker__item--active' : ''}`}
                    key={profile.id}
                    onClick={() => {
                      setSelected(profile);
                      setPickerOpen(false);
                    }}
                  >
                    <ProfileAvatar profile={profile} />
                    <span className="picker__text">
                      <strong>
                        {profile.name} {profile.surname}
                      </strong>
                      <small>{profile.hint ?? 'Персональная история года'}</small>
                    </span>
                    {selected?.id === profile.id ? <Check aria-hidden="true" size={18} /> : null}
                  </button>
                ))}
              </div>
            </div>
          ) : null}
        </div>
      </header>

      <div className="start__hero">
        <div className="start__content">
          <p className="start__eyebrow">ИТОГИ {year}</p>

          <h1 className="start__title">Какой вы пользователь Авито?</h1>

          <p className="start__subtitle">
            Соберём ваши находки, сделки и интересы в историю {year} года.
          </p>

          {!loading && error ? (
            <StartInlineState
              title="Не удалось загрузить профили"
              description={error}
              onRetry={() => setLoadAttempt((attempt) => attempt + 1)}
            />
          ) : null}

          {!loading && !error && profiles.length === 0 ? (
            <StartInlineState
              title="Тестовые профили пока не добавлены"
              description="Заполните базу демо-данными и повторите."
              onRetry={() => setLoadAttempt((attempt) => attempt + 1)}
            />
          ) : null}

          <button
            type="button"
            className="button button--start"
            disabled={!selected || loading || Boolean(error)}
            onClick={() => selected && onStart(selected)}
          >
            Открыть мои итоги
            <ArrowRight aria-hidden="true" size={20} />
          </button>

          <p className="start__demo-note">
            {loading ? 'Загружаем демо-сценарии…' : 'Сценарий можно сменить в списке сверху'}
          </p>
        </div>

        <div className="start__art" aria-hidden="true">
          <span className="start__art-label">Ваш год на Авито</span>
          <span className="start__year start__year--blue">{String(year).slice(0, 1)}</span>
          <span className="start__year start__year--coral">{String(year).slice(1, 2)}</span>
          <span className="start__year start__year--green">{String(year).slice(2, 3)}</span>
          <span className="start__year start__year--violet">{String(year).slice(3, 4)}</span>
          <span className="start__art-sticker">находки</span>
          <span className="start__art-sticker start__art-sticker--second">интересы</span>
        </div>
      </div>
    </main>
  );
}

function ProfileAvatar({ profile }: { profile: Profile | null }) {
  return (
    <span className="picker__avatar" aria-hidden="true">
      {profile?.name.slice(0, 1) ?? '—'}
      {profile?.avatarUrl ? (
        <img
          src={profile.avatarUrl}
          alt=""
          onError={(event) => {
            event.currentTarget.hidden = true;
          }}
        />
      ) : null}
    </span>
  );
}

function StartInlineState({
  title,
  description,
  onRetry,
}: {
  title: string;
  description: string;
  onRetry: () => void;
}) {
  return (
    <div className="profile-selection__state" role="alert">
      <strong>{title}</strong>
      <p>{description}</p>
      <button type="button" onClick={onRetry}>
        <RefreshCw aria-hidden="true" size={16} />
        Повторить
      </button>
    </div>
  );
}
