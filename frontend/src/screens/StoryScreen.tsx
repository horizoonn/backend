import { useEffect, useRef, useState } from 'react';
import type { PointerEvent as ReactPointerEvent } from 'react';

import { shareRecap, userMessage } from '../api/client';
import { SlideView } from '../components/Slides';
import type { Profile, Recap } from '../api/types';
import { isSupportedSlide } from '../lib/recapLogic';

interface StoryScreenProps {
  recap: Recap;
  profile: Profile;
  onExit: () => void;
}

export function StoryScreen({ recap, profile, onExit }: StoryScreenProps) {
  const storyContentRef = useRef<HTMLDivElement>(null);
  const gestureStartRef = useRef<{ pointerId: number; x: number; y: number } | null>(null);
  const [index, setIndex] = useState(0);
  const [archetypeEvidence, setArchetypeEvidence] = useState(false);
  const [categoryQuizAnswers, setCategoryQuizAnswers] = useState<Record<string, string>>({});
  const [shareState, setShareState] = useState<
    | { status: 'idle' }
    | { status: 'loading' }
    | { status: 'success'; message: string; url: string }
    | { status: 'failed'; message: string }
  >({ status: 'idle' });

  const slides = recap.slides.filter(isSupportedSlide);
  const slide = slides[index];
  const isLast = index === slides.length - 1;
  const isIntro = slide?.type === 'intro';
  const isActiveDays = slide?.type === 'active_days';
  const isViews = slide?.type === 'views';
  const isMessages = slide?.type === 'messages';
  const isFavorites = slide?.type === 'favorites';
  const isFavoriteCategory = slide?.type === 'favorite_category';
  const isPurchases = slide?.type === 'purchases';
  const isSales = slide?.type === 'sales';
  const isInterests = slide?.type === 'interests';
  const isArchetype = slide?.type === 'archetype';
  const isFinal = slide?.type === 'final';
  const categoryQuizKey = isFavoriteCategory ? slide.category.id : undefined;
  const categoryQuizEnabled = isFavoriteCategory && (slide.quizOptions?.length ?? 0) >= 2;
  const selectedCategoryQuizOptionId = categoryQuizKey
    ? categoryQuizAnswers[categoryQuizKey]
    : undefined;
  const isCategoryQuizQuestion = categoryQuizEnabled && !selectedCategoryQuizOptionId;
  const favoriteCategorySlide = slides.find(
    (item): item is Extract<(typeof slides)[number], { type: 'favorite_category' }> =>
      item.type === 'favorite_category',
  );
  const interestsSlide = slides.find(
    (item): item is Extract<(typeof slides)[number], { type: 'interests' }> =>
      item.type === 'interests',
  );
  const storyClassName = [
    'story',
    'story--light',
    isIntro ? 'story--intro' : '',
    isActiveDays ? 'story--active-days' : '',
    isViews ? 'story--views' : '',
    isMessages ? 'story--messages' : '',
    isFavorites ? 'story--favorites' : '',
    isFavoriteCategory ? 'story--category' : '',
    isPurchases ? 'story--purchases' : '',
    isSales ? 'story--sales' : '',
    isInterests ? 'story--interests' : '',
    isArchetype ? 'story--archetype' : '',
    isFinal ? 'story--final' : '',
  ]
    .filter(Boolean)
    .join(' ');

  const share = async () => {
    if (shareState.status === 'loading') {
      return;
    }

    const existingUrl = shareState.status === 'success' ? shareState.url : undefined;
    if (existingUrl) {
      await deliverSharedLink(existingUrl);
      return;
    }

    setShareState({ status: 'loading' });
    try {
      const link = await shareRecap(recap.id);
      await deliverSharedLink(link.url);
    } catch (cause: unknown) {
      console.error(cause);
      setShareState({ status: 'failed', message: userMessage(cause) });
    }
  };

  const deliverSharedLink = async (url: string) => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: `Итоги ${recap.year} — Авито`,
          text: `${profile.name} делится итогами года`,
          url,
        });
        setShareState({ status: 'success', message: 'Публичная ссылка отправлена', url });
        return;
      } catch (cause: unknown) {
        if (cause instanceof DOMException && cause.name === 'AbortError') {
          setShareState({ status: 'success', message: 'Публичная ссылка создана', url });
          return;
        }
        console.warn(cause);
      }
    }

    try {
      await navigator.clipboard.writeText(url);
      setShareState({ status: 'success', message: 'Публичная ссылка скопирована', url });
    } catch (cause: unknown) {
      console.warn(cause);
      window.prompt('Скопируйте публичную ссылку на итоги', url);
      setShareState({ status: 'success', message: 'Публичная ссылка создана', url });
    }
  };

  const showPrevious = () => {
    if (isArchetype && archetypeEvidence) {
      setArchetypeEvidence(false);
      return;
    }

    setArchetypeEvidence(false);
    setIndex((current) => Math.max(0, current - 1));
  };

  const showNext = () => {
    if (isCategoryQuizQuestion) {
      return;
    }

    if (isArchetype && !archetypeEvidence) {
      setArchetypeEvidence(true);
      return;
    }

    setArchetypeEvidence(false);
    if (isLast) {
      onExit();
      return;
    }

    setIndex((current) => current + 1);
  };

  const handlePointerDown = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (!event.isPrimary || event.button !== 0 || isInteractiveTarget(event.target)) {
      gestureStartRef.current = null;
      return;
    }

    gestureStartRef.current = {
      pointerId: event.pointerId,
      x: event.clientX,
      y: event.clientY,
    };
  };

  const handlePointerUp = (event: ReactPointerEvent<HTMLDivElement>) => {
    const start = gestureStartRef.current;
    gestureStartRef.current = null;

    if (!start || start.pointerId !== event.pointerId || isFinal) {
      return;
    }

    const deltaX = event.clientX - start.x;
    const deltaY = event.clientY - start.y;
    const horizontalDistance = Math.abs(deltaX);

    if (horizontalDistance >= 48 && horizontalDistance > Math.abs(deltaY) * 1.25) {
      if (deltaX < 0) {
        showNext();
      } else if (index > 0 || archetypeEvidence) {
        showPrevious();
      }
      return;
    }

    if (window.innerWidth > 767 || horizontalDistance > 10 || Math.abs(deltaY) > 10) {
      return;
    }

    const bounds = event.currentTarget.getBoundingClientRect();
    const position = (event.clientX - bounds.left) / bounds.width;

    if (position <= 0.35 && (index > 0 || archetypeEvidence)) {
      showPrevious();
    } else if (position >= 0.65) {
      showNext();
    }
  };

  useEffect(() => {
    const frameId = window.requestAnimationFrame(() => {
      storyContentRef.current?.scrollTo({ top: 0 });
      storyContentRef.current?.focus({ preventScroll: true });
    });

    return () => window.cancelAnimationFrame(frameId);
  }, [index, archetypeEvidence, selectedCategoryQuizOptionId]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (
        event.defaultPrevented ||
        event.altKey ||
        event.ctrlKey ||
        event.metaKey ||
        event.shiftKey
      ) {
        return;
      }

      if (event.key === 'Escape' && archetypeEvidence) {
        event.preventDefault();
        setArchetypeEvidence(false);
        return;
      }

      if (event.key === 'ArrowLeft' && (index > 0 || archetypeEvidence)) {
        event.preventDefault();
        if (isArchetype && archetypeEvidence) {
          setArchetypeEvidence(false);
        } else {
          setArchetypeEvidence(false);
          setIndex((current) => Math.max(0, current - 1));
        }
        return;
      }

      if (event.key === 'ArrowRight' && !isFinal) {
        event.preventDefault();
        if (isCategoryQuizQuestion) {
          return;
        }

        if (isArchetype && !archetypeEvidence) {
          setArchetypeEvidence(true);
        } else {
          setArchetypeEvidence(false);
          setIndex((current) => Math.min(slides.length - 1, current + 1));
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [archetypeEvidence, index, isArchetype, isCategoryQuizQuestion, isFinal, slides.length]);

  if (!slide) {
    return null;
  }

  return (
    <div className={storyClassName}>
      <header className="story__top">
        <button type="button" className="story__back" onClick={onExit}>
          ‹ К профилям
        </button>

        <div
          className="story__progress"
          role="progressbar"
          aria-valuenow={index + 1}
          aria-valuemin={1}
          aria-valuemax={slides.length}
        >
          {slides.map((item, position) => (
            <span
              key={`${item.type}-${position}`}
              className={`story__segment${position <= index ? ' story__segment--done' : ''}`}
            />
          ))}
        </div>

        <span className="story__counter">
          {index + 1} / {slides.length}
        </span>
      </header>

      <div
        className="story__content"
        ref={storyContentRef}
        role="region"
        tabIndex={-1}
        aria-label={`Слайд ${index + 1} из ${slides.length}`}
        onPointerDown={handlePointerDown}
        onPointerUp={handlePointerUp}
        onPointerCancel={() => {
          gestureStartRef.current = null;
        }}
      >
        <SlideView
          slide={slide}
          recapId={recap.id}
          profileName={profile.name}
          onShare={() => void share()}
          shareDisabled={shareState.status === 'loading'}
          shareLabel={shareState.status === 'loading' ? 'Создаём публичную ссылку…' : undefined}
          archetypeEvidence={archetypeEvidence}
          recapYear={recap.year}
          recapArchetype={recap.archetype}
          onExit={onExit}
          shareFeedback={
            shareState.status === 'success' || shareState.status === 'failed'
              ? { message: shareState.message, failed: shareState.status === 'failed' }
              : undefined
          }
          shareUrl={shareState.status === 'success' ? shareState.url : undefined}
          selectedCategoryQuizOptionId={selectedCategoryQuizOptionId}
          onSelectCategoryQuizOption={(categoryId) => {
            if (!categoryQuizKey) {
              return;
            }

            setCategoryQuizAnswers((current) => ({
              ...current,
              [categoryQuizKey]: categoryId,
            }));
          }}
          favoriteCategorySlide={favoriteCategorySlide}
          interestSummary={interestsSlide?.shiftSummary}
        />
      </div>

      {!isFinal ? (
        <footer className="story__bottom">
          <button
            type="button"
            className="story__circle"
            disabled={index === 0 && !archetypeEvidence}
            aria-label="Предыдущий слайд"
            onClick={showPrevious}
          >
            ←
          </button>

          <button
            type="button"
            className="button button--light"
            disabled={isCategoryQuizQuestion}
            onClick={showNext}
          >
            {isArchetype
              ? archetypeEvidence
                ? 'К итогам →'
                : 'Почему?'
              : isLast
                ? `К профилям, ${profile.name}`
                : isIntro
                  ? 'Начать →'
                  : 'Дальше →'}
          </button>
        </footer>
      ) : null}
    </div>
  );
}

function isInteractiveTarget(target: EventTarget | null): boolean {
  return (
    target instanceof Element &&
    target.closest('button, a, input, select, textarea, [role="button"]') !== null
  );
}
