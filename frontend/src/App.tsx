import { useCallback, useEffect, useState } from 'react';

import { ApiError, createRecap, getRecap, getSharedRecap, userMessage } from './api/client';
import type { ErrorCode, Profile, Recap, SharedRecap } from './api/types';
import { GeneratingScreen } from './screens/GeneratingScreen';
import { LowActivityScreen, RecapErrorScreen } from './screens/RecapFailureScreen';
import { PublicLoadingScreen } from './screens/PublicLoadingScreen';
import { PublicRecapScreen } from './screens/PublicRecapScreen';
import { StartScreen } from './screens/StartScreen';
import { StoryScreen } from './screens/StoryScreen';

const RECAP_YEAR = Number(import.meta.env.VITE_RECAP_YEAR ?? new Date().getFullYear() - 1);
const SHARED_RECAP_PATH = /^\/shared-recaps\/([A-Za-z0-9_-]{22})\/?$/;

type Screen =
  | { name: 'start' }
  | { name: 'generating'; profile: Profile }
  | { name: 'story'; profile: Profile; recap: Recap }
  | { name: 'failed'; profile: Profile; code?: ErrorCode; message: string }
  | { name: 'public-loading'; token: string }
  | { name: 'public-ready'; recap: SharedRecap }
  | { name: 'public-failed'; message: string };

function initialScreen(): Screen {
  const { pathname } = window.location;
  if (!pathname.startsWith('/shared-recaps/')) {
    return { name: 'start' };
  }

  const match = SHARED_RECAP_PATH.exec(pathname);
  if (!match?.[1]) {
    return { name: 'public-failed', message: 'Публичная ссылка имеет неверный формат.' };
  }

  return { name: 'public-loading', token: match[1] };
}

export function App() {
  const [screen, setScreen] = useState<Screen>(initialScreen);

  const start = useCallback((profile: Profile) => {
    setScreen({ name: 'generating', profile });
  }, []);

  const backToStart = useCallback(() => {
    setScreen({ name: 'start' });
  }, []);

  const retryRecap = useCallback((profile: Profile) => {
    setScreen({ name: 'generating', profile });
  }, []);

  useEffect(() => {
    if (screen.name !== 'public-loading') {
      return;
    }

    let cancelled = false;
    const load = async () => {
      try {
        const recap = await getSharedRecap(screen.token);
        if (!cancelled) {
          setScreen({ name: 'public-ready', recap });
        }
      } catch (cause: unknown) {
        console.error(cause);
        if (!cancelled) {
          setScreen({ name: 'public-failed', message: userMessage(cause) });
        }
      }
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [screen]);

  useEffect(() => {
    if (screen.name !== 'generating') {
      return;
    }

    const { profile } = screen;
    let cancelled = false;
    const generate = async () => {
      try {
        const recapId = await createRecap(profile.id, RECAP_YEAR);
        const recap = await getRecap(recapId);
        if (!cancelled) {
          setScreen({ name: 'story', profile, recap });
        }
      } catch (cause: unknown) {
        console.error(cause);
        if (!cancelled) {
          setScreen({
            name: 'failed',
            profile,
            code: cause instanceof ApiError ? cause.code : undefined,
            message: userMessage(cause),
          });
        }
      }
    };

    void generate();
    return () => {
      cancelled = true;
    };
  }, [screen]);

  switch (screen.name) {
    case 'start':
      return <StartScreen year={RECAP_YEAR} onStart={start} />;

    case 'generating':
      return <GeneratingScreen profile={screen.profile} year={RECAP_YEAR} />;

    case 'failed':
      return screen.code === 'not_enough_activity' ? (
        <LowActivityScreen profile={screen.profile} onExit={backToStart} />
      ) : (
        <RecapErrorScreen
          message={screen.message}
          onRetry={() => retryRecap(screen.profile)}
          onExit={backToStart}
        />
      );

    case 'story':
      return <StoryScreen recap={screen.recap} profile={screen.profile} onExit={backToStart} />;

    case 'public-loading':
      return <PublicLoadingScreen />;

    case 'public-ready':
      return <PublicRecapScreen recap={screen.recap} />;

    case 'public-failed':
      return (
        <div className="state">
          <p className="state__title">Публичные итоги не открылись</p>
          <p className="state__note">{screen.message}</p>
          <a className="button button--light" href="/">
            Создать свои итоги
          </a>
        </div>
      );
  }
}
