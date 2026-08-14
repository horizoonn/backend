import { useEffect, useState } from 'react';

import type { Profile } from '../api/types';
import { BrandLoader } from '../components/BrandLoader';

const GENERATING_PHRASES = [
  'Собираем ваш год',
  'Ищем главную находку',
  'Определяем ваш стиль на Авито',
] as const;

interface GeneratingScreenProps {
  profile: Profile;
  year: number;
}

export function GeneratingScreen({ profile, year }: GeneratingScreenProps) {
  const [phraseIndex, setPhraseIndex] = useState(0);
  const phrase = GENERATING_PHRASES[phraseIndex] ?? GENERATING_PHRASES[0];

  useEffect(() => {
    const intervalId = window.setInterval(() => {
      setPhraseIndex((index) => (index + 1) % GENERATING_PHRASES.length);
    }, 1200);

    return () => window.clearInterval(intervalId);
  }, []);

  return (
    <main className="generating" aria-labelledby="generating-title">
      <div className="generating__content">
        <BrandLoader />

        <div className="generating__copy">
          <h1 className="generating__title" id="generating-title">
            Собираем ваш {year} год
          </h1>

          <div className="generating__phrase" aria-hidden="true">
            <p key={phraseIndex}>{phrase}</p>
          </div>

          <p className="visually-hidden" role="status">
            Собираем ваш год. Итоги откроются автоматически.
          </p>
        </div>

        <p className="generating__profile">
          Для профиля <strong>{profile.name}</strong>
        </p>
      </div>
    </main>
  );
}
