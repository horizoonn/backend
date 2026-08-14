import type { Profile } from '../api/types';
import { avitoHomeRedirectUrl } from '../api/client';

interface LowActivityScreenProps {
  profile: Profile;
  onExit: () => void;
}

export function LowActivityScreen({ profile, onExit }: LowActivityScreenProps) {
  return (
    <main className="recap-state recap-state--quiet" aria-labelledby="low-activity-title">
      <div className="recap-state__content">
        <div className="recap-state__pin" aria-hidden="true" />
        <p className="recap-state__label">ИТОГИ ГОДА</p>
        <h1 id="low-activity-title">В этом году было мало активности</h1>
        <p className="recap-state__lead">
          Чтобы собрать осмысленные итоги, нужно чуть больше действий.
        </p>
        <p className="recap-state__context">
          {profile.name}, зато следующий год уже можно наполнить находками, сообщениями и сделками.
        </p>

        <div className="recap-state__actions">
          <a className="button recap-state__concept-action" href={avitoHomeRedirectUrl}>
            Перейти на Авито
          </a>
          <button type="button" className="recap-state__secondary" onClick={onExit}>
            Выбрать другой профиль
          </button>
        </div>
      </div>
    </main>
  );
}

interface RecapErrorScreenProps {
  message: string;
  onRetry: () => void;
  onExit: () => void;
}

export function RecapErrorScreen({ message, onRetry, onExit }: RecapErrorScreenProps) {
  return (
    <main className="recap-state recap-state--error" aria-labelledby="recap-error-title">
      <div className="recap-state__content" role="alert">
        <div className="recap-state__pin" aria-hidden="true" />
        <p className="recap-state__label">НЕ ПОЛУЧИЛОСЬ</p>
        <h1 id="recap-error-title">Итоги не собрались</h1>
        <p className="recap-state__lead">{message}</p>

        <div className="recap-state__actions">
          <button type="button" className="button recap-state__primary" onClick={onRetry}>
            Повторить
          </button>
          <button type="button" className="recap-state__secondary" onClick={onExit}>
            Выбрать другой профиль
          </button>
        </div>
      </div>
    </main>
  );
}
